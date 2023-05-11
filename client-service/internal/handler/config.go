package handler

import "github.com/UshakovN/stock-predictor-service/postgres"

type Config struct {
  ClientServiceApiToken string           `yaml:"client_service_api_token" required:"true"`
  MediaServiceApiToken  string           `yaml:"media_service_api_token" required:"true"`
  AuthServicePrefix     string           `yaml:"auth_service_prefix" required:"true"`
  MediaServicePrefix    string           `yaml:"media_service_prefix" required:"true"`
  StorageConfig         *postgres.Config `yaml:"storage_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
