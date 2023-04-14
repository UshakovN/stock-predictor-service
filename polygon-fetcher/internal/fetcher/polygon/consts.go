package polygon

import "time"

const fetcherName = "polygon_fetcher"

const (
  fetcherModeTotal   = 0
  fetcherModeCurrent = 1
  fetcherRetryCount  = 10
)

const (
  recentlyFetchedSleepInterval  = 1 * time.Hour
  encounteredErrorSleepInterval = 10 * time.Minute
  recentlyThresholdInterval     = 24 * time.Hour
)

const (
  polygonReqsLimit   = 5
  polygonReqPerDur   = 1 * time.Minute
  polygonWaitDur     = 10 * time.Second
  polygonDeadlineDur = 120 * time.Second
)

const (
  respCursorKey = "cursor"
  respStatusOK  = "OK"
)

const (
  basePrefixApi = "https://api.polygon.io"

  tickersApi       = "/v3/reference/tickers"
  stocksApi        = "/v2/aggs/ticker/%s/range/%d/%s/%s/%s"
  tickerDetailsApi = "/v3/reference/tickers/%s"

  apiTokenKey = "apiKey"
)

const defaultStringValue = "N/A"
