package manager

import (
	"context"
	"crypto/sha256"
	"fmt"
	"main/internal/domain"
	"main/internal/queue"
	"main/internal/storage"
	"mime"
	"os"
	"strings"

	"github.com/UshakovN/stock-predictor-service/errs"
	"github.com/UshakovN/stock-predictor-service/utils"
)

const DirStoredMedia = "./stored_media"

type MediaService interface {
	GetMedia(input *domain.GetMediaInput) (*domain.Media, error)
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
	mediaFileId, err := formMediaFileId(input.Name, input.Section, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("cannot form media file name: %v", err)
	}
	storedMedia, found, err := m.storage.GetStoredMedia(mediaFileId)
	if err != nil {
		return nil, fmt.Errorf("cannot get stored media file: %v", err)
	}
	if !found {
		return nil, errs.NewError(errs.ErrTypeNotFoundContent, &errs.LogMessage{
			Err: fmt.Errorf("media with name '%s', section '%s', id '%s' not found",
				input.Name, input.Section, mediaFileId),
		})
	}
	media := &domain.Media{
		SourceUrl: storedMedia.FormedURL,
	}
	return media, nil
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
			message.MetaInfo.ContentType,
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

		err = m.storage.PutStoredMedia(&domain.StoredMedia{
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

func createNewMediaFileOrIgnore(fileName, sectionName, contentType string, overwrite bool, content []byte) (*createFileResult, error) {
	const (
		filePathTemplate = "%s/%s.%s" // stored-media-path/file-id.extension
		fileCreateMode   = 0644       // user read-write | group read | other read
	)
	fileExtensions, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return nil, fmt.Errorf("cannot get file extensions for content type '%s': %v", contentType, err)
	}
	if len(fileExtensions) == 0 {
		return nil, fmt.Errorf("file extension not found for content type '%s'", contentType)
	}
	fileExt := normalizeFileExtension(fileExtensions[0])

	// get other file name for base security
	fileId, err := formMediaFileId(fileName, sectionName, contentType)
	if err != nil {
		return nil, fmt.Errorf("cannot form media file id: %v", err)
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

func formMediaFileId(fileName, sectionName, contentType string) (string, error) {
	sb := strings.Builder{}
	sb.WriteString(fileName)
	sb.WriteString(sectionName)
	sb.WriteString(contentType)

	fileInfo := []byte(sb.String())
	hashedFileInfo := fmt.Sprintf("%x", sha256.Sum256(fileInfo))

	return hashedFileInfo, nil
}

func normalizeFileExtension(fileExtension string) string {
	fileExtension = strings.Trim(fileExtension, ".")
	correctExtensionsMap := map[string]string{
		"jpe": "jpg",
	}
	if correctExtension, ok := correctExtensionsMap[fileExtension]; ok {
		return correctExtension
	}
	return fileExtension
}
