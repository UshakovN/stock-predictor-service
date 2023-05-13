package handler

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/queue"
  "main/internal/service"
  "main/internal/storage"
  "net/http"
  "strings"

  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  mediaservice "github.com/UshakovN/stock-predictor-service/contract/media-service"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/hash"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

const fromMediaServiceHttp = "mediaservice_http"

type Handler struct {
  ctx        context.Context
  service    service.MediaService
  auth       authservice.Client
  hostPrefix string
}

func NewHandler(ctx context.Context, hostPrefix string, config *Config) (*Handler, error) {
  msQueue, err := queue.NewMediaServiceQueue(ctx, config.QueueConfig)
  if err != nil {
    return nil, fmt.Errorf("cannot create new media service queue: %v", err)
  }
  msStorage, err := storage.NewStorage(ctx, config.StorageConfig)
  hashManager := hash.NewManager(config.MediaServiceHashSalt)

  mediaService, err := service.NewMediaService(ctx, &service.Config{
    MsQueue:     msQueue,
    Storage:     msStorage,
    HashManager: hashManager,
  })
  if err != nil {
    return nil, fmt.Errorf("cannot create new media service: %v", err)
  }
  auth := authservice.NewClient(ctx, config.AuthServicePrefix, config.MediaServiceApiToken)

  return &Handler{
    ctx:        ctx,
    service:    mediaService,
    auth:       auth,
    hostPrefix: hostPrefix, // inject host prefix for serve media content
  }, nil
}

func (h *Handler) BindRouter() {
  http.Handle("/stored_media/", bindFileServer())

  http.Handle("/get", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleGet)))
  http.Handle("/get-batch", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleGetBatch)))
  http.Handle("/put-queue", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandlePutQueue)))
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))

  log.Printf("handler router is configured")
}

func bindFileServer() http.Handler {
  fs := http.FileServer(http.Dir(service.DirStoredMedia))
  return http.StripPrefix("/stored_media/", fs)
}

// HandleGet
//
// @Summary Get method for stored media
// @Description Get method provide serving url to single stored media
// @Tags Media
// @Produce            application/json
// @Param request body mediaservice.GetRequest true "Request"
// @Success 200 {object} mediaservice.GetResponse
// @Failure 400, 401, 403, 500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /get [post]
//
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) error {
  req := &mediaservice.GetRequest{}
  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := req.Validate(); err != nil {
    return err
  }
  media, err := h.service.GetMedia(&domain.GetMediaInput{
    Name:      req.Name,
    Section:   req.Section,
    From:      fromMediaServiceHttp,
    Timestamp: utils.NowTimestampUTC(),
  })
  if err != nil {
    return err
  }
  if err = utils.WriteResponse(w, h.formGetResponse(media), http.StatusOK); err != nil {
    return nil
  }
  return nil
}

// HandleGetBatch
//
// @Summary GetBatch method for stored media
// @Description GetBatch method provide serving urls to batch of stored media
// @Tags Media
// @Produce            application/json
// @Param request body mediaservice.GetBatchRequest true "Request"
// @Success 200 {object} mediaservice.GetBatchResponse
// @Failure 400, 401, 403, 500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /get-batch [post]
//
func (h *Handler) HandleGetBatch(w http.ResponseWriter, r *http.Request) error {
  req := &mediaservice.GetBatchRequest{}
  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := req.Validate(); err != nil {
    return err
  }
  inputs := make([]*domain.GetMediaInput, 0, len(req.Parts))

  for _, part := range req.Parts {
    inputs = append(inputs, &domain.GetMediaInput{
      Name:      part.Name,
      Section:   part.Section,
      From:      fromMediaServiceHttp,
      Timestamp: utils.NowTimestampUTC(),
    })
  }
  mediaBatch, err := h.service.GetMediaBatch(inputs)
  if err != nil {
    return err
  }

  if err := utils.WriteResponse(w, h.formGetBatchResponse(mediaBatch), http.StatusOK); err != nil {
    return err
  }
  return nil
}

// HandlePutQueue
//
// @Summary PutQueue method for asynchronous media storing
// @Description PutQueue method put media in handling queue for asynchronous storing
// @Tags Media
// @Produce            application/json
// @Param request body mediaservice.PutRequest true "Request"
// @Success 200 {object} mediaservice.PutResponse
// @Failure 400, 401, 403, 500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /put-queue [post]
//
func (h *Handler) HandlePutQueue(w http.ResponseWriter, r *http.Request) error {
  const queuedRespField = true

  req := &mediaservice.PutRequest{}
  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := req.Validate(); err != nil {
    return err
  }

  if err := h.service.PutMedia(&domain.PutMediaInput{
    Name:      req.Name,
    Section:   req.Section,
    Content:   req.Content,
    Overwrite: req.Overwrite,
    From:      fromMediaServiceHttp,
    Timestamp: utils.NowTimestampUTC(),
  }); err != nil {
    return err
  }

  if err := utils.WriteResponse(w, &mediaservice.PutResponse{
    Queued: queuedRespField,
  }, http.StatusAccepted); err != nil {
    return err
  }

  return nil
}

// HandleHealth
//
// @Summary Health check method
// @Description Health method check http server health
// @Tags Health
// @Produce application/json
// @Success 200 {object} common.HealthResponse
// @Success 500 {object} errs.Error
// @Router /health [get]
//
func (h *Handler) HandleHealth(w http.ResponseWriter, _ *http.Request) error {
  if err := utils.WriteResponse(w, &common.HealthResponse{
    Success: true,
  }, http.StatusOK); err != nil {
    return err
  }
  return nil
}

func (h *Handler) ContinuouslyServeHttp(port string) {
  err := http.ListenAndServe(fmt.Sprint(":", port), nil)
  if err != nil {
    log.Fatalf("listen and serve error: %v", err)
  }
}

func (h *Handler) ContinuouslyServeQueue() {
  err := h.service.HandleQueueMessages()
  if err != nil {
    log.Fatalf("handle queue messages error: %v", err)
  }
}

func (h *Handler) formGetResponse(media *domain.Media) *mediaservice.GetResponse {
  var sourceUrl string

  if media.Found {
    sourceUrl = formMediaFileUrl(h.hostPrefix, media.Path)
  }
  return &mediaservice.GetResponse{
    Found:     media.Found,
    Name:      media.Name,
    Section:   media.Section,
    SourceUrl: sourceUrl,
  }
}

func (h *Handler) formGetBatchResponse(mediaBatch []*domain.Media) *mediaservice.GetBatchResponse {
  parts := make([]*mediaservice.GetResponse, 0, len(mediaBatch))

  for _, media := range mediaBatch {
    parts = append(parts, h.formGetResponse(media))
  }
  return &mediaservice.GetBatchResponse{
    Parts: parts,
  }
}

func formMediaFileUrl(hostPrefix, filePath string) string {
  const (
    storedMediaDirPrefix = "./"
    fileUrlTemplate      = "%s/%s" // host-prefix/file-path-with-extension
  )
  filePath = strings.TrimPrefix(filePath, storedMediaDirPrefix)

  return fmt.Sprintf(fileUrlTemplate, hostPrefix, filePath)
}
