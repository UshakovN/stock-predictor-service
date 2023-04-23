package handler

import (
  "github.com/UshakovN/stock-predictor-service/postgres"
  "github.com/UshakovN/stock-predictor-service/rabbitmq"
)

type Config struct {
  QueueConfig   *rabbitmq.Config `yaml:"queue_config" required:"true"`
  StorageConfig *postgres.Config `yaml:"storage_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
