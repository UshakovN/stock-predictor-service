package postgres

import (
  "context"
  "fmt"
  "main/pkg/utils"
  "time"

  "github.com/jackc/pgconn"
  "github.com/jackc/pgx/v4"
  "github.com/jackc/pgx/v4/pgxpool"
)

const connTimeout = 5 * time.Second

type Client interface {
  Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
  Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
  QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewClient(ctx context.Context, config *Config) (Client, error) {
  strConn := config.ConnectString()
  var (
    pgxConn *pgxpool.Pool
    err     error
  )
  err = utils.DoWithRetry(func() error {
    ctx, cancel := context.WithTimeout(ctx, connTimeout)
    defer cancel()

    if pgxConn, err = pgxpool.Connect(ctx, strConn); err != nil {
      return fmt.Errorf("cannot connect to posgtres pgx driver: %v", err)
    }
    return nil
  })
  if err != nil {
    return nil, fmt.Errorf("connection to posgtres pgx driver failed: %v", err)
  }

  return pgxConn, nil
}
