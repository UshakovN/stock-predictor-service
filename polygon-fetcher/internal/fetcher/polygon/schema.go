package polygon

import "time"

type tickerResult struct {
  Ticker         string    `json:"ticker"`
  Name           string    `json:"name"`
  Market         string    `json:"market"`
  Locale         string    `json:"locale"`
  Type           string    `json:"type"`
  Cik            string    `json:"cik"`
  Active         bool      `json:"active"`
  CurrencyName   string    `json:"currency_name"`
  LastUpdatedUtc time.Time `json:"last_updated_utc"`
}

type tickersResponse struct {
  Results []*tickerResult `json:"results"`
  Status  string          `json:"status"`
  Count   int             `json:"count"`
  NextUrl string          `json:"next_url"`
}

type stockResult struct {
  Open      float64 `json:"o"`
  Close     float64 `json:"c"`
  Highest   float64 `json:"h"`
  Lowest    float64 `json:"l"`
  Timestamp int64   `json:"t"`
  Volume    float64 `json:"v"`
}

type stocksResponse struct {
  Adjusted     bool           `json:"adjusted"`
  QueryCount   int            `json:"queryCount"`
  RequestId    string         `json:"request_id"`
  StockResults []*stockResult `json:"results"`
  ResultsCount int            `json:"resultsCount"`
  Status       string         `json:"status"`
  Ticker       string         `json:"ticker"`
}

type tickerDetailsResults struct {
  Active          bool                   `json:"active"`
  Address         *tickerDetailsAddress  `json:"address"`
  Branding        *tickerDetailsBranding `json:"branding"`
  Cik             string                 `json:"cik"`
  CurrencyName    string                 `json:"currency_name"`
  Description     string                 `json:"description"`
  HomepageUrl     string                 `json:"homepage_url"`
  ListDate        string                 `json:"list_date"`
  Locale          string                 `json:"locale"`
  Market          string                 `json:"market"`
  Name            string                 `json:"name"`
  PhoneNumber     string                 `json:"phone_number"`
  PrimaryExchange string                 `json:"primary_exchange"`
  SicCode         string                 `json:"sic_code"`
  SicDescription  string                 `json:"sic_description"`
  Ticker          string                 `json:"ticker"`
  TickerRoot      string                 `json:"ticker_root"`
  TotalEmployees  int                    `json:"total_employees"`
}

type tickerDetailsAddress struct {
  Address1   string `json:"address1"`
  City       string `json:"city"`
  PostalCode string `json:"postal_code"`
  State      string `json:"state"`
}

type tickerDetailsBranding struct {
  IconUrl string `json:"icon_url"`
  LogoUrl string `json:"logo_url"`
}

type tickerDetailsResponse struct {
  Results *tickerDetailsResults `json:"results"`
  Status  string                `json:"status"`
}
