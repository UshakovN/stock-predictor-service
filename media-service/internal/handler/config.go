package handler

import (
	"main/internal/queue/rabbitmq"
	"main/internal/storage/postgres"
)

type Config struct {
	QueueConfig   *rabbitmq.Config `yaml:"queue_config" required:"true"`
	StorageConfig *postgres.Config `yaml:"storage_config" required:"true"`
}

func NewConfig() *Config {
	return &Config{}
}
