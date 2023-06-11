package handler

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/service"
  "main/internal/storage"

  _ "main/docs"

  "net/http"

  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  clientservice "github.com/UshakovN/stock-predictor-service/contract/client-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  mediaservice "github.com/UshakovN/stock-predictor-service/contract/media-service"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/swagger"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type Handler struct {
  ctx     context.Context
  service service.ClientService
  auth    authservice.Client
  swagger *swagger.Handler
}

func (h *Handler) BindRouter() {
  http.Handle("/tickers/pages", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleTickersPages)))
  http.Handle("/tickers", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleTickers)))
  http.Handle("/stocks", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleStocks)))
  http.Handle("/stocks/pages", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleStocksPages)))
  http.Handle("/subscribe", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleSubscribe)))
  http.Handle("/unsubscribe", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleUnsubscribe)))
  http.Handle("/subscriptions", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandleSubscriptions)))
  http.Handle("/predictions", errs.MiddlewareErr(h.auth.AuthMiddleware(h.HandlePredictions)))
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))
  http.Handle("/swagger/", errs.MiddlewareErr(h.swagger.HandleSwagger()))
}

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  newStorage, err := storage.NewStorage(ctx, config.StorageConfig)
  if err != nil {
    return nil, fmt.Errorf("cannot create new storage: %v", err)
  }
  mediaClient := mediaservice.NewClient(ctx,
    config.MediaServicePrefix,
    config.MediaServiceApiToken,
  )
  newService := service.NewClientService(ctx, &service.Config{
    Storage:     newStorage,
    MediaClient: mediaClient,
  })
  authClient := authservice.NewClient(ctx,
    config.AuthServicePrefix,
    config.ClientServiceApiToken,
  )
  return &Handler{
    ctx:     ctx,
    service: newService,
    auth:    authClient,
    swagger: swagger.NewHandler(config.SwaggerConfig),
  }, nil
}

// HandleTickersPages
//
// @Summary Tickers pages method
// @Description Tickers pages method calculate total tickers pages count for specified page size
// @Tags Resources
// @Produce            application/json
// @Param request query clientservice.PagesRequest true "Request"
// @Success 200 {object} clientservice.PagesResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /tickers/pages [get]
//
func (h *Handler) HandleTickersPages(w http.ResponseWriter, r *http.Request) error {
  return h.handleCalculatePages(service.CalculatePagesResourceTicker)(w, r)
}

// HandleTickers
//
// @Summary Tickers model method
// @Description Tickers method provide tickers models for client with pagination, filtration, sorting and media fields
// @Tags Resources
// @Produce            application/json
// @Param request body clientservice.TickersRequest true "Request"
// @Success 200 {object} clientservice.TickersResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /tickers [post]
//
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

// HandleStocksPages
//
// @Summary Stocks pages method
// @Description Stocks pages method calculate total stocks pages count for specified page size
// @Tags Resources
// @Produce            application/json
// @Param request query clientservice.PagesRequest true "Request"
// @Success 200 {object} clientservice.PagesResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /stocks/pages [get]
//
func (h *Handler) HandleStocksPages(w http.ResponseWriter, r *http.Request) error {
  return h.handleCalculatePages(service.CalculatePagesResourceStock)(w, r)
}

// HandleStocks
//
// @Summary Stocks model method
// @Description Stocks method provide stocks models for client with pagination, filtration, sorting
// @Tags Resources
// @Produce            application/json
// @Param request body clientservice.StocksRequest true "Request"
// @Success 200 {object} clientservice.StocksResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /stocks [post]
//
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

// HandleSubscribe
//
// @Summary Subscribe method subscribe client to the ticker
// @Description Subscribe method create subscription model for client with specified ticker and store it
// @Tags Subscriptions
// @Produce            application/json
// @Param request body clientservice.SubscribeRequest true "Request"
// @Success 201 {object} clientservice.SubscribeResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /subscribe [post]
//
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

// HandleUnsubscribe
//
// @Summary Unsubscribe method unsubscribe client from the ticker
// @Description Unsubscribe method deactivate subscription model for client on the ticker and update stored model
// @Tags Subscriptions
// @Produce            application/json
// @Param request body clientservice.UnsubscribeRequest true "Request"
// @Success 200 {object} clientservice.UnsubscribeResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /unsubscribe [post]
//
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

// HandleSubscriptions
//
// @Summary Subscriptions model method
// @Description Subscriptions method provide subscriptions models for client with filtration by active subscriptions
// @Tags Subscriptions
// @Produce            application/json
// @Param request query clientservice.SubscriptionsRequest false "Request"
// @Success 200 {object} clientservice.SubscriptionsResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /subscriptions [get]
//
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

// HandlePredictions
//
// @Summary Predictions model method
// @Description Predictions method provide stocks price dynamic predictions for client tickers subscriptions
// @Tags Subscriptions
// @Produce application/json
// @Success 200 {object} clientservice.PredictsResponse
// @Failure 400,401,403,404,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /predictions [get]
//
func (h *Handler) HandlePredictions(w http.ResponseWriter, r *http.Request) error {
  userId, err := getUserIdFromReqCtx(r)
  if err != nil {
    return err
  }
  req := &clientservice.PredictsRequest{}

  if err = utils.ReadRequest(r, req); err != nil {
    return err
  }
  stocksPredicts, err := h.service.GetStocksPredicts(userId)
  if err != nil {
    return err
  }
  resp := &clientservice.PredictsResponse{
    Success: true,
  }
  if err = utils.FillFrom(stocksPredicts, resp); err != nil {
    return err
  }
  if err = utils.WriteResponse(w, resp, http.StatusOK); err != nil {
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

func (h *Handler) handleCalculatePages(resource string) func(w http.ResponseWriter, r *http.Request) error {
  return func(w http.ResponseWriter, r *http.Request) error {
    req := &clientservice.PagesRequest{}

    if err := utils.ReadRequest(r, req); err != nil {
      return err
    }
    if err := req.Validate(); err != nil {
      return err
    }
    pagesCount, err := h.service.CalculatePages(&domain.CalculatePagesInput{
      Resource: resource,
      PageSize: req.PageSize,
    })
    if err != nil {
      return err
    }
    if err = utils.WriteResponse(w, &clientservice.PagesResponse{
      Success:    true,
      TotalCount: pagesCount,
    }, http.StatusOK); err != nil {
      return err
    }
    return nil
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
