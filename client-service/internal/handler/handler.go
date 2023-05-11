package handler

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/service"
  "main/internal/storage"
  "net/http"

  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  clientservice "github.com/UshakovN/stock-predictor-service/contract/client-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  mediaservice "github.com/UshakovN/stock-predictor-service/contract/media-service"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type Handler struct {
  ctx     context.Context
  service service.ClientService
  auth    authservice.Client
}

func (h *Handler) BindRouter() {
  http.Handle("/tickers", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleTickers)))
  http.Handle("/stocks", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleStocks)))
  http.Handle("/subscriptions", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleSubscriptions)))
  http.Handle("/subscribe", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleSubscribe)))
  http.Handle("/unsubscribe", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleUnsubscribe)))
  //http.Handle("/predicts", nil) // TODO: implement /predicts handler
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))
}

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  newStorage, err := storage.NewStorage(ctx, config.StorageConfig)
  if err != nil {
    return nil, fmt.Errorf("cannot create new storage: %v", err)
  }
  mediaClient := mediaservice.NewClient(ctx, config.MediaServicePrefix, config.MediaServiceApiToken)

  newService := service.NewClientService(ctx, &service.Config{
    Storage:     newStorage,
    MediaClient: mediaClient,
  })
  authClient := authservice.NewClient(ctx, config.AuthServicePrefix, config.ClientServiceApiToken)

  return &Handler{
    ctx:     ctx,
    service: newService,
    auth:    authClient,
  }, nil
}

func (h *Handler) HandleTickers(w http.ResponseWriter, r *http.Request) error {
  req := &clientservice.TickersRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := confirmResourceRequest(r, req.ResourceRequest); err != nil {
    return err
  }
  input := &domain.GetInput{}

  if err := utils.FillFrom(req, input); err != nil {
    return err
  }
  tickers, err := h.service.GetTickers(input)
  if err != nil {
    return err
  }
  resp := &clientservice.TickersResponse{}

  if err := utils.FillFrom(tickers, &resp.Tickers); err != nil {
    return err
  }
  resp.ResourceResponse = &clientservice.ResourceResponse{
    Success: true,
    Count:   len(resp.Tickers),
  }
  if err := utils.WriteResponse(w, resp, http.StatusOK); err != nil {
    return err
  }
  return nil
}

func (h *Handler) HandleStocks(w http.ResponseWriter, r *http.Request) error {
  req := &clientservice.StocksRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := confirmResourceRequest(r, req.ResourceRequest); err != nil {
    return err
  }
  input := &domain.GetInput{}

  if err := utils.FillFrom(req, input); err != nil {
    return err
  }
  stocks, err := h.service.GetStocks(input)
  if err != nil {
    return err
  }
  resp := &clientservice.StocksResponse{}

  if err := utils.FillFrom(stocks, &resp.Stocks); err != nil {
    return err
  }
  resp.ResourceResponse = &clientservice.ResourceResponse{
    Success: true,
    Count:   len(resp.Stocks),
  }
  if err := utils.WriteResponse(w, resp, http.StatusOK); err != nil {
    return err
  }
  return nil
}

func (h *Handler) HandleSubscribe(w http.ResponseWriter, r *http.Request) error {
  userId, err := getUserIdFromReqCtx(r)
  if err != nil {
    return err
  }
  req := &clientservice.SubscribeRequest{}

  if err = utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err = req.Validate(); err != nil {
    return err
  }

  if err = h.service.Subscribe(userId, req.TickerId); err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &clientservice.SubscribeResponse{
    Success: true,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) error {
  userId, err := getUserIdFromReqCtx(r)
  if err != nil {
    return err
  }
  req := &clientservice.UnsubscribeRequest{}

  if err = utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err = req.Validate(); err != nil {
    return err
  }
  if err = h.service.Unsubscribe(userId, req.TickerId); err != nil {
    return err
  }

  if err = utils.WriteResponse(w, &clientservice.UnsubscribeResponse{
    Success: true,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleSubscriptions(w http.ResponseWriter, r *http.Request) error {
  userId, err := getUserIdFromReqCtx(r)
  if err != nil {
    return err
  }
  req := &clientservice.SubscriptionsRequest{}

  if err = utils.ReadRequest(r, req); err != nil {
    return err
  }
  subs, err := h.service.GetSubscriptions(userId, req.FilterActive)
  if err != nil {
    return err
  }
  resp := &clientservice.SubscriptionsResponse{
    Success: true,
  }
  if err = utils.FillFrom(subs, &resp.Parts); err != nil {
    return err
  }

  if err := utils.WriteResponse(w, resp, http.StatusOK); err != nil {
    return err
  }
  return nil
}

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

func confirmResourceRequest(r *http.Request, req *clientservice.ResourceRequest) error {
  serviceAccess, err := getServiceAccessFromReqCtx(r)
  if err != nil {
    return fmt.Errorf("cannot get service access from request content: %v", err)
  }
  if serviceAccess {
    return nil
  }
  if req == nil {
    return errs.NewError(errs.ErrTypeMalformedRequest, nil)
  }
  if req.Pagination == nil {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "pagination must be specified", nil)
  }
  return nil
}

func getServiceAccessFromReqCtx(req *http.Request) (bool, error) {
  serviceAccess, err := utils.GetCtxValue[bool](req.Context(), authservice.CtxKeyServiceAccess{})
  if err != nil {
    if errs.ErrIs(err, utils.ErrCtxValueNotFound) {
      return false, nil
    }
    return false, err
  }
  return serviceAccess, nil
}

func getUserIdFromReqCtx(req *http.Request) (string, error) {
  userId, err := utils.GetCtxValue[string](req.Context(), authservice.CtxKeyUserId{})
  if err != nil {
    return "", fmt.Errorf("cannot get user id from request context: %v", err)
  }
  return userId, nil
}
