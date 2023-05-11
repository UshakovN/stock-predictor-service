package storage

import "time"

type Ticker struct {
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

type Stock struct {
  StockId       string    `json:"stock_id"`
  TickerId      string    `json:"ticker_id"`
  OpenPrice     float64   `json:"open_price"`
  ClosePrice    float64   `json:"close_price"`
  HighestPrice  float64   `json:"highest_price"`
  LowestPrice   float64   `json:"lowest_price"`
  TradingVolume int       `json:"trading_volume"`
  StockedAt     time.Time `json:"stocked_time"`
  CreatedAt     time.Time `json:"created_at"`
}

type Subscription struct {
  SubscriptionId string    `json:"subscription_id"`
  UserId         string    `json:"user_id"`
  TickerId       string    `json:"ticker_id"`
  Active         bool      `json:"active"`
  CreatedAt      time.Time `json:"created_at"`
  ModifiedAt     time.Time `json:"modified_at"`
}

type Predict struct {
}
