package fetcher

import (
  "context"
  "main/internal/queue"
  "main/internal/storage"
  "sync"

  "github.com/UshakovN/stock-predictor-service/httpclient"
)

type Fetcher interface {
  ContinuouslyFetch()
  SaveFetcherState()
  SetTickerId(tickerId string)
}

type fetcher struct {
  ctx      context.Context
  client   httpclient.HttpClient
  storage  storage.Storage
  msQueue  queue.MediaServiceQueue
  state    *state
  once     *sync.Once
  tickerId string
}

func NewFetcher(ctx context.Context, config *Config) (Fetcher, error) {
  client := httpclient.NewClient(
    httpclient.WithContext(ctx),
    httpclient.WithQueryApiToken(
      apiTokenKey,
      config.ApiToken,
    ),
    httpclient.WithRequestsLimit(
      polygonReqsLimit,
      polygonReqPerDur,
      polygonWaitDur,
      polygonDeadlineDur,
    ))

  fetcherState := newFetcherState(
    config.ModeTotalHours,
    config.ModeCurrentHours,
  )

  fetcherStorage, err := storage.NewStorage(ctx, config.StorageConfig)
  if err != nil {
    return nil, err
  }
  msQueue, err := queue.NewMediaServiceQueue(ctx, config.QueueConfig)
  if err != nil {
    return nil, err
  }

  return &fetcher{
    ctx:     ctx,
    client:  client,
    storage: fetcherStorage,
    msQueue: msQueue,
    state:   fetcherState,
    once:    &sync.Once{},
  }, nil
}
