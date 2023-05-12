package clientservice

import (
  "fmt"
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
  List    *ListFilter    `json:"list,omitempty"`
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

type ListFilter struct {
  Field  string `json:"field"`
  Values []any  `json:"values"`
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

func (r *ResourceRequest) Validate() error {
  const (
    maxCountPerPage = 1000
  )
  if r.Pagination == nil {
    return nil
  }
  if r.Pagination.Count <= 0 {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "pagination count must be positive", nil)
  }
  if r.Pagination.Count > maxCountPerPage {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      fmt.Sprintf("pagination count must be lower than %d", maxCountPerPage), nil)
  }
  if r.Pagination.Page <= 0 {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "pagination page must be positive", nil)
  }
  return nil
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

type SubscriptionsRequest struct {
  FilterActive bool `json:"filter_active"`
}

type Subscription struct {
  TickerId   string    `json:"ticker_id"`
  Active     bool      `json:"active"`
  CreatedAt  time.Time `json:"created_at"`
  ModifiedAt time.Time `json:"modified_at"`
}

type SubscriptionsResponse struct {
  Success bool            `json:"success"`
  Parts   []*Subscription `json:"parts"`
}

type Predict struct {
  TickerId          string    `json:"ticker_id"`
  DatePredict       time.Time `json:"date_predict,omitempty"`
  PredictedMovement string    `json:"predicted_movement,omitempty"`
  CreatedAt         time.Time `json:"created_at,omitempty"`
}

type ModelInfo struct {
  Accuracy  float64   `json:"accuracy"`
  CreatedAt time.Time `json:"created_at"`
}

type PredictsRequest struct{}

type PredictsResponse struct {
  Success   bool       `json:"success"`
  ModelInfo *ModelInfo `json:"model_info"`
  Parts     []*Predict `json:"parts"`
}
