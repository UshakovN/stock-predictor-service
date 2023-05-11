package service

import (
  "main/internal/storage"
  "time"

  "github.com/UshakovN/stock-predictor-service/auth"
  "github.com/UshakovN/stock-predictor-service/hash"
)

type Config struct {
  TokenTtl        time.Duration
  TokenManager    auth.TokenManager
  PasswordManager hash.Manager
  Storage         storage.Storage
}
