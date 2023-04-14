package handler

import (
	"fmt"
	"main/internal/queue/rabbitmq"
	"main/internal/storage/postgres"

	"github.com/UshakovN/stock-predictor-service/utils"
)

type Config struct {
	QueueConfig   *rabbitmq.Config `yaml:"queue_config" required:"true"`
	StorageConfig *postgres.Config `yaml:"storage_config" required:"true"`
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
	if c.QueueConfig == nil {
		return fmt.Errorf("queue config is a nil")
	}

	return utils.CheckRequiredFields(c)
}
