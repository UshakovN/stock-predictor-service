package storage

import "time"

type StoredMedia struct {
  StoredMediaId string    `json:"stored_media_id"`
  Extension     string    `json:"extension"`
  CreatedBy     string    `json:"created_by"`
  CreatedAt     time.Time `json:"created_at"`
}
