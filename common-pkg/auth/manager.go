package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const claimSub = "sub"

type TokenManager interface {
	NewJWT(id string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

type manager struct {
	signKey string
}

func NewManager(signKey string) (TokenManager, error) {
	if signKey == "" {
		return nil, fmt.Errorf("sign key is mandatory")
	}
	return &manager{
		signKey: signKey,
	}, nil
}

func (m *manager) NewJWT(sub string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Subject:   sub,
	})
	return token.SignedString([]byte(m.signKey))
}

func (m *manager) Parse(accessToken string) (string, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected sign in method")
		}
		return []byte(m.signKey), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("malformed jwt claims")
	}
	return claims[claimSub].(string), nil
}

func (m *manager) NewRefreshToken() (string, error) {
	gen, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	token := gen.String()
	if token == "" {
		return "", fmt.Errorf("failed to create new refresh token")
	}
	return token, nil
}
