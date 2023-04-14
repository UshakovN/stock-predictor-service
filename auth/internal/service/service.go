package service

import (
	"context"
	"fmt"
	"main/internal/domain"
	"main/internal/storage"
	"time"

	"github.com/UshakovN/stock-predictor-service/auth"
	"github.com/UshakovN/stock-predictor-service/errs"
	"github.com/UshakovN/stock-predictor-service/hash"
	"github.com/UshakovN/stock-predictor-service/utils"
	"github.com/google/uuid"
)

type UserAuthService interface {
	SignUp(input *domain.SignUpInput) (*domain.Tokens, error)
	SignIn(input *domain.SignInInput) (*domain.Tokens, error)
	RefreshTokens(refreshToken string) (*domain.Tokens, error)
	CheckUser(userId string) (*domain.UserInfo, error)
}

type userAuthService struct {
	ctx             context.Context
	tokenTtl        time.Duration
	tokenManager    auth.TokenManager
	passwordManager hash.PasswordManager
	storage         storage.Storage
}

func NewService(ctx context.Context, config *Config) UserAuthService {
	return &userAuthService{
		ctx:             ctx,
		tokenTtl:        config.TokenTtl,
		tokenManager:    config.TokenManager,
		passwordManager: config.PasswordManager,
		storage:         config.Storage,
	}
}

func (s *userAuthService) SignUp(input *domain.SignUpInput) (*domain.Tokens, error) {
	passwordHash := s.passwordManager.Hash(input.Password)

	userId, err := s.newUserId()
	if err != nil {
		return nil, fmt.Errorf("cannot form user id: %v", err)
	}

	if err = s.storage.PutUser(&storage.ServiceUser{
		UserId:       userId,
		Email:        input.Email,
		PasswordHash: passwordHash,
		FullName:     input.FullName,
		Active:       true,
		CreatedAt:    utils.NotTimeUTC(),
	}); err != nil {
		return nil, fmt.Errorf("cannot put user to storage: %v", err)
	}

	userTokens, err := s.newUserTokens(userId)
	if err != nil {
		return nil, fmt.Errorf("cannot create user tokens: %v", err)
	}
	if err := s.storage.PutToken(&storage.RefreshToken{
		TokenId: userTokens.Refresh,
		Active:  true,
	}); err != nil {
		return nil, fmt.Errorf("cannot put refresh token to storage: %v", err)
	}

	return userTokens, nil
}

func (s *userAuthService) SignIn(input *domain.SignInInput) (*domain.Tokens, error) {
	serviceUser, found, err := s.storage.GetUser("", input.Email)
	if err != nil {
		return nil, fmt.Errorf("cannot get user from storage: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("user with email '%s' not found", input.Email)
	}

	if serviceUser.PasswordHash != s.passwordManager.Hash(input.Password) {
		return nil, fmt.Errorf("specified wrong password")
	}
	userTokens, err := s.newUserSession(serviceUser.UserId)
	if err != nil {
		return nil, fmt.Errorf("cannot create user session: %v", err)
	}

	return userTokens, nil
}

func (s *userAuthService) CheckUser(userId string) (*domain.UserInfo, error) {
	serviceUser, found, err := s.storage.GetUser(userId, "")
	if err != nil {
		return nil, fmt.Errorf("cannot get user from storage: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("user with id '%s' not found", userId)
	}
	return &domain.UserInfo{
		Email:     serviceUser.Email,
		FullName:  serviceUser.FullName,
		CreatedAt: serviceUser.CreatedAt,
	}, nil
}

func (s *userAuthService) newUserTokens(userId string) (*domain.Tokens, error) {
	accessToken, err := s.tokenManager.NewJWT(userId, s.tokenTtl)
	if err != nil {
		return nil, fmt.Errorf("cannot create jwt token: %v", err)
	}

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("cannot refresh token: %v", err)
	}

	return &domain.Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *userAuthService) newUserId() (string, error) {
	getId, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("cannot create uuid for user id: %v", err)
	}
	userId := getId.String()

	return userId, nil
}

func (s *userAuthService) newUserSession(userId string) (*domain.Tokens, error) {
	userTokens, err := s.newUserTokens(userId)
	if err != nil {
		return nil, fmt.Errorf("cannot create user tokens: %v", err)
	}

	if err := s.storage.PutToken(&storage.RefreshToken{
		TokenId: userTokens.Refresh,
		Active:  true,
	}); err != nil {
		return nil, fmt.Errorf("cannot put refresh token to storage: %v", err)
	}

	return userTokens, err
}

func (s *userAuthService) RefreshTokens(refreshToken string) (*domain.Tokens, error) {
	storedToken, found, err := s.storage.GetToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("cannot get token from storage: %v", err)
	}
	if !found {
		return nil, errs.NewError(errs.ErrTypeMalformedToken, nil)
	}
	if err = confirmRefreshToken(storedToken); err != nil {
		return nil, err
	}
	storedToken.Active = false

	if err = s.storage.UpdateToken(storedToken); err != nil {
		return nil, fmt.Errorf("cannot update refresh token: %v", err)
	}
	userTokens, err := s.newUserSession(storedToken.UserId)

	return userTokens, nil
}

func confirmRefreshToken(token *storage.RefreshToken) error {
	if !token.Active {
		return errs.NewError(errs.ErrTypeExpiredToken, nil)
	}
	if token.UserId == "" {
		return errs.NewError(errs.ErrTypeForbidden, nil)
	}
	return nil
}
