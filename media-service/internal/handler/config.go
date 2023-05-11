package handler

import (
  "github.com/UshakovN/stock-predictor-service/postgres"
  "github.com/UshakovN/stock-predictor-service/rabbitmq"
)

type Config struct {
  AuthServicePrefix    string           `yaml:"auth_service_prefix" required:"true"`
  MediaServiceHashSalt string           `yaml:"media_service_hash_salt" required:"true"`
  MediaServiceApiToken string           `yaml:"media_service_api_token" required:"true"`
  QueueConfig          *rabbitmq.Config `yaml:"queue_config" required:"true"`
  StorageConfig        *postgres.Config `yaml:"storage_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
