package storage

import (
  "context"
  "fmt"
  "regexp"
  "sync"

  sq "github.com/Masterminds/squirrel"
  "github.com/UshakovN/stock-predictor-service/postgres"
  "github.com/jackc/pgx/v4"
  log "github.com/sirupsen/logrus"
)

type Storage interface {
  GetTickers(option *GetOption) ([]*Ticker, error)
  GetStocks(option *GetOption) ([]*Stock, error)
  UpdateSubscription(sub *Subscription) error
  GetSubscriptions(userId string, filterActive bool) ([]*Subscription, error)
  GetOptionForTicker(tickerId string) *GetOption
}

type queryBuilder interface {
  ToSql() (string, []any, error)
}

type storage struct {
  ctx    context.Context
  client postgres.Client
}

func NewStorage(ctx context.Context, config *postgres.Config) (Storage, error) {
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
    `select
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
        created_at
    from ticker inner join ticker_details
        on ticker_id = ticker_details.ticker_id`)

  query, err = option.Stuff(query)
  if err != nil {
    return nil, err
  }
  query = unambiguousGetTickersQuery(query)

  rows, err := s.doRawQuery(query)
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
    `select 
        stock_id,
        ticker_id, 
        open_price,
        close_price,
        highest_price,
        lowest_price,
        trading_volume,
        stocked_at,
        created_at
      from stock`)

  query, err = options.Stuff(query)
  if err != nil {
    return nil, err
  }
  rows, err := s.doRawQuery(query)
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

func (s *storage) UpdateSubscription(sub *Subscription) error {
  return s.client.BeginTxFunc(s.ctx, pgx.TxOptions{
    IsoLevel: pgx.Serializable,
  },
    func(tx pgx.Tx) error {
      found, err := s.hasSubscription(tx, sub)
      if err != nil {
        return err
      }
      if found {
        if err = s.updateSubscription(tx, sub); err != nil {
          return err
        }
      } else {
        if err = s.putSubscription(tx, sub); err != nil {
          return err
        }
      }
      return nil
    })
}

func (s *storage) putSubscription(tx pgx.Tx, sub *Subscription) error {
  builder := postgres.NewInsertBuilder().
    Into(`subscription`).
    Columns(
      `subscription_id`,
      `user_id`,
      `ticker_id`,
      `active`,
      `created_at`,
      `modified_at`,
    ).
    Values(
      sub.SubscriptionId,
      sub.UserId,
      sub.TickerId,
      sub.Active,
      sub.CreatedAt,
      sub.ModifiedAt,
    )
  if err := s.doExec(tx, builder); err != nil {
    return err
  }
  return nil
}

func (s *storage) updateSubscription(tx pgx.Tx, sub *Subscription) error {
  builder := postgres.NewUpdateBuilder().
    Table(`subscription`).
    SetMap(map[string]any{
      `active`:      sub.Active,
      `modified_at`: sub.ModifiedAt,
    }).
    Where(sq.Eq{
      `user_id`:   sub.UserId,
      `ticker_id`: sub.TickerId,
    })
  if err := s.doExec(tx, builder); err != nil {
    return err
  }
  return nil
}

func (s *storage) hasSubscription(tx pgx.Tx, sub *Subscription) (bool, error) {
  builder := postgres.NewSelectBuilder().
    Columns(`COUNT(*)`).
    From(`subscription`).
    Where(sq.Eq{
      `user_id`:   sub.UserId,
      `ticker_id`: sub.TickerId,
    })
  queriedRows, err := s.doQuery(tx, builder)
  if err != nil {
    return false, err
  }
  var subCount int

  if _, err := scanFirstQueriedRow(queriedRows, &subCount); err != nil {
    return false, err
  }
  return subCount != 0, nil
}

func (s *storage) GetSubscriptions(userId string, filterActive bool) ([]*Subscription, error) {
  builder := postgres.NewSelectBuilder().
    Columns(
      `subscription_id`,
      `user_id`,
      `ticker_id`,
      `active`,
      `created_at`,
      `modified_at`,
    ).
    From(`subscription`)

  where := sq.Eq{
    `user_id`: userId,
  }
  if filterActive {
    where[`active`] = true
  }
  builder = builder.Where(where)

  queriedRows, err := s.doQuery(nil, builder)
  if err != nil {
    return nil, err
  }
  var subs []*Subscription

  for {
    sub := &Subscription{}
    found, err := scanQueriedRow(queriedRows,
      &sub.SubscriptionId,
      &sub.UserId,
      &sub.TickerId,
      &sub.Active,
      &sub.CreatedAt,
      &sub.ModifiedAt,
    )
    if err != nil {
      return nil, err
    }
    if !found {
      break
    }
    subs = append(subs, sub)
  }
  return subs, nil
}

func (s *storage) doRawQuery(query string, args ...any) (pgx.Rows, error) {
  rows, err := s.client.Query(s.ctx, query, args...)
  if err != nil {
    return nil, fmt.Errorf("cannot do query: %v", err)
  }
  return rows, err
}

func (s *storage) doExec(tx pgx.Tx, builder queryBuilder) error {
  query, args := mustBuildQuery(builder)
  var (
    execFunc postgres.ExecFunc
  )
  if tx != nil {
    execFunc = tx.Exec
  } else {
    execFunc = s.client.Exec
  }
  if _, err := execFunc(s.ctx, query, args...); err != nil {
    return fmt.Errorf("cannot do exec: %v", err)
  }
  return nil
}

func (s *storage) doQuery(tx pgx.Tx, builder queryBuilder) (pgx.Rows, error) {
  query, args := mustBuildQuery(builder)
  var (
    queryFunc postgres.QueryFunc
  )
  if tx != nil {
    queryFunc = tx.Query
  } else {
    queryFunc = s.client.Query
  }
  rows, err := queryFunc(s.ctx, query, args...)
  if err != nil {
    return nil, fmt.Errorf("cannot do query: %v", err)
  }
  return rows, err
}

func mustBuildQuery(builder queryBuilder) (string, []any) {
  query, args, err := builder.ToSql()
  if err != nil {
    log.Fatalf("build insert query failed: %v", err)
  }
  return query, args
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

func scanFirstQueriedRow(rows pgx.Rows, fields ...any) (bool, error) {
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

var ambiguousColumn = regexp.MustCompile(`[^.]ticker_id`)

func unambiguousGetTickersQuery(query string) string {
  const unambiguous = ` ticker.ticker_id `
  query = ambiguousColumn.ReplaceAllLiteralString(query, unambiguous)
  query = sanitizeQuery(query)
  return query
}

func (s *storage) GetOptionForTicker(tickerId string) *GetOption {
  const (
    tickerIdField = "ticker_id"
  )
  return &GetOption{
    Filters: FiltersOption{
      {
        Border: &BorderFilter{
          Field:   tickerIdField,
          Value:   tickerId,
          Compare: EqTokenizer{},
        },
      },
    },
  }
}
