package storage

import (
  "errors"
  "fmt"
  "regexp"
  "strings"

  "github.com/UshakovN/stock-predictor-service/postgres"
)

var (
  ErrMalformedPagination  = errors.New("malformed pagination option")
  ErrMalformedSort        = errors.New("malformed sort option")
  ErrMalformedFilter      = errors.New("malformed filter option")
  ErrFiltersHasDuplicates = errors.New("filters has duplicates")
  ErrOnlyOneFilterType    = errors.New("only one filter type must be specified")
)

var (
  sanFieldRegex       = regexp.MustCompile(`[^a-zA-Z-_]`)
  sanQueryFirstRegex  = regexp.MustCompile(`\s`)
  sanQuerySecondRegex = regexp.MustCompile(` {2,}`)
)

const (
  SortOrderAsc  = "asc"
  SortOrderDesc = "desc"
)

type GetOption struct {
  Pagination *PaginationOption
  Sort       *SortOption
  Filters    FiltersOption
}

func (o *GetOption) HasPagination() bool {
  return o != nil && o.Pagination != nil
}

func (o *GetOption) HasSort() bool {
  return o != nil && o.Sort != nil
}

func (o *GetOption) HasFilters() bool {
  return o != nil && len(o.Filters) > 0
}

type QueryStuff interface {
  Stuff(query string) (string, error)
}

type PaginationOption struct {
  Page  int
  Count int
}

func (p *PaginationOption) Stuff(query string) (string, error) {
  if p.Page < 0 || p.Count < 0 {
    return "", ErrMalformedPagination
  }
  const pageShift = 1

  offset := (p.Page - pageShift) * p.Count
  limit := p.Count
  query = fmt.Sprintf(`%s offset %d limit %d`, sanitizeQuery(query), offset, limit)

  return query, nil
}

type SortOption struct {
  Field string
  Order string
}

func (s *SortOption) Stuff(query string) (string, error) {
  if s.Field == "" || (s.Order != SortOrderAsc && s.Order != SortOrderDesc) {
    return "", ErrMalformedSort
  }
  query = fmt.Sprintf(`%s order by %s %s`,
    sanitizeQuery(query),
    sanitizeField(s.Field),
    sanitizeField(s.Order))

  return query, nil
}

type FiltersOption []*FilterPart

func (f FiltersOption) Stuff(query string) (string, error) {
  var err error

  for _, filter := range f {
    if query, err = filter.stuff(query); err != nil {
      return "", err
    }
  }
  const (
    first       = 1
    other       = -1
    union       = "and"
    placeholder = "{where}"
    statement   = "where"
  )
  query = strings.Replace(query, placeholder, statement, first)
  query = strings.Replace(query, placeholder, union, other)

  return query, nil
}

type FilterPart struct {
  Border  *BorderFilter
  Between *BetweenFilter
}

func (f *FilterPart) stuff(query string) (string, error) {
  if f.Border != nil && f.Between != nil {
    return "", ErrOnlyOneFilterType
  }
  if f.Border != nil {
    return f.Border.stuff(query)
  }
  if f.Between != nil {
    return f.Between.stuff(query)
  }
  return query, nil
}

type BorderFilter struct {
  Field   string
  Value   any
  Compare CompareTokenizer
}

type BetweenFilter struct {
  Field       string
  LeftBorder  any
  RightBorder any
}

func (f *BorderFilter) stuff(query string) (string, error) {
  if f.Field == "" || f.Value == nil || f.Compare == nil {
    return "", ErrMalformedFilter
  }
  query = fmt.Sprintf("%s {where} %s %s %v",
    sanitizeQuery(query),
    sanitizeField(f.Field),
    f.Compare.Token(),
    quoteValue(f.Value))

  return query, nil
}

func (f *BetweenFilter) stuff(query string) (string, error) {
  if f.Field == "" || f.LeftBorder == nil || f.RightBorder == nil {
    return "", ErrMalformedFilter
  }
  query = fmt.Sprintf("%s {where} %s between %v and %v",
    sanitizeQuery(query),
    sanitizeField(f.Field),
    quoteValue(f.LeftBorder),
    quoteValue(f.RightBorder))

  return query, nil
}

func (o *GetOption) Stuff(query string) (string, error) {
  if o == nil {
    return query, nil
  }
  var err error

  if o.HasFilters() {
    if query, err = o.Filters.Stuff(query); err != nil {
      return "", err
    }
  }
  if o.HasSort() {
    if query, err = o.Sort.Stuff(query); err != nil {
      return "", err
    }
  }
  if o.HasPagination() {
    if query, err = o.Pagination.Stuff(query); err != nil {
      return "", err
    }
  }

  return query, nil
}

func sanitizeQuery(query string) string {
  const space = " "
  query = sanQueryFirstRegex.ReplaceAllLiteralString(query, space)
  query = sanQuerySecondRegex.ReplaceAllLiteralString(query, space)
  return query
}

func sanitizeField(field string) string {
  const blank = ""
  field = sanFieldRegex.ReplaceAllLiteralString(field, blank)
  field = strings.ToLower(field)
  return field
}

func quoteValue(value any) string {
  return postgres.QuoteArg(value)
}
