package hash

import (
  "crypto/sha256"
  "fmt"
)

type PasswordManager interface {
  Hash(password string) string
}

type manager struct {
  hashSalt string
}

func NewManager(salt string) PasswordManager {
  return &manager{
    hashSalt: salt,
  }
}

func (m *manager) Hash(password string) string {
  hash := sha256.New()
  hash.Write([]byte(password))
  return fmt.Sprintf("%x", hash.Sum([]byte(m.hashSalt)))
}
