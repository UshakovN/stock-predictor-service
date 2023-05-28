package fetcher

import (
  "context"
  "fmt"
  "main/internal/domain"
  "main/internal/queue"
  "main/internal/storage"
  "sync"
  "time"

  "github.com/UshakovN/stock-predictor-service/httpclient"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
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

type state struct {
  finished         bool
  updatedAt        *time.Time
  modeCode         int
  modeTotalHours   int
  modeCurrentHours int
  ticker           *stateRequest
  tickerDetails    *stateRequest
  stocks           *stateRequest
}

type stateRequest struct {
  requestURL string
  used       bool
}

func newFetcherState(modeTotalHours, modeCurrentHours int) *state {
  return &state{
    modeCode:         fetcherModeTotal,
    modeTotalHours:   modeTotalHours,
    modeCurrentHours: modeCurrentHours,
    ticker:           &stateRequest{},
    tickerDetails:    &stateRequest{},
    stocks:           &stateRequest{},
  }
}

func (s *state) SetFinished() {
  s.finished = true
}

func (s *state) ResetFinished() {
  s.finished = false
}

func (s *state) SetUpdatedTime(t time.Time) {
  s.updatedAt = &t
}

func (s *state) SetModeCode(mode int) {
  if mode != fetcherModeTotal && mode != fetcherModeCurrent {
    log.Warnf("invalid fetcher mode code: %d. mode code do not set. possible: %d - total, %d - current",
      mode, fetcherModeTotal, fetcherModeCurrent)
    return
  }
  s.modeCode = mode
  log.Infof("current fetcher mode: %d. possible: %d - total, %d - current",
    mode, fetcherModeTotal, fetcherModeCurrent)
}

func (f *fetcher) SetTickerId(tickerId string) {
  if tickerId == "" {
    log.Warnf("ticker id is empty. ticker id do not set")
    return
  }
  f.tickerId = tickerId
}

func (f *fetcher) hasRecentlyFetched() bool {
  if !f.state.finished || f.state.updatedAt == nil {
    return false
  }
  updatedAt := *f.state.updatedAt
  thresholdTime := updatedAt.Add(recentlyThresholdInterval)
  return thresholdTime.After(time.Now())
}

func (f *fetcher) ContinuouslyFetch() {
  if f.tickerId == "" {
    // if not specified ticker id
    if err := f.loadFetcherState(); err != nil {
      log.Errorf("state loading from storage failed: %v", err)
    }
  }
  f.state.SetModeCode(fetcherModeTotal)

  tryLeft := fetcherRetryCount
  // fetch with retries
  for tryLeft >= 0 {
    if f.hasRecentlyFetched() {
      log.Printf("recently fetched. wait %v before the next fetch",
        recentlyFetchedSleepInterval)

      time.Sleep(recentlyFetchedSleepInterval)
      f.state.ResetFinished() // reset finished field
      continue
    }
    if err := f.FetchInfo(); err != nil {
      log.Errorf("fetching error: %v. wait %v before the next fetch",
        err, encounteredErrorSleepInterval)

      time.Sleep(encounteredErrorSleepInterval)
      tryLeft--
      continue
    }
    f.state.SetUpdatedTime(utils.NotTimeUTC())
    f.state.SetFinished() // set finished field

    // set retry count again
    tryLeft = fetcherRetryCount
    log.Println("successfully fetching finished")

    if f.tickerId != "" {
      return
    }
    f.once.Do(func() {
      f.state.SetModeCode(fetcherModeCurrent)
    })
  }
  log.Fatalf("fetching failed and stopped")
}

func (f *fetcher) SaveFetcherState() {
  if err := f.storage.PutFetcherState(createFetcherState(f.state)); err != nil {
    log.Fatalf("cannot put fetcher state to storage: %v", err)
  }
}

func (f *fetcher) loadFetcherState() error {
  state, found, err := f.storage.GetFetcherState()
  if err != nil {
    return fmt.Errorf("cannot get fetcher state from storage: %v", err)
  }
  if !found {
    return nil
  }
  // set fields from storage state
  f.state.ticker.requestURL = utils.StripString(state.TickerReqUrl)
  f.state.tickerDetails.requestURL = utils.StripString(state.TickerDetailsReqUrl)
  f.state.stocks.requestURL = utils.StripString(state.StockReqUrl)
  f.state.updatedAt = &state.CreatedAt
  f.state.finished = state.Finished

  return nil
}

func createFetcherState(state *state) *domain.FetcherState {
  if state == nil {
    return nil
  }
  return &domain.FetcherState{
    TickerReqUrl:        utils.StripString(state.ticker.requestURL),
    TickerDetailsReqUrl: utils.StripString(state.tickerDetails.requestURL),
    StockReqUrl:         utils.StripString(state.stocks.requestURL),
    CreatedAt:           utils.NotTimeUTC(),
  }
}
