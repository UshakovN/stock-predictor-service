package handler

import "github.com/UshakovN/stock-predictor-service/postgres"

type Config struct {
  MediaServicePrefix string           `yaml:"media_service_prefix" required:"true"`
  StorageConfig      *postgres.Config `yaml:"storage_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
