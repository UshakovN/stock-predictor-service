package storage

import (
	"context"
	"fmt"

	"github.com/UshakovN/stock-predictor-service/postgres"
	"github.com/jackc/pgx/v4"
)

type ExchangeStorage interface {
	GetTickers(option *GetOption) ([]*Ticker, error)
	GetStocks(option *GetOption) ([]*Stock, error)
}

type storage struct {
	ctx    context.Context
	client postgres.Client
}

func NewStorage(ctx context.Context, config *postgres.Config) (ExchangeStorage, error) {
	client, err := postgres.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot create new postgres client: %v", err)
	}
	return &storage{
		ctx:    ctx,
		client: client,
	}, nil
}

func (s *storage) GetTickers(option *GetOption) ([]*Ticker, error) {
	var (
		query string
		err   error
	)
	query = sanitizeQuery(
		`select (
        ticker_id,
        company_name,
        company_locale,
        company_description,
        company_state,
        company_city,
        company_address,
        homepage_url,
        currency_name,
        total_employees,
        active,
        created_at,
    ) from ticker inner join ticker_details
        on ticker.ticker_id = ticker_details.ticker_id`)

	query, err = option.Stuff(query)
	if err != nil {
		return nil, err
	}
	rows, err := s.doQuery(query)
	if err != nil {
		return nil, fmt.Errorf("cannot do query: %v", err)
	}
	var tickers []*Ticker

	for {
		ticker := &Ticker{}
		found, err := scanQueriedRow(rows,
			&ticker.TickerId,
			&ticker.CompanyName,
			&ticker.CompanyLocale,
			&ticker.CompanyDescription,
			&ticker.CompanyState,
			&ticker.CompanyCity,
			&ticker.CompanyAddress,
			&ticker.HomepageUrl,
			&ticker.CurrencyName,
			&ticker.TotalEmployees,
			&ticker.Active,
			&ticker.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot scan queried row: %v", err)
		}
		if !found {
			break
		}
		tickers = append(tickers, ticker)
	}

	return tickers, nil
}

func (s *storage) GetStocks(options *GetOption) ([]*Stock, error) {
	var (
		query string
		err   error
	)
	query = sanitizeQuery(
		`select (
        stock_id,
        ticker_id, 
        open_price,
        close_price,
        highest_price,
        lowest_price,
        trading_volume,
        stocked_at,
        created_at,
      ) from stock`)

	query, err = options.Stuff(query)
	if err != nil {
		return nil, err
	}
	rows, err := s.doQuery(query)
	if err != nil {
		return nil, fmt.Errorf("cannot do query: %v", err)
	}
	var stocks []*Stock

	for {
		stock := &Stock{}
		found, err := scanQueriedRow(rows,
			&stock.StockId,
			&stock.TickerId,
			&stock.OpenPrice,
			&stock.ClosePrice,
			&stock.HighestPrice,
			&stock.LowestPrice,
			&stock.TradingVolume,
			&stock.StockedAt,
			&stock.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot scan queried row: %v", err)
		}
		if !found {
			break
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func (s *storage) doQuery(query string, args ...any) (pgx.Rows, error) {
	rows, err := s.client.Query(s.ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot do query: %v", err)
	}
	return rows, err
}

func scanQueriedRow(rows pgx.Rows, fields ...any) (bool, error) {
	var hasRow bool
	if rows.Next() {
		if err := rows.Scan(fields...); err != nil {
			return false, fmt.Errorf("cannot scan queried row: %v", err)
		}
		hasRow = true
	}
	return hasRow, nil
}
