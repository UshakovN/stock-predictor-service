package fetcher

import (
  "fmt"
  "main/internal/domain"
  "net/url"
  "time"

  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

func (f *fetcher) getTickersResponse(query string) (*tickersResponse, error) {
  reqURL := buildRequestURL(f.state.ticker, func() string {
    return fmt.Sprint(basePrefixApi, tickersApi, "?", query)
  })

  resp, err := f.client.Get(reqURL, nil)
  if err != nil {
    return nil, fmt.Errorf("cannot get response: %v", err)
  }
  tickersResp := &tickersResponse{}

  if err = f.client.ParseResponse(resp, tickersResp); err != nil {
    return nil, fmt.Errorf("cannot parse response: %v", err)
  }
  return tickersResp, nil
}

func buildTickersQuery(tickerId string) url.Values {
  query := url.Values{}

  query.Add("active", "true")
  query.Add("order", "asc")

  if tickerId != "" {
    query.Add("ticker", tickerId)
  }
  return query
}

func (f *fetcher) fetchTickerDetails(tickerId string) (*domain.TickerDetails, error) {
  resp, err := f.getTickerDetailsResponse(tickerId)
  if err != nil {
    return nil, fmt.Errorf("cannot get ticker details response: %v", err)
  }
  if resp.Status != respStatusOK {
    return nil, fmt.Errorf("bad response status: %s", resp.Status)
  }
  if resp.Results == nil {
    return nil, fmt.Errorf("ticker details results not found")
  }

  if err = f.sendMessagesToPutTickerBranding(tickerId, resp.Results.Branding); err != nil {
    log.Errorf("cannot send messages to put branding for ticker '%s': %v", tickerId, err)
  }
  details, err := createTickerDetails(resp.Results)
  if err != nil {
    return nil, fmt.Errorf("cannot create ticker details: %v", err)
  }

  return details, nil
}

func (f *fetcher) getTickerDetailsResponse(tickerId string) (*tickerDetailsResponse, error) {
  reqURL := buildRequestURL(f.state.tickerDetails, func() string {
    tickerDetailsQuery := fmt.Sprintf(tickerDetailsApi, tickerId)
    return fmt.Sprint(basePrefixApi, tickerDetailsQuery)
  })

  resp, err := f.client.Get(reqURL, nil)
  if err != nil {
    return nil, fmt.Errorf("cannot get response: %v", err)
  }
  tickerDetailsResp := &tickerDetailsResponse{}

  if err = f.client.ParseResponse(resp, tickerDetailsResp); err != nil {
    return nil, fmt.Errorf("cannot parse response: %v", err)
  }
  return tickerDetailsResp, nil
}

func (f *fetcher) fetchTickerDetailsAndStocks(tickerId string) error {
  tickerDetails, err := f.fetchTickerDetails(tickerId)
  if err != nil {
    return fmt.Errorf("cannot fetch ticker details for ticker '%s': %v", tickerId, err)
  }
  if err = f.storage.PutTickerDetails(tickerDetails); err != nil {
    return fmt.Errorf("cannot put ticker details for ticker '%s' to storage: %v", tickerId, err)
  }
  if err = f.fetchStocks(&fetchStocksOption{
    TickerId: tickerId,
  }); err != nil {
    return fmt.Errorf("cannot fetch stocks for ticker '%s': %v", tickerId, err)
  }
  return nil
}

func (f *fetcher) FetchInfo() error { // fetch tickers with details and their stocks
  var err error
  // first we must fetch stocks for stored tickers
  if err = f.fetchStocksForStoredTickers(); err != nil {
    return fmt.Errorf("cannot fetch stocks for stored tickers: %v", err)
  }
  //
  // TODO: remove this
  log.Infof("finished fecth stocks for stored tickers")
  return nil
  // TODO: remove this
  //
  if err = f.fetchTickers(&fetchTickersOption{
    TickerId: f.tickerId, // if ticker id not specified will be fetched all tickers
  }); err != nil {
    return fmt.Errorf("cannot fetch new tickers: %v", err)
  }
  return nil
}

type fetchTickersOption struct {
  TickerId string
}

func (o *fetchTickersOption) Validate() error {
  if o.TickerId == "" {
    log.Warnf("ticker id not specified in fetch tickers option. will be fetched all tickers")
  }
  return nil
}

func (f *fetcher) fetchTickers(options *fetchTickersOption) error {
  if err := options.Validate(); err != nil {
    return fmt.Errorf("fetch ticker option validation failed: %v", err)
  }
  query := buildTickersQuery(options.TickerId)

  for {
    queryStr := query.Encode()

    tickersResp, err := f.getTickersResponse(queryStr)
    if err != nil {
      return fmt.Errorf("cannot get tickers response: %v", err)
    }
    respStatus := tickersResp.Status
    cursorURL := tickersResp.NextUrl

    if respStatus != respStatusOK {
      return fmt.Errorf("bad response status: %s", respStatus)
    }
    if tickersResp.Count == 0 {
      break
    }
    for _, tickerRespResult := range tickersResp.Results {
      if tickerRespResult == nil {
        continue
      }
      ticker, err := createTicker(tickerRespResult)
      if err != nil {
        return fmt.Errorf("cannot create ticker: %v", err)
      }
      if err = f.storage.PutTicker(ticker); err != nil {
        return fmt.Errorf("cannot put ticker to storage: %v", err)
      }
      if err = f.fetchTickerDetailsAndStocks(ticker.TickerId); err != nil {
        return fmt.Errorf("cannot fetch full ticker info: %v", err)
      }
    }
    if cursorURL == "" {
      break
    }
    cursor, err := url.Parse(cursorURL)
    if err != nil {
      return fmt.Errorf("cannot parse cursor URL: %s", cursorURL)
    }
    cursorValue := cursor.Query().Get(respCursorKey)
    if cursorValue == "" {
      break
    }
    query.Set(respCursorKey, cursorValue)
  }
  return nil
}

func (f *fetcher) fetchStocksForStoredTickers() error {
  tickers, err := f.storage.GetTickers()
  if err != nil {
    return fmt.Errorf("cannot get ticker from storage: %v", err)
  }
  mustPopularTickers := map[string]struct{}{ // TODO: remove this
    "AMZN": {},
    "AAPL": {},
    "MSFT": {},
    "IBM":  {},
  }
  for _, ticker := range tickers {
    if _, ok := mustPopularTickers[ticker.TickerId]; !ok { // TODO: remove this
      continue
    }
    if err = f.fetchStocks(&fetchStocksOption{
      TickerId:           ticker.TickerId,
      NotUseRequestState: true,
    }); err != nil {
      return fmt.Errorf("cannot fetch stocks for stored ticker '%s': %v", ticker.TickerId, err)
    }
  }
  log.Infof("stocks successfully fetched for '%d' stored tickers", len(tickers))

  return nil
}

func (f *fetcher) buildStocksReqURL(tickerId string) string {
  nowT := time.Now()
  fromT := nowT
  toT := nowT
  sub := f.state.modeCurrentHours

  if f.state.modeCode == fetcherModeTotal {
    sub = f.state.modeTotalHours
  }
  dur := time.Duration(sub) * time.Hour
  from := fromT.Add(-dur).Format("2006-01-02")
  to := toT.Format("2006-01-02")

  multiplier := 1
  timespan := "day"

  rangeQuery := fmt.Sprintf(stocksApi, tickerId, multiplier, timespan, from, to)
  reqURL := fmt.Sprint(basePrefixApi, rangeQuery)
  return reqURL
}

type fetchStocksOption struct {
  TickerId           string
  NotUseRequestState bool
}

func (o *fetchStocksOption) Validate() error {
  if o.TickerId == "" {
    return fmt.Errorf("ticker id must be specified")
  }
  return nil
}

func (f *fetcher) fetchStocks(option *fetchStocksOption) error {
  if err := option.Validate(); err != nil {
    return fmt.Errorf("fetch stocks option validation failed: %v", err)
  }
  var reqURL string

  if option.NotUseRequestState {
    reqURL = f.buildStocksReqURL(option.TickerId)
  } else {
    reqURL = buildRequestURL(f.state.stocks, func() string {
      return f.buildStocksReqURL(option.TickerId)
    })
  }

  for {
    resp, err := f.client.Get(reqURL, nil)
    if err != nil {
      return fmt.Errorf("cannot get response")
    }
    stockResp := &stocksResponse{}

    if err := f.client.ParseResponse(resp, stockResp); err != nil {
      return fmt.Errorf("cannot parse reponse: %v", err)
    }

    if stockResp.QueryCount == 0 && stockResp.Count == 0 {
      log.Warnf("stock prices not found for ticker: %s", option.TickerId)
      return nil
    }

    for _, stockRes := range stockResp.StockResults {
      stock, err := createStock(option.TickerId, stockRes)
      if err != nil {
        return fmt.Errorf("cannot create stock: %v", err)
      }
      if err = f.storage.PutStock(stock); err != nil {
        return fmt.Errorf("cannot put stock to storage: %v", err)
      }
    }
    if stockResp.NextURL == "" {
      break
    }
    reqURL = stockResp.NextURL
  }

  return nil
}

func createTicker(res *tickerResult) (*domain.Ticker, error) {
  if res == nil {
    return nil, nil
  }
  ticker := &domain.Ticker{
    TickerId:          res.Ticker,
    CompanyName:       res.Name,
    CompanyLocale:     res.Locale,
    CurrencyName:      res.CurrencyName,
    TickerCik:         res.Cik,
    Active:            res.Active,
    CreatedAt:         utils.NotTimeUTC(),
    ExternalUpdatedAt: utils.TimeToUTC(res.LastUpdatedUtc),
  }
  if err := utils.SetDefaultStringValues(ticker, defaultStringValue); err != nil {
    return nil, err
  }
  return ticker, nil
}

func createTickerDetails(res *tickerDetailsResults) (*domain.TickerDetails, error) {
  if res == nil {
    return nil, nil
  }
  details := &domain.TickerDetails{
    TickerId:           res.Ticker,
    CompanyDescription: res.Description,
    HomepageUrl:        res.HomepageUrl,
    PhoneNumber:        res.PhoneNumber,
    TotalEmployees:     res.TotalEmployees,
  }
  if res.Address != nil {
    details.CompanyState = res.Address.State
    details.CompanyCity = utils.TitleString(res.Address.City)
    details.CompanyAddress = utils.TitleString(res.Address.Address1)
    details.CompanyPostalCode = res.Address.PostalCode
    details.CreatedAt = utils.NotTimeUTC()
  }
  if err := utils.SetDefaultStringValues(details, defaultStringValue); err != nil {
    return nil, err
  }
  return details, nil
}

func createStock(tickerId string, res *stockResult) (*domain.Stock, error) {
  if res == nil {
    return nil, nil
  }
  const sepId = "-"
  stock := &domain.Stock{
    StockId:       fmt.Sprint(tickerId, sepId, res.Timestamp),
    TickerId:      tickerId,
    OpenPrice:     res.Open,
    ClosePrice:    res.Close,
    HighestPrice:  res.Highest,
    LowestPrice:   res.Lowest,
    TradingVolume: res.Volume,
    StockedAt:     utils.TimestampToTimeUTC(res.Timestamp),
    CreatedAt:     utils.NotTimeUTC(),
  }
  if err := utils.SetDefaultStringValues(stock, defaultStringValue); err != nil {
    return nil, err
  }
  return stock, nil
}

func buildRequestURL(stateRequest *stateRequest, builder func() string) string {
  var reqURL string
  // if request URL not used and set in state
  if !stateRequest.used && stateRequest.requestURL != "" {
    reqURL = stateRequest.requestURL
    // use it once
    stateRequest.used = true
  } else {
    // else form new request URL
    reqURL = builder()
    // save it in state
    stateRequest.requestURL = reqURL
  }
  return reqURL
}
