package utils

import (
  "fmt"
  "os"

  "gopkg.in/yaml.v3"
)

type Config interface {
  Parse(configPath string) error
}

func ParseYamlConfig[T Config](path string, config T) error {
  file, err := os.Open(path)
  if err != nil {
    return fmt.Errorf("cannot open config file: %v", err)
  }
  if err := yaml.NewDecoder(file).Decode(config); err != nil {
    return fmt.Errorf("cannot decode config: %v", err)
  }
  return nil
}
