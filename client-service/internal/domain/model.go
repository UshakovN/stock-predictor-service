package domain

import "time"

type Ticker struct {
	Fields *TickerFields `json:"fields"`
	Media  *TickerMedia  `json:"media"`
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
	IconUrl string `json:"icon_url"`
	LogoUrl string `json:"logo_url"`
}

type Stock struct {
	StockId       string    `json:"stock_id"`
	TickerId      string    `json:"ticker_id"`
	OpenPrice     float64   `json:"open_price"`
	ClosePrice    float64   `json:"close_price"`
	HighestPrice  float64   `json:"highest_price"`
	LowestPrice   float64   `json:"lowest_price"`
	TradingVolume float64   `json:"trading_volume"`
	StockedAt     time.Time `json:"stocked_time"`
	CreatedAt     time.Time `json:"created_at"`
}
