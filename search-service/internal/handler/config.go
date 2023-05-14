package handler

import "github.com/UshakovN/stock-predictor-service/swagger"

type Config struct {
  SuggestApiToken       string          `yaml:"suggest_api_token" required:"true"`
  ClientServiceApiToken string          `yaml:"client_service_api_token" required:"true"`
  ClientServicePrefix   string          `yaml:"client_service_prefix" required:"true"`
  ElasticSearchPrefix   string          `yaml:"elastic_search_prefix" required:"true"`
  AuthServicePrefix     string          `yaml:"auth_service_prefix" required:"true"`
  SuggestUpdateHours    int             `yaml:"suggest_update_hours" required:"true"`
  ElasticIndexName      string          `yaml:"elastic_index_name" required:"true"`
  SwaggerConfig         *swagger.Config `yaml:"swagger_config" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
