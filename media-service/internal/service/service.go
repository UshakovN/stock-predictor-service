package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"main/internal/domain"
	"main/internal/queue"
	"main/internal/storage"
	"os"
	"strings"

	"github.com/UshakovN/stock-predictor-service/utils"
)

const DirStoredMedia = "./stored_media"

type MediaService interface {
	GetMedia(input *domain.GetMediaInput) (*domain.Media, error)
	GetMediaBatch(inputs []*domain.GetMediaInput) ([]*domain.Media, error)
	PutMedia(input *domain.PutMediaInput) error
	HandleQueueMessages() error
}

type mediaService struct {
	ctx        context.Context
	msQueue    queue.MediaServiceQueue
	storage    storage.Storage
	hostPrefix string
}

func NewMediaService(ctx context.Context, config *Config) (MediaService, error) {
	service := &mediaService{
		ctx:        ctx,
		msQueue:    config.MsQueue,
		storage:    config.Storage,
		hostPrefix: config.HostPrefix,
	}
	err := checkDirForStoredMedia(DirStoredMedia)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (m *mediaService) GetMedia(input *domain.GetMediaInput) (*domain.Media, error) {
	mediaFileId := formMediaFileId(input.Name, input.Section)

	storedMedia, found, err := m.storage.GetStoredMedia(mediaFileId)
	if err != nil {
		return nil, fmt.Errorf("cannot get stored media file: %v", err)
	}
	media := &domain.Media{
		Name:    input.Name,
		Section: input.Section,
	}
	if found {
		media.Found = true
		media.SourceUrl = storedMedia.FormedURL
	}
	return media, nil
}

func (m *mediaService) GetMediaBatch(inputs []*domain.GetMediaInput) ([]*domain.Media, error) {
	mediaFileIds := make([]string, 0, len(inputs))

	for _, input := range inputs {
		fileId := formMediaFileId(input.Name, input.Section)
		mediaFileIds = append(mediaFileIds, fileId)
	}
	storedMedia, err := m.storage.GetStoredMediaBatch(mediaFileIds)
	if err != nil {
		return nil, fmt.Errorf("cannot get stored media batch: %v", err)
	}

	storedMediaMap := utils.ToMap(storedMedia, func(media *storage.StoredMedia) string {
		return media.StoredMediaId
	})
	mediaBatch := make([]*domain.Media, 0, len(inputs))

	for inputIdx, input := range inputs {
		fileId := mediaFileIds[inputIdx]

		media := &domain.Media{
			Name:    input.Name,
			Section: input.Section,
		}
		if storedMedia, found := storedMediaMap[fileId]; found {
			media.Found = true
			media.SourceUrl = storedMedia.FormedURL
		}
		mediaBatch = append(mediaBatch, media)
	}

	return mediaBatch, nil
}

func (m *mediaService) PutMedia(input *domain.PutMediaInput) error {
	if err := m.msQueue.PublishMessage(input.ToMessage()); err != nil {
		return fmt.Errorf("cannot publish message to media service queue: %v", err)
	}
	return nil
}

func (m *mediaService) HandleQueueMessages() error {
	return m.msQueue.ConsumeMessages(func(message *domain.PutMessage) error {
		createFileResult, err := createNewMediaFileOrIgnore(
			message.MetaInfo.Name,
			message.MetaInfo.Section,
			message.MetaInfo.Overwrite,
			message.Content,
		)
		if err != nil {
			return fmt.Errorf("cannot create new media file: %v", err)
		}
		if !createFileResult.createdOrOverwritten {
			return nil
		}
		formedFileUrl := formMediaFileUrl(m.hostPrefix, createFileResult.filePath)

		err = m.storage.PutStoredMedia(&storage.StoredMedia{
			StoredMediaId: createFileResult.formedFileId,
			FormedURL:     formedFileUrl,
			CreatedBy:     message.MetaInfo.From,
			CreatedAt:     utils.NotTimeUTC(),
		})
		if err != nil {
			return fmt.Errorf("cannot put stored media to storage: %v", err)
		}

		return nil
	})
}

type createFileResult struct {
	createdOrOverwritten bool
	formedFileId         string
	filePath             string
}

func createNewMediaFileOrIgnore(
	fileName string,
	sectionName string,
	overwrite bool,
	content []byte,
) (*createFileResult, error) {

	const (
		filePathTemplate = "%s/%s.%s" // stored-media-path/file-id.extension
		fileCreateMode   = 0644       // user read-write | group read | other read
	)

	// get other file name for base security
	fileId := formMediaFileId(fileName, sectionName)

	fileExt, err := utils.ExtractFileExtension(fileName)
	if err != nil {
		return nil, fmt.Errorf("cannot extract media file extension: %v", err)
	}

	// form path for file
	filePath := fmt.Sprintf(filePathTemplate, DirStoredMedia, fileId, fileExt)

	if fileStat, err := os.Stat(filePath); err == nil && !fileStat.IsDir() && !overwrite {
		// if file found, it is not directory and overwrite flag not set
		return &createFileResult{
			createdOrOverwritten: false,
		}, nil
	}
	if err := os.WriteFile(filePath, content, fileCreateMode); err != nil {
		return nil, fmt.Errorf("cannot write content to file '%s': %v", filePath, err)
	}

	return &createFileResult{
		createdOrOverwritten: true,
		formedFileId:         fileId,
		filePath:             filePath,
	}, nil
}

func checkDirForStoredMedia(dirPath string) error {
	baseDirStat, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("base directory '%s' for stored media does not exist: %v",
				baseDirStat, err)
		}
		return fmt.Errorf("cannot get stat about base directory '%s' for stored media: %v",
			baseDirStat, err)
	}
	if !baseDirStat.IsDir() {
		return fmt.Errorf("check stored media path: '%s' is not a directory", baseDirStat)
	}
	return nil
}

func formMediaFileUrl(hostPrefix, filePath string) string {
	const (
		storedMediaDirPrefix = "./"
		fileUrlTemplate      = "%s/%s" // host-prefix/file-path-with-extension
	)
	filePath = strings.TrimPrefix(filePath, storedMediaDirPrefix)

	return fmt.Sprintf(fileUrlTemplate, hostPrefix, filePath)
}

func formMediaFileId(fileName, sectionName string) string {
	sb := strings.Builder{}
	sb.WriteString(fileName)
	sb.WriteString(sectionName)

	fileInfo := []byte(sb.String())
	hashedFileInfo := fmt.Sprintf("%x", sha256.Sum256(fileInfo))

	return hashedFileInfo
}