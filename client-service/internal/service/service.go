package service

import (
	"context"
	"fmt"
	"main/internal/domain"
	"main/internal/storage"

	"github.com/UshakovN/stock-predictor-service/errs"
)

type Service interface {
	GetTickers(input *domain.GetInput) ([]*domain.Ticker, error)
	GetStocks(input *domain.GetInput) ([]*domain.Stock, error)
}

type service struct {
	ctx             context.Context
	exchangeStorage storage.ExchangeStorage
}

func NewService(ctx context.Context, exchangeStorage storage.ExchangeStorage) Service {
	return &service{
		ctx:             ctx,
		exchangeStorage: exchangeStorage,
	}
}

func (s *service) GetTickers(input *domain.GetInput) ([]*domain.Ticker, error) {
	stored, err := s.exchangeStorage.GetTickers(input.ParseOption())
	if err := formStorageError(err); err != nil {
		return nil, err
	}
	var (
		tickers []*domain.Ticker
	)
	for _, stored := range stored {
		tickers = append(tickers, formTicker(stored))
	}
	return tickers, nil
}

func (s *service) GetStocks(input *domain.GetInput) ([]*domain.Stock, error) {
	stored, err := s.exchangeStorage.GetStocks(input.ParseOption())
	if err := formStorageError(err); err != nil {
		return nil, err
	}
	var (
		stocks []*domain.Stock
	)
	for _, stored := range stored {
		stocks = append(stocks, formStock(stored))
	}
	return stocks, nil
}

func formTicker(stored *storage.Ticker) *domain.Ticker {
	return &domain.Ticker{
		Fields: &domain.TickerFields{
			TickerId:           stored.TickerId,
			CompanyName:        stored.CompanyName,
			CompanyLocale:      stored.CompanyLocale,
			CompanyDescription: stored.CompanyDescription,
			CompanyState:       stored.CompanyState,
			CompanyCity:        stored.CompanyCity,
			CompanyAddress:     stored.CompanyAddress,
			HomepageUrl:        stored.HomepageUrl,
			CurrencyName:       stored.CurrencyName,
			TotalEmployees:     stored.TotalEmployees,
			Active:             stored.Active,
			CreatedAt:          stored.CreatedAt,
		},
	}
}

func formStock(stored *storage.Stock) *domain.Stock {
	return &domain.Stock{
		StockId:       stored.StockId,
		TickerId:      stored.TickerId,
		OpenPrice:     stored.OpenPrice,
		ClosePrice:    stored.ClosePrice,
		HighestPrice:  stored.HighestPrice,
		LowestPrice:   stored.LowestPrice,
		TradingVolume: stored.TradingVolume,
		StockedAt:     stored.StockedAt,
		CreatedAt:     stored.CreatedAt,
	}
}

func formStorageError(err error) error {
	if errs.ErrIs(err,
		storage.ErrMalformedPagination,
		storage.ErrMalformedFilter,
		storage.ErrMalformedFilter,
	) {
		return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest, err.Error(), nil)
	}
	if err != nil {
		return fmt.Errorf("cannot get from storage: %v", err)
	}
	return nil
}
