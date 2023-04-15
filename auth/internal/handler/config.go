package handler

import (
	"fmt"

	"github.com/UshakovN/stock-predictor-service/postgres"
	"github.com/UshakovN/stock-predictor-service/utils"
)

type Config struct {
	PasswordSalt    string           `yaml:"password_salt" required:"true"`
	TokenSignKey    string           `yaml:"token_sign_key" required:"true"`
	TokenTtlMinutes int64            `yaml:"token_ttl_minutes" required:"true"`
	StorageConfig   *postgres.Config `yaml:"storage_config" required:"true"`
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
	return utils.CheckRequiredFields(c)
}
