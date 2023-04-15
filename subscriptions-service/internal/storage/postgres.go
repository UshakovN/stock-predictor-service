package storage

import (
	"context"

	"github.com/UshakovN/stock-predictor-service/postgres"
)

type Storage interface {
	GetTickers() error
	GetStocks() error
	GetSubscriptions() error
	PutSubscription() error
	DeleteSubscription() error
}

type storage struct {
	ctx    context.Context
	client *postgres.Client
}

func NewStorage(ctx context.Context, config *postgres.Config) (Storage, error) {
	//client, err := postgres.NewClient(ctx, config)
	//if err != nil {
	//  return nil, fmt.Errorf("cannot create new postgres client: %v", err)
	//}
	//return &storage{
	//  ctx:    ctx,
	//  client: client,
	//}, nil

	return nil, nil
}

func (c *storage) GetTickers() ([]*Ticker, error) {

	builder := postgres.NewSelectBuilder().
		From(`ticker`).
		Columns(
			``,
		)

	return nil, nil
}
