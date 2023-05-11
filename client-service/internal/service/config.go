package service

import (
  "main/internal/storage"

  mediaservice "github.com/UshakovN/stock-predictor-service/contract/media-service"
)

type Config struct {
  Storage     storage.Storage
  MediaClient mediaservice.Client
}
