package polygon

import (
  "fmt"
  "main/internal/queue/rabbitmq"
  "main/internal/storage/postgres"
  "main/pkg/utils"
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

func (c *Config) Parse(configPath string) error {
  if c == nil {
    return fmt.Errorf("fetcher config is a nil")
  }
  if err := utils.ParseYamlConfig(configPath, c); err != nil {
    return fmt.Errorf("cannot parse yaml config: %v", err)
  }
  if c.StorageConfig == nil {
    return fmt.Errorf("storage config is a nil")
  }
  if c.QueueConfig == nil {
    return fmt.Errorf("queue config is a nil")
  }
  return utils.CheckRequiredFields(c)
}
