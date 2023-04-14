package service

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/storage"
  "main/pkg/auth"
  "main/pkg/hash"
  "main/pkg/utils"
  "time"

  "github.com/google/uuid"
)

type UserAuthService interface {
  SignUp(input *domain.SignUpInput) (*domain.Tokens, error)
  SignIn(input *domain.SignInInput) (*domain.Tokens, error)
  CheckUser(email string) (*domain.UserInfo, error)
  DeactivateUser(email string) error // TODO: complete this
  DeleteUser(email string) error     // TODO: complete this
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
  if err = s.storage.PutUser(&domain.ServiceUser{
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

  return userTokens, nil
}

func (s *userAuthService) SignIn(input *domain.SignInInput) (*domain.Tokens, error) {
  serviceUser, found, err := s.storage.GetUser(input.Email)
  if err != nil {
    return nil, fmt.Errorf("cannot get user from storage: %v", err)
  }
  if !found {
    return nil, fmt.Errorf("user with email '%s' not found", input.Email)
  }
  if serviceUser.PasswordHash != s.passwordManager.Hash(input.Password) {
    return nil, fmt.Errorf("specified wrong password")
  }
  userTokens, err := s.newUserTokens(serviceUser.UserId)
  if err != nil {
    return nil, fmt.Errorf("cannot create user tokens: %v", err)
  }

  return userTokens, nil
}

func (s *userAuthService) CheckUser(email string) (*domain.UserInfo, error) {
  serviceUser, found, err := s.storage.GetUser(email)
  if err != nil {
    return nil, fmt.Errorf("cannot get user from storage: %v", err)
  }
  if !found {
    return nil, fmt.Errorf("user with email '%s' not found", email)
  }

  return &domain.UserInfo{
    Email:     email,
    FullName:  serviceUser.FullName,
    CreatedAt: serviceUser.CreatedAt,
  }, nil
}

func (s *userAuthService) DeactivateUser(email string) error {
  return nil // TODO: complete this
}

func (s *userAuthService) DeleteUser(email string) error {
  return nil // TODO: complete this
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
  gen, err := uuid.NewUUID()
  if err != nil {
    return "", fmt.Errorf("cannot create uuid for user id: %v", err)
  }
  userId := gen.String()

  return userId, nil
}
