package storage

import "time"

type StoredMedia struct {
	StoredMediaId string    `json:"stored_media_id"`
	FormedURL     string    `json:"formed_url"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}
