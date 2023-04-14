package storage

import "time"

type ServiceUser struct {
	UserId       string    `json:"user_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	FullName     string    `json:"full_name"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type RefreshToken struct {
	TokenId string `json:"token_id"`
	Active  bool   `json:"active"`
	UserId  string `json:"user_id"`
}
