package domain

import "time"

type ServiceUser struct {
  UserId       string    `json:"user_id"`
  Email        string    `json:"email"`
  PasswordHash string    `json:"password_hash"`
  FullName     string    `json:"full_name"`
  Active       bool      `json:"active"`
  CreatedAt    time.Time `json:"created_at"`
}
