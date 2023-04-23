package client_service

import (
  "time"

  "github.com/UshakovN/stock-predictor-service/errs"
)

type ResourceRequest struct {
  Pagination *Pagination `json:"pagination,omitempty"`
  Sort       *Sort       `json:"sort,omitempty"`
  Filters    []*Filter   `json:"filters,omitempty"`
}

type ResourceResponse struct {
  Success bool `json:"success"`
  Count   int  `json:"count"`
}

type TickersRequest struct {
  *ResourceRequest
  With *WithFields `json:"with,omitempty"`
}

type Pagination struct {
  Page  int `json:"page"`
  Count int `json:"count"`
}

type Sort struct {
  Field string `json:"field"`
  Order string `json:"order"`
}

type Filter struct {
  Border  *BorderFilter  `json:"border,omitempty"`
  Between *BetweenFilter `json:"between,omitempty"`
}

type BorderFilter struct {
  Field   string `json:"field"`
  Value   any    `json:"value"`
  Compare string `json:"compare"`
}

type BetweenFilter struct {
  Field       string `json:"field"`
  LeftBorder  any    `json:"left_border"`
  RightBorder any    `json:"right_border"`
}

type WithFields struct {
  Media bool `json:"media"`
}

type Ticker struct {
  Fields *TickerFields `json:"fields"`
  Media  *TickerMedia  `json:"media,omitempty"`
}

type TickerFields struct {
  TickerId           string    `json:"ticker_id"`
  CompanyName        string    `json:"company_name"`
  CompanyLocale      string    `json:"company_locale"`
  CompanyDescription string    `json:"company_description"`
  CompanyState       string    `json:"company_state"`
  CompanyCity        string    `json:"company_city"`
  CompanyAddress     string    `json:"company_address"`
  HomepageUrl        string    `json:"homepage_url"`
  CurrencyName       string    `json:"currency_name"`
  TotalEmployees     int       `json:"total_employees"`
  Active             bool      `json:"active"`
  CreatedAt          time.Time `json:"created_at"`
}

type TickerMedia struct {
  Found bool   `json:"found"`
  Url   string `json:"url"`
}

type TickersResponse struct {
  *ResourceResponse
  Tickers []*Ticker `json:"tickers"`
}

type StocksRequest struct {
  *ResourceRequest
}

type Stock struct {
  Ticker        string    `json:"ticker_id"`
  OpenPrice     float64   `json:"open_price"`
  ClosePrice    float64   `json:"close_price"`
  HighestPrice  float64   `json:"highest_price"`
  LowestPrice   float64   `json:"lowest_price"`
  TradingVolume float64   `json:"trading_volume"`
  StockedTime   time.Time `json:"stocked_time"`
  CreatedAt     time.Time `json:"created_at"`
}

type StocksResponse struct {
  *ResourceResponse
  Stocks []*Stock `json:"stocks"`
}

type SubscribeRequest struct {
  TickerId string `json:"ticker_id"`
}

type SubscribeResponse struct {
  Success bool `json:"success"`
}

type UnsubscribeRequest struct {
  TickerId string `json:"ticker_id"`
}

type UnsubscribeResponse struct {
  Success bool `json:"success"`
}

func (r *SubscribeRequest) Validate() error {
  if r.TickerId == "" {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "ticker_id field must be specified", nil)
  }
  return nil
}

func (r *UnsubscribeRequest) Validate() error {
  if r.TickerId == "" {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "ticker_id field must be specified", nil)
  }
  return nil
}

type SubscriptionsRequest struct{}

type Subscription struct {
  SubscriptionId string    `json:"subscription_id"`
  TickerId       string    `json:"ticker_id"`
  Active         bool      `json:"active"`
  CreatedAt      time.Time `json:"created_at"`
  ModifiedAt     time.Time `json:"modified_at"`
}

type SubscriptionsResponse struct {
  Success bool            `json:"success"`
  Parts   []*Subscription `json:"parts"`
}
