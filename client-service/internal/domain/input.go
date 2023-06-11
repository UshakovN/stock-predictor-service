package domain

import (
  "main/internal/storage"
)

type GetInput struct {
  Pagination *PaginationInput `json:"pagination"`
  Sort       *SortInput       `json:"sort"`
  Filters    []*FilterInput   `json:"filters"`
  With       *WithFields      `json:"with"`
}

type PaginationInput struct {
  Page  int `json:"page"`
  Count int `json:"count"`
}

type SortInput struct {
  Field string `json:"field"`
  Order string `json:"order"`
}

type FilterInput struct {
  Border  *BorderFilter  `json:"border"`
  Between *BetweenFilter `json:"between"`
  List    *ListFilter    `json:"list"`
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

func (p *PaginationInput) Option() *storage.PaginationOption {
  if p == nil {
    return nil
  }
  return &storage.PaginationOption{
    Page:  p.Page,
    Count: p.Count,
  }
}

func (s *SortInput) Option() *storage.SortOption {
  if s == nil {
    return nil
  }
  return &storage.SortOption{
    Field: s.Field,
    Order: s.Order,
  }
}

func (f *BorderFilter) Option() *storage.BorderFilter {
  if f == nil {
    return nil
  }
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

func (f *BetweenFilter) Option() *storage.BetweenFilter {
  if f == nil {
    return nil
  }
  return &storage.BetweenFilter{
    Field:       f.Field,
    LeftBorder:  f.LeftBorder,
    RightBorder: f.RightBorder,
  }
}

func (f *ListFilter) Option() *storage.ListFilter {
  if f == nil {
    return nil
  }
  return &storage.ListFilter{
    Field:  f.Field,
    Values: f.Values,
  }
}

func (g *GetInput) ParseOption() *storage.GetOption {
  filters := make([]*storage.FilterPart, 0, len(g.Filters))

  for _, filter := range g.Filters {
    filters = append(filters, &storage.FilterPart{
      Border:  filter.Border.Option(),
      Between: filter.Between.Option(),
      List:    filter.List.Option(),
    })
  }
  return &storage.GetOption{
    Pagination: g.Pagination.Option(),
    Sort:       g.Sort.Option(),
    Filters:    filters,
  }
}

func (w *WithFields) HasMedia() bool {
  if w == nil {
    return false
  }
  return w.Media
}

type CalculatePagesInput struct {
  Resource string
  PageSize int
}
