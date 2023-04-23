package service

import (
  "main/internal/storage"

  "github.com/UshakovN/stock-predictor-service/httpclient"
)

type Config struct {
  Storage   storage.Storage
  ApiClient httpclient.HttpClient
}
