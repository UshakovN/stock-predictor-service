package storage

import (
	"context"
	"fmt"
	"main/internal/domain"
	"sync"
	"sync/atomic"

	sq "github.com/Masterminds/squirrel"
	"github.com/UshakovN/stock-predictor-service/postgres"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

const counterInc = 1

type queryBuilder interface {
	ToSql() (string, []any, error)
}

type Storage interface {
	PutTicker(ticker *domain.Ticker) error
	PutTickerDetails(ticker *domain.TickerDetails) error
	PutStock(stock *domain.Stock) error
	PutFetcherState(state *domain.FetcherState) error
	GetFetcherState() (*domain.FetcherState, bool, error)
}

type storage struct {
	ctx      context.Context
	client   postgres.Client
	counters *storageCounters
}

type storageCounters struct {
	stock         atomic.Uint64
	ticker        atomic.Uint64
	tickerDetails atomic.Uint64
}

func NewStorage(ctx context.Context, config *postgres.Config) (Storage, error) {
	client, err := postgres.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot create new postgres client: %v", err)
	}
	return &storage{
		ctx:      ctx,
		client:   client,
		counters: newStorageCounters(),
	}, nil
}

func newStorageCounters() *storageCounters {
	return &storageCounters{
		stock:         atomic.Uint64{},
		ticker:        atomic.Uint64{},
		tickerDetails: atomic.Uint64{},
	}
}

func (s *storage) PutTicker(ticker *domain.Ticker) error {
	if ticker == nil {
		return fmt.Errorf("ticker is a nil")
	}
	builder := sq.Insert(`ticker`).
		Columns(
			`ticker_id`,
			`company_name`,
			`company_locale`,
			`currency_name`,
			`ticker_cik`,
			`active`,
			`created_at`,
			`external_updated_at`,
		).
		Values(
			ticker.TickerId,
			ticker.CompanyName,
			ticker.CompanyLocale,
			ticker.CurrencyName,
			ticker.TickerCik,
			ticker.Active,
			ticker.CreatedAt,
			ticker.ExternalUpdatedAt,
		).
		Suffix(`ON CONFLICT (ticker_id) DO NOTHING`).
		PlaceholderFormat(sq.Dollar)

	if err := s.doPutQuery(builder); err != nil {
		return err
	}
	log.Infof("put ticker '%s' for company '%s' to storage. total: %d",
		ticker.TickerId, ticker.CompanyName, s.counters.ticker.Add(counterInc))

	return nil
}

func (s *storage) PutTickerDetails(tickerDetails *domain.TickerDetails) error {
	if tickerDetails == nil {
		return fmt.Errorf("ticker details is a nil")
	}
	builder := sq.Insert(`ticker_details`).
		Columns(
			`ticker_id`,
			`company_description`,
			`homepage_url`,
			`phone_number`,
			`total_employees`,
			`company_state`,
			`company_city`,
			`company_address`,
			`company_postal_code`,
		).
		Values(
			tickerDetails.TickerId,
			tickerDetails.CompanyDescription,
			tickerDetails.HomepageUrl,
			tickerDetails.PhoneNumber,
			tickerDetails.TotalEmployees,
			tickerDetails.CompanyState,
			tickerDetails.CompanyCity,
			tickerDetails.CompanyAddress,
			tickerDetails.CompanyPostalCode,
		).
		Suffix(`ON CONFLICT (ticker_id) DO NOTHING`).
		PlaceholderFormat(sq.Dollar)

	if err := s.doPutQuery(builder); err != nil {
		return err
	}
	log.Infof("put ticker details for ticker '%s' to storage. total: %d",
		tickerDetails.TickerId, s.counters.tickerDetails.Add(counterInc))

	return nil
}

func (s *storage) PutStock(stock *domain.Stock) error {
	if stock == nil {
		return fmt.Errorf("stock is a nil")
	}
	builder := sq.Insert(`stock`).
		Columns(
			`stock_id`,
			`ticker_id`,
			`open_price`,
			`close_price`,
			`highest_price`,
			`lowest_price`,
			`trading_volume`,
			`stocked_at`,
			`created_at`,
		).
		Values(
			stock.StockId,
			stock.TickerId,
			stock.OpenPrice,
			stock.ClosePrice,
			stock.HighestPrice,
			stock.LowestPrice,
			stock.TradingVolume,
			stock.StockedAt,
			stock.CreatedAt,
		).
		Suffix(`ON CONFLICT (stock_id) DO NOTHING`).
		PlaceholderFormat(sq.Dollar)

	if err := s.doPutQuery(builder); err != nil {
		return err
	}
	log.Infof("put stock '%s' for ticker '%s' in storage. total: %d",
		stock.StockId, stock.TickerId, s.counters.stock.Add(counterInc))

	return nil
}

func (s *storage) doPutQuery(builder queryBuilder) error {
	query, args := mustBuildQuery(builder)
	if _, err := s.client.Exec(s.ctx, query, args...); err != nil {
		return fmt.Errorf("cannot do exec: %v", err)
	}
	return nil
}

func mustBuildQuery(builder queryBuilder) (string, []any) {
	query, args, err := builder.ToSql()
	if err != nil {
		log.Fatalf("build insert query failed: %v", err)
	}
	return query, args
}

func (s *storage) PutFetcherState(state *domain.FetcherState) error {
	builder := sq.Insert(`fetcher_state`).
		Columns(
			// `fetcher_state_id` is serial type, autoincrement
			`ticker_req_url`,
			`ticker_details_req_url`,
			`stock_req_url`,
			`created_at`,
			`finished`,
		).
		Values(
			state.TickerReqUrl,
			state.TickerDetailsReqUrl,
			state.StockReqUrl,
			state.CreatedAt,
			state.Finished,
		).
		PlaceholderFormat(sq.Dollar)

	if err := s.doPutQuery(builder); err != nil {
		return err
	}
	log.Infof("sucessfully put fetcher state to storage")

	return nil
}

func (s *storage) GetFetcherState() (*domain.FetcherState, bool, error) {
	const (
		queryLimit = 5
	)
	builder := sq.Select(
		`state_id`,
		`ticker_req_url`,
		`ticker_details_req_url`,
		`stock_req_url`,
		`created_at`,
		`finished`,
	).
		From(`fetcher_state`).
		OrderBy(`created_at DESC`).
		Limit(queryLimit).
		PlaceholderFormat(sq.Dollar)

	state := &domain.FetcherState{}
	found, err := s.doGetQuery(builder,
		&state.StateId,
		&state.TickerReqUrl,
		&state.TickerDetailsReqUrl,
		&state.StockReqUrl,
		&state.CreatedAt,
		&state.Finished,
	)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	log.Infof("sucessfully get fetcher state at '%s' from storage",
		state.CreatedAt)

	return state, true, nil
}

func (s *storage) doGetQuery(builder queryBuilder, fields ...any) (bool, error) {
	query, args := mustBuildQuery(builder)
	rows, err := s.client.Query(s.ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("cannot do query: %v", err)
	}
	return scanFirstQueriedRow(rows, fields)
}

func scanFirstQueriedRow(rows pgx.Rows, fields []any) (bool, error) {
	var (
		hasRows bool
		err     error
	)
	once := sync.Once{}
	for rows.Next() {
		once.Do(func() {
			err = rows.Scan(fields...)
			hasRows = true
		})
	}
	return hasRows, err
}
