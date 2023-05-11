package hash

import (
  "crypto/sha256"
  "fmt"
)

type Manager interface {
  Hash(message string) string
}

type manager struct {
  hashSalt string
}

func NewManager(salt string) Manager {
  return &manager{
    hashSalt: salt,
  }
}

func (m *manager) Hash(message string) string {
  hash := sha256.New()
  hash.Write([]byte(message))
  hash.Write([]byte(m.hashSalt))
  return fmt.Sprintf("%x", hash.Sum(nil))
}
