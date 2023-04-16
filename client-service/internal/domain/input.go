package domain

import (
	"main/internal/storage"
)

type GetInput struct {
	Pagination    *Pagination    `json:"pagination"`
	Sort          *Sort          `json:"sort"`
	Filter        *Filter        `json:"filter"`
	BetweenFilter *BetweenFilter `json:"between_filter"`
}

type Pagination struct {
	Page  int `json:"page"`
	Count int `json:"count"`
}

func (p *Pagination) Option() *storage.PaginationOption {
	return &storage.PaginationOption{
		Page:  p.Page,
		Count: p.Count,
	}
}

type Sort struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

func (s *Sort) Option() *storage.SortOption {
	return &storage.SortOption{
		Field: s.Field,
		Order: s.Order,
	}
}

type Filter struct {
	Field   string `json:"field"`
	Value   any    `json:"value"`
	Compare string `json:"compare"`
}

func (f *Filter) Option() *storage.BorderFilter {
	var compare storage.CompareTokenizer

	switch f.Compare {
	case "eq":
		compare = storage.EqTokenizer{}
	case "gt":
		compare = storage.GtTokenizer{}
	case "gte":
		compare = storage.GteTokenizer{}
	case "lt":
		compare = storage.LtTokenizer{}
	case "lte":
		compare = storage.LteTokenizer{}
	}

	return &storage.BorderFilter{
		Field:   f.Field,
		Value:   f.Value,
		Compare: compare,
	}
}

type BetweenFilter struct {
	Field       string `json:"field"`
	LeftBorder  any    `json:"left_border"`
	RightBorder any    `json:"right_border"`
}

func (f *BetweenFilter) Option() *storage.BetweenFilter {
	return &storage.BetweenFilter{
		Field:       f.Field,
		LeftBorder:  f.LeftBorder,
		RightBorder: f.RightBorder,
	}
}

func (g *GetInput) ParseOption() *storage.GetOption {
	return &storage.GetOption{
		Pagination: g.Pagination.Option(),
		Sort:       g.Sort.Option(),
		Filter: &storage.FilterOption{
			Border:  g.Filter.Option(),
			Between: g.BetweenFilter.Option(),
		},
	}
}
