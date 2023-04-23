package utils

import (
  "fmt"

  "github.com/google/uuid"
)

func NewUUID() (string, error) {
  gen, err := uuid.NewUUID()
  if err != nil {
    return "", fmt.Errorf("cannot create uuid: %v", err)
  }
  return gen.String(), nil
}
