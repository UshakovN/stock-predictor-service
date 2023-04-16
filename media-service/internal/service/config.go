package service

import (
	"main/internal/queue"
	"main/internal/storage"
)

type Config struct {
	MsQueue    queue.MediaServiceQueue
	Storage    storage.Storage
	HostPrefix string
}
