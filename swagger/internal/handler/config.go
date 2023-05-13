package handler

import "github.com/UshakovN/stock-predictor-service/swagger"

type Config struct {
  SwaggerServices []*SwaggerService `yaml:"swagger_services" required:"true"`
  SwaggerConfig   *swagger.Config   `yaml:"swagger_config" required:"true"`
}

type SwaggerService struct {
  Name   string `yaml:"name" required:"true"`
  Prefix string `yaml:"prefix" required:"true"`
}

func NewConfig() *Config {
  return &Config{}
}
