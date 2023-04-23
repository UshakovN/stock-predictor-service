package config

import (
  "fmt"
  "os"

  "github.com/UshakovN/stock-predictor-service/utils"
  "gopkg.in/yaml.v3"
)

func Parse(path string, config any) error {
  file, err := os.Open(path)
  if err != nil {
    return fmt.Errorf("cannot open config file: %v", err)
  }
  if err := yaml.NewDecoder(file).Decode(config); err != nil {
    return fmt.Errorf("cannot decode config: %v", err)
  }
  if err := utils.CheckRequiredFields(config); err != nil {
    return fmt.Errorf("fields check failed: %v", err)
  }
  return nil
}
