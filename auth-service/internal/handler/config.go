package handler

import (
  "github.com/UshakovN/stock-predictor-service/postgres"
  "github.com/UshakovN/stock-predictor-service/swagger"
)

type Config struct {
  PasswordSalt    string           `yaml:"password_salt" required:"true"`
  TokenSignKey    string           `yaml:"token_sign_key" required:"true"`
  TokenTtlMinutes int64            `yaml:"token_ttl_minutes" required:"true"`
  StorageConfig   *postgres.Config `yaml:"storage_config" required:"true"`
  SwaggerConfig   *swagger.Config  `yaml:"swagger_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
