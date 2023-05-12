package service

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/storage"

  mediaservice "github.com/UshakovN/stock-predictor-service/contract/media-service"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type ClientService interface {
  GetTickers(input *domain.GetInput) ([]*domain.Ticker, error)
  GetStocks(input *domain.GetInput) ([]*domain.Stock, error)
  Subscribe(userId, tickerId string) error
  Unsubscribe(userId, tickerId string) error
  GetSubscriptions(userId string, filterActive bool) ([]*domain.Subscription, error)
  GetStocksPredicts(userId string) (*domain.StocksPredicts, error)
}

type service struct {
  ctx         context.Context
  storage     storage.Storage
  mediaClient mediaservice.Client
}

func NewClientService(ctx context.Context, config *Config) ClientService {
  return &service{
    ctx:         ctx,
    storage:     config.Storage,
    mediaClient: config.MediaClient,
  }
}

func (s *service) GetTickers(input *domain.GetInput) ([]*domain.Ticker, error) {
  stored, err := s.storage.GetTickers(input.ParseOption())
  if err := handleStorageError(err); err != nil {
    return nil, err
  }
  tickers := make([]*domain.Ticker, 0, len(stored))

  for _, stored := range stored {
    tickers = append(tickers, formTicker(stored))
  }
  if len(tickers) != 0 && input.With.HasMedia() {
    if err = s.fillTickersMedia(tickers); err != nil {
      log.Errorf("cannot fill tickers media. send response without media fields. error: %v", err)
    }
  }

  return tickers, nil
}

func (s *service) GetStocks(input *domain.GetInput) ([]*domain.Stock, error) {
  stored, err := s.storage.GetStocks(input.ParseOption())
  if err := handleStorageError(err); err != nil {
    return nil, err
  }
  stocks := make([]*domain.Stock, 0, len(stored))

  for _, stored := range stored {
    stocks = append(stocks, formStock(stored))
  }
  return stocks, nil
}

func formTicker(stored *storage.Ticker) *domain.Ticker {
  return &domain.Ticker{
    Fields: &domain.TickerFields{
      TickerId:           stored.TickerId,
      CompanyName:        stored.CompanyName,
      CompanyLocale:      stored.CompanyLocale,
      CompanyDescription: stored.CompanyDescription,
      CompanyState:       stored.CompanyState,
      CompanyCity:        stored.CompanyCity,
      CompanyAddress:     stored.CompanyAddress,
      HomepageUrl:        stored.HomepageUrl,
      CurrencyName:       stored.CurrencyName,
      TotalEmployees:     stored.TotalEmployees,
      Active:             stored.Active,
      CreatedAt:          stored.CreatedAt,
    },
  }
}

func formStock(stored *storage.Stock) *domain.Stock {
  return &domain.Stock{
    StockId:       stored.StockId,
    TickerId:      stored.TickerId,
    OpenPrice:     stored.OpenPrice,
    ClosePrice:    stored.ClosePrice,
    HighestPrice:  stored.HighestPrice,
    LowestPrice:   stored.LowestPrice,
    TradingVolume: stored.TradingVolume,
    StockedAt:     stored.StockedAt,
    CreatedAt:     stored.CreatedAt,
  }
}

func handleStorageError(err error) error {
  if errs.ErrIs(err,
    storage.ErrMalformedPagination,
    storage.ErrMalformedFilter,
    storage.ErrMalformedFilter,
    storage.ErrMustContainOneFilterType,
  ) {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest, err.Error(), nil)
  }
  if err != nil {
    return fmt.Errorf("cannot get from storage: %v", err)
  }
  return nil
}

func (s *service) fillTickersMedia(tickers []*domain.Ticker) error {
  resp, err := s.getTickersMediaBatchResp(tickers)
  if err != nil {
    return fmt.Errorf("cannot get tickers media batch: %v", err)
  }
  if len(resp.Parts) != len(tickers) {
    return fmt.Errorf("tickers count and response parts mismatch")
  }
  for idx, ticker := range tickers {
    media := resp.Parts[idx]

    ticker.Media = &domain.TickerMedia{
      Found: media.Found,
      Url:   media.SourceUrl,
    }
  }
  return nil
}

func (s *service) getTickersMediaBatchResp(tickers []*domain.Ticker) (*mediaservice.GetBatchResponse, error) {
  req := s.formMediaServiceRequest(tickers)

  resp, err := s.mediaClient.GetBatch(req)
  if err != nil {
    return nil, fmt.Errorf("media client cannot get batch: %v", err)
  }
  return resp, nil
}

func (s *service) formMediaServiceRequest(tickers []*domain.Ticker) *mediaservice.GetBatchRequest {
  const (
    logoNameTemplate  = "%s-logo.svg"
    referencesSection = "polygon_references"
  )
  parts := make([]*mediaservice.GetRequest, 0, len(tickers))

  for _, ticker := range tickers {
    if ticker.Fields == nil {
      continue
    }
    tickerId := ticker.Fields.TickerId
    logoName := fmt.Sprintf(logoNameTemplate, tickerId)

    parts = append(parts, &mediaservice.GetRequest{
      Name:    logoName,
      Section: referencesSection,
    })
  }
  return &mediaservice.GetBatchRequest{
    Parts: parts,
  }
}

func (s *service) Subscribe(userId string, tickerId string) error {
  if err := s.mustFoundTicker(tickerId); err != nil {
    return err
  }
  subId, err := utils.NewUUID()
  if err != nil {
    return fmt.Errorf("cannot create subscription id: %v", err)
  }
  nowTime := utils.NotTimeUTC()

  if err = s.storage.UpdateSubscription(&storage.Subscription{
    SubscriptionId: subId,
    UserId:         userId,
    TickerId:       tickerId,
    Active:         true,
    CreatedAt:      nowTime,
    ModifiedAt:     nowTime,
  }); err != nil {
    return fmt.Errorf("cannot update storage subscription: %v", err)
  }
  return nil
}

func (s *service) Unsubscribe(userId, tickerId string) error {
  if err := s.mustFoundTicker(tickerId); err != nil {
    return err
  }
  if err := s.storage.UpdateSubscription(&storage.Subscription{
    UserId:     userId,
    TickerId:   tickerId,
    Active:     false,
    ModifiedAt: utils.NotTimeUTC(),
  }); err != nil {
    return fmt.Errorf("cannot update storage subscription: %v", err)
  }
  return nil
}

func (s *service) GetSubscriptions(userId string, filterActive bool) ([]*domain.Subscription, error) {
  stored, err := s.storage.GetSubscriptions(userId, filterActive)
  if err != nil {
    return nil, fmt.Errorf("cannot get storage subscriptions: %v", err)
  }
  subs := make([]*domain.Subscription, 0, len(stored))

  for _, stored := range stored {
    subs = append(subs, formSubscription(stored))
  }
  return subs, nil
}

func (s *service) GetStocksPredicts(userId string) (*domain.StocksPredicts, error) {
  stored, err := s.storage.GetStocksPredicts(userId, utils.NowDateUTC())
  if err != nil {
    if errs.ErrIs(err, storage.ErrNotFoundInStorage) {
      return nil, errs.NewError(errs.ErrTypeNotFoundContent, nil)
    }
    return nil, fmt.Errorf("cannot get stocks predicts: %v", err)
  }
  stocksPredicts := formStockPredictions(stored)

  return stocksPredicts, nil
}

func (s *service) mustFoundTicker(tickerId string) error {
  tickers, err := s.storage.GetTickers(s.storage.GetOptionForTicker(tickerId))
  if err != nil {
    return fmt.Errorf("cannot get tickers from storage: %v", err)
  }
  if len(tickers) == 0 {
    return errs.NewError(errs.ErrTypeNotFoundContent, nil)
  }
  return nil
}

func formSubscription(stored *storage.Subscription) *domain.Subscription {
  return &domain.Subscription{
    TickerId:   stored.TickerId,
    Active:     stored.Active,
    CreatedAt:  stored.CreatedAt,
    ModifiedAt: stored.ModifiedAt,
  }
}

func formStockPredictions(stored *storage.StocksPredicts) *domain.StocksPredicts {
  parts := make([]*domain.Predict, 0, len(stored.Parts))

  for _, part := range stored.Parts {
    parts = append(parts, &domain.Predict{
      TickerId:          part.TickerId,
      DatePredict:       part.DatePredict,
      PredictedMovement: part.PredictedMovement,
      CreatedAt:         part.CreatedAt,
    })
  }
  modelInfo := &domain.ModelInfo{
    Accuracy:  stored.ModelInfo.Accuracy,
    CreatedAt: stored.ModelInfo.CreatedAt,
  }
  return &domain.StocksPredicts{
    ModelInfo: modelInfo,
    Parts:     parts,
  }

}
