package service

import (
  "main/internal/queue"
  "main/internal/storage"

  "github.com/UshakovN/stock-predictor-service/hash"
)

type Config struct {
  MsQueue     queue.MediaServiceQueue
  Storage     storage.Storage
  HashManager hash.Manager
}
