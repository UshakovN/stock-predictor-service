package handler

import (
	"context"
	"fmt"
	"main/internal/domain"
	"main/internal/queue"
	"main/internal/service"
	"main/internal/storage"
	"net/http"

	"github.com/UshakovN/stock-predictor-service/errs"
	"github.com/UshakovN/stock-predictor-service/utils"
	log "github.com/sirupsen/logrus"
)

const fromMediaServiceHttp = "media_service_http"

type Handler struct {
	ctx     context.Context
	service service.MediaService
}

func NewHandler(ctx context.Context, hostPrefix string, config *Config) (*Handler, error) {
	msQueue, err := queue.NewMediaServiceQueue(ctx, config.QueueConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create new media service queue: %v", err)
	}
	msStorage, err := storage.NewStorage(ctx, config.StorageConfig)

	service, err := service.NewMediaService(ctx, &service.Config{
		MsQueue:    msQueue,
		Storage:    msStorage,
		HostPrefix: hostPrefix, // inject host prefix for serve media content
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create new media service: %v", err)
	}
	return &Handler{
		ctx:     ctx,
		service: service,
	}, nil
}

func (h *Handler) BindRouter() {
	http.Handle("/stored_media/", bindFileServer())

	http.Handle("/get", errs.MiddlewareErr(h.HandleGet))
	http.Handle("/get-batch", errs.MiddlewareErr(h.HandleGetBatch))
	http.Handle("/put-queue", errs.MiddlewareErr(h.HandlePutQueue))
	http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))

	log.Printf("handler router is configured")
}

func bindFileServer() http.Handler {
	fs := http.FileServer(http.Dir(service.DirStoredMedia))
	return http.StripPrefix("/stored_media/", fs)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) error {
	req := &GetRequest{}
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
	if err = utils.WriteResponse(w, formGetResponse(media), http.StatusOK); err != nil {
		return nil
	}
	return nil
}

func (h *Handler) HandleGetBatch(w http.ResponseWriter, r *http.Request) error {
	req := &GetBatchRequest{}
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

	if err := utils.WriteResponse(w, formGetBatchResponse(mediaBatch), http.StatusOK); err != nil {
		return err
	}
	return nil
}

func (h *Handler) HandlePutQueue(w http.ResponseWriter, r *http.Request) error {
	const queuedRespField = true

	req := &PutRequest{}
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

	if err := utils.WriteResponse(w, &PutResponse{
		Queued: queuedRespField,
	}, http.StatusAccepted); err != nil {
		return err
	}

	return nil
}

func (h *Handler) HandleHealth(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("/ok"))
	if err != nil {
		return fmt.Errorf("cannot write to response writer: %v", err)
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

func formGetResponse(media *domain.Media) *GetResponse {
	return &GetResponse{
		Found:     media.Found,
		Name:      media.Name,
		Section:   media.Section,
		SourceUrl: media.SourceUrl,
	}
}

func formGetBatchResponse(mediaBatch []*domain.Media) *GetBatchResponse {
	parts := make([]*GetResponse, 0, len(mediaBatch))

	for _, media := range mediaBatch {
		parts = append(parts, formGetResponse(media))
	}
	return &GetBatchResponse{
		Parts: parts,
	}
}
