package service

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/storage"

  media_service "github.com/UshakovN/stock-predictor-service/contract/media-service"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/httpclient"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type ClientService interface {
  GetTickers(input *domain.GetInput) ([]*domain.Ticker, error)
  GetStocks(input *domain.GetInput) ([]*domain.Stock, error)
  Subscribe(userId, tickerId string) error
  Unsubscribe(userId, tickerId string) error
  Subscriptions(userId string, filterActive bool) ([]*domain.Subscription, error)
}

type service struct {
  ctx       context.Context
  storage   storage.Storage
  apiClient httpclient.HttpClient
}

func NewClientService(ctx context.Context, config *Config) ClientService {
  return &service{
    ctx:       ctx,
    storage:   config.Storage,
    apiClient: config.ApiClient,
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
    storage.ErrFiltersHasDuplicates,
    storage.ErrOnlyOneFilterType,
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

func (s *service) getTickersMediaBatchResp(tickers []*domain.Ticker) (*media_service.GetBatchResponse, error) {
  req := s.formMediaServiceRequest(tickers)

  content, err := s.apiClient.Post("/get-batch", req, nil)
  if err != nil {
    return nil, fmt.Errorf("cannot get response from media service: %v", err)
  }
  resp := &media_service.GetBatchResponse{}

  if err = s.apiClient.ParseResponse(content, resp); err != nil {
    return nil, fmt.Errorf("cannot parse media service response: %v", err)
  }

  return resp, nil
}

func (s *service) formMediaServiceRequest(tickers []*domain.Ticker) *media_service.GetBatchRequest {
  const (
    logoNameTemplate  = "%s-logo.svg"
    referencesSection = "polygon_references"
  )
  parts := make([]*media_service.GetRequest, 0, len(tickers))

  for _, ticker := range tickers {
    if ticker.Fields == nil {
      continue
    }
    tickerId := ticker.Fields.TickerId
    logoName := fmt.Sprintf(logoNameTemplate, tickerId)

    parts = append(parts, &media_service.GetRequest{
      Name:    logoName,
      Section: referencesSection,
    })
  }
  return &media_service.GetBatchRequest{
    Parts: parts,
  }
}

func (s *service) Subscribe(userId string, tickerId string) error {
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

func (s *service) Subscriptions(userId string, filterActive bool) ([]*domain.Subscription, error) {
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

func formSubscription(stored *storage.Subscription) *domain.Subscription {
  return &domain.Subscription{
    SubscriptionId: stored.SubscriptionId,
    UserId:         stored.UserId,
    TickerId:       stored.TickerId,
    Active:         stored.Active,
    CreatedAt:      stored.CreatedAt,
    ModifiedAt:     stored.ModifiedAt,
  }
}
