package service

import (
  "main/internal/storage"
  "main/pkg/auth"
  "main/pkg/hash"
  "time"
)

type Config struct {
  TokenTtl        time.Duration
  TokenManager    auth.TokenManager
  PasswordManager hash.PasswordManager
  Storage         storage.Storage
}
