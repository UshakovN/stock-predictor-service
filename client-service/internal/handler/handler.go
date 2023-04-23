package handler

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/service"
  "main/internal/storage"
  "net/http"

  "github.com/UshakovN/stock-predictor-service/contract/client-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/httpclient"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type Handler struct {
  ctx     context.Context
  service service.ClientService
}

func (h *Handler) BindRouter() {
  http.Handle("/tickers", errs.MiddlewareErr(h.HandleTickers))
  http.Handle("/stocks", errs.MiddlewareErr(h.HandleStocks))
  http.Handle("/subscriptions", nil)
  http.Handle("/subscribe", errs.MiddlewareErr(h.HandleSubscribe))
  http.Handle("/unsubscribe", errs.MiddlewareErr(h.HandleUnsubscribe))
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))
}

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  newStorage, err := storage.NewStorage(ctx, config.StorageConfig)
  if err != nil {
    return nil, fmt.Errorf("cannot create exchange storage: %v", err)
  }
  apiClient := httpclient.NewClient(
    httpclient.WithContext(ctx),
    httpclient.WithApiPrefix(config.MediaServicePrefix),
  )
  clientService := service.NewClientService(ctx, &service.Config{
    Storage:   newStorage,
    ApiClient: apiClient,
  })

  return &Handler{
    ctx:     ctx,
    service: clientService,
  }, nil
}

func (h *Handler) HandleTickers(w http.ResponseWriter, r *http.Request) error {
  req := &client_service.TickersRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
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
  resp := &client_service.TickersResponse{}

  if err := utils.FillFrom(tickers, &resp.Tickers); err != nil {
    return err
  }
  resp.ResourceResponse = &client_service.ResourceResponse{
    Success: true,
    Count:   len(resp.Tickers),
  }
  if err := utils.WriteResponse(w, resp, http.StatusOK); err != nil {
    return err
  }
  return nil
}

func (h *Handler) HandleStocks(w http.ResponseWriter, r *http.Request) error {
  req := &client_service.StocksRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
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
  resp := &client_service.StocksResponse{}

  if err := utils.FillFrom(stocks, &resp.Stocks); err != nil {
    return err
  }
  resp.ResourceResponse = &client_service.ResourceResponse{
    Success: true,
    Count:   len(resp.Stocks),
  }
  if err := utils.WriteResponse(w, resp, http.StatusOK); err != nil {
    return err
  }
  return nil
}

func (h *Handler) HandleSubscribe(w http.ResponseWriter, r *http.Request) error {
  req := &client_service.SubscribeRequest{}
  var err error

  if err = utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err = req.Validate(); err != nil {
    return err
  }
  // TODO: get user id from auth

  if err = h.service.Subscribe("", req.TickerId); err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &client_service.SubscribeResponse{
    Success: true,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) error {
  req := &client_service.UnsubscribeRequest{}
  var err error

  if err = req.Validate(); err != nil {
    return err
  }
  if err = h.service.Unsubscribe("", req.TickerId); err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &client_service.UnsubscribeResponse{
    Success: true,
  }, http.StatusOK); err != nil {
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
