package fetcher

import (
  "main/internal/queue/rabbitmq"

  "github.com/UshakovN/stock-predictor-service/postgres"
)

type Config struct {
  ModeTotalHours   int              `yaml:"total_mode_hours" required:"true"`
  ModeCurrentHours int              `yaml:"current_mode_hours" required:"true"`
  ApiToken         string           `yaml:"api_token" required:"true"`
  StorageConfig    *postgres.Config `yaml:"storage_config" required:"true"`
  QueueConfig      *rabbitmq.Config `yaml:"queue_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
