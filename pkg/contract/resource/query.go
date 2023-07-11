package resource

import (
  "fmt"
  "net/url"
  "strconv"
  "strings"
)

const (
  keyPagination = "pagination"
  keyFilters    = "filter"
  keySort       = "sort"

  filterSuffixBorder  = "border"
  filterSuffixBetween = "between"
  filterSuffixList    = "list"

  keySuffix = "%s[%s]"
)

var (
  keyFilterBorder  = fmt.Sprintf(keySuffix, keyFilters, filterSuffixBorder)
  keyFilterBetween = fmt.Sprintf(keySuffix, keyFilters, filterSuffixBetween)
  keyFilterList    = fmt.Sprintf(keySuffix, keyFilters, filterSuffixList)
)

var keysQuery = []string{
  keyFilterBorder,
  keyFilterBetween,
  keyFilterList,
  keyPagination,
  keySort,
}

type Query struct {
  parsed url.Values
}

func NewQuery(options ...Option) (*Query, error) {
  q := &Query{}
  var err error

  for _, option := range options {
    if err = option(q); err != nil {
      return nil, err
    }
  }
  return q, nil
}

type Option func(*Query) error

func WithEncodedQuery(query string) Option {
  return func(q *Query) error {
    parsed, err := url.ParseQuery(query)
    if err != nil {
      return fmt.Errorf("cannot parse query: %v", err)
    }
    q.parsed = parsed
    return nil
  }
}

func WithParsedQuery(parsed url.Values) Option {
  return func(q *Query) error {
    q.parsed = parsed
    return nil
  }
}

func (q *Query) ParseRequest() (*Request, error) {
  if q.parsed == nil {
    return nil, fmt.Errorf("parsed query not specified")
  }
  req := &Request{}
  var err error

  if req.Pagination, err = q.parsePagination(); err != nil {
    return nil, err
  }
  if req.Sort, err = q.parseSort(); err != nil {
    return nil, err
  }
  if req.Filters, err = q.parseFilters(); err != nil {
    return nil, err
  }
  return req, nil
}

func (q *Query) SanitizeQuery() url.Values {
  parsed := q.parsed

  for _, keyQuery := range keysQuery {
    parsed.Del(keyQuery)
  }
  return parsed
}

func (q *Query) parsePagination() (*Pagination, error) {
  var parsedPagination *Pagination

  const (
    sortPartIdx    = 0
    sortPartsCount = 1
  )
  const partsCount = 2

  if paginationParts := q.parsed[keyPagination]; len(paginationParts) == sortPartsCount {
    paginationPart := paginationParts[sortPartIdx]

    parts, err := splitPartForm(paginationPart, partsCount)
    if err != nil {
      return nil, fmt.Errorf("pagination must have form [page:<p>,count:<c>]. error: %v", err)
    }
    const (
      keyPage  = "page"
      keyCount = "count"
    )
    var (
      partPage  string
      partCount string
    )
    for _, part := range parts {
      if partPage == "" {
        if partPage, err = parseFormPartVal(part, keyPage); err != nil {
          return nil, err
        }
      }
      if partCount == "" {
        if partCount, err = parseFormPartVal(part, keyCount); err != nil {
          return nil, err
        }
      }
    }
    castedPage, err := castPaginationValue(partPage)
    if err != nil {
      return nil, fmt.Errorf("pagination has malformed page: %v", err)
    }
    castedCount, err := castPaginationValue(partCount)
    if err != nil {
      return nil, fmt.Errorf("pagination has malformed count: %v", err)
    }
    parsedPagination = &Pagination{
      Page:  castedPage,
      Count: castedCount,
    }
  } else if len(paginationParts) > sortPartsCount {
    return nil, fmt.Errorf("multiple specified pagination")
  }
  return parsedPagination, nil
}

func (q *Query) parseSort() (*Sort, error) {
  var parsedSort *Sort

  const (
    sortPartIdx    = 0
    sortPartsCount = 1
  )
  const partsCount = 2

  if sortParts := q.parsed[keySort]; len(sortParts) == sortPartsCount {
    sortPart := sortParts[sortPartIdx]

    parts, err := splitPartForm(sortPart, partsCount)
    if err != nil {
      return nil, fmt.Errorf("sort must have form [field:<f>,order:<asc|desc>]. error: %v", err)
    }
    const (
      keyField = "field"
      keyOrder = "order"
    )
    var (
      partField string
      partOrder string
    )
    for _, part := range parts {
      if partField == "" {
        if partField, err = parseFormPartVal(part, keyField); err != nil {
          return nil, err
        }
      }
      if partOrder == "" {
        if partOrder, err = parseFormPartVal(part, keyOrder); err != nil {
          return nil, err
        }
      }
    }
    if partField == "" || hasMalformedSortOrder(partOrder) {
      return nil, fmt.Errorf("sort must have form [field:<f>,order:<asc|desc>]. error: one of parts not specified or malformed")
    }
    parsedSort = &Sort{
      Field: partField,
      Order: partOrder,
    }
  } else if len(sortParts) > sortPartsCount {
    return nil, fmt.Errorf("multiple sort not supported")
  }
  return parsedSort, nil
}

func (q *Query) parseFilters() ([]*Filter, error) {
  var (
    parsed  []*Filter
    filters []*Filter
    err     error
  )
  for _, parse := range []func() ([]*Filter, error){
    q.parseBorderFilters,
    q.parseBetweenFilters,
    q.parseListFilters,
  } {
    if filters, err = parse(); err != nil {
      return nil, err
    }
    if len(filters) != 0 {
      parsed = append(parsed, filters...)
    }
  }
  return parsed, nil
}

func (q *Query) parseListFilters() ([]*Filter, error) {
  const (
    partValSep = "|"
    partsCount = 2
  )
  const (
    keyField  = "field"
    keyValues = "values"
  )
  list := q.parsed[keyFilterList]
  listFilters := make([]*Filter, 0, len(list))

  for _, list := range list {
    parts, err := splitPartForm(list, partsCount)
    if err != nil {
      return nil, fmt.Errorf("list filter must have form [field:<f>,values:<v|...>]. error: %v", err)
    }
    var (
      partField  string
      partValues []string
    )
    for _, part := range parts {
      if partField == "" {
        if partField, err = parseFormPartVal(part, keyField); err != nil {
          return nil, err
        }
      }
      if len(partValues) == 0 {
        var merged string

        if merged, err = parseFormPartVal(part, keyValues); err != nil {
          return nil, err
        }
        if merged != "" {
          partValues = strings.Split(merged, partValSep)
        }
      }
    }
    if partField == "" || len(partValues) == 0 {
      return nil, fmt.Errorf("list filter must have form [field:<f>,values:<v|...>]. error: one of parts not specified or malformed")
    }
    castedValues := make([]any, 0, len(partValues))

    for _, partValue := range partValues {
      castedValues = append(castedValues, castFieldValue(partValue))
    }
    listFilters = append(listFilters, &Filter{
      List: &ListFilter{
        Field:  partField,
        Values: castedValues,
      },
    })
  }
  return listFilters, nil
}

func (q *Query) parseBetweenFilters() ([]*Filter, error) {
  const partsCount = 2

  const (
    keyField = "field"
    keyLeft  = "left"
    keyRight = "right"
  )
  between := q.parsed[keyFilterBetween]
  betweenFilters := make([]*Filter, 0, len(between))

  for _, between := range between {
    parts, err := splitPartForm(between, partsCount)
    if err != nil {
      return nil, fmt.Errorf("between filter must have form [field:<f>,left:<l>,right:<r>]. error: %v", err)
    }
    var (
      partField string
      partLeft  string
      partRight string
    )
    for _, part := range parts {
      if partField == "" {
        if partField, err = parseFormPartVal(part, keyField); err != nil {
          return nil, err
        }
      }
      if partLeft == "" {
        if partLeft, err = parseFormPartVal(part, keyLeft); err != nil {
          return nil, err
        }
      }
      if partRight == "" {
        if partRight, err = parseFormPartVal(part, keyRight); err != nil {
          return nil, err
        }
      }
    }
    if partField == "" || partLeft == "" || partRight == "" {
      return nil, fmt.Errorf("between filter must have form [field:<f>,left:<l>,right:<r>]. error: one of parts not specified or malformed")
    }
    betweenFilters = append(betweenFilters, &Filter{
      Between: &BetweenFilter{
        Field:       partField,
        LeftBorder:  castFieldValue(partLeft),
        RightBorder: castFieldValue(partRight),
      },
    })
  }
  return betweenFilters, nil
}

func (q *Query) parseBorderFilters() ([]*Filter, error) {
  const partsCount = 3

  const (
    keyField   = "field"
    keyValue   = "value"
    keyCompare = "compare"
  )
  borders := q.parsed[keyFilterBorder]
  borderFilters := make([]*Filter, 0, len(borders))

  for _, border := range borders {
    parts, err := splitPartForm(border, partsCount)
    if err != nil {
      return nil, fmt.Errorf("border filter must have form [field:<f>,value:<v>,compare:<eq|gt|gte|lt|lte>]. error: %v", err)
    }
    var (
      partField   string
      partValue   string
      partCompare string
    )
    for _, part := range parts {
      if partField == "" {
        if partField, err = parseFormPartVal(part, keyField); err != nil {
          return nil, err
        }
      }
      if partValue == "" {
        if partValue, err = parseFormPartVal(part, keyValue); err != nil {
          return nil, err
        }
      }
      if partCompare == "" {
        if partCompare, err = parseFormPartVal(part, keyCompare); err != nil {
          return nil, err
        }
      }
    }
    if partField == "" || partValue == "" || hasMalformedFilterCompare(partCompare) {
      return nil, fmt.Errorf("border filter must have form [field:<f>,value:<v>,compare:<eq|gt|gte|lt|lte>]. error: one of parts not specified or malformed")
    }
    borderFilters = append(borderFilters, &Filter{
      Border: &BorderFilter{
        Field:   partField,
        Value:   castFieldValue(partValue),
        Compare: partCompare,
      },
    })
  }
  return borderFilters, nil
}

func splitPartForm(s string, partsCount int) ([]string, error) {
  const partsSep = ","

  if hasMalformedBrackets(s) {
    return nil, fmt.Errorf("has malformed brackets")
  }
  s = stripFormBrackets(s)

  parts := strings.Split(s, partsSep)

  if len(parts) < partsCount {
    return nil, fmt.Errorf("has not enought parts")
  }
  return parts, nil
}

func castPaginationValue(s string) (int, error) {
  val, err := strconv.Atoi(s)
  if err != nil {
    return 0, fmt.Errorf("cannot cast %s to integer", s)
  }
  if val < 0 {
    return 0, fmt.Errorf("value must be positive integer")
  }
  return val, nil
}

func castFieldValue(val string) any {
  if val, err := strconv.Atoi(val); err == nil {
    return val
  }
  if val, err := strconv.ParseFloat(val, 64); err == nil {
    return val
  }
  if val, err := strconv.ParseBool(val); err == nil {
    return val
  }
  return val
}

func hasMalformedBrackets(s string) bool {
  const (
    prefix = "["
    suffix = "]"
  )
  return !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, suffix)
}

func stripFormBrackets(s string) string {
  const (
    prefix = "["
    suffix = "]"
  )
  s = strings.TrimPrefix(s, prefix)
  s = strings.TrimSuffix(s, suffix)
  return s
}

func parseFormPartVal(part, key string) (string, error) {
  const (
    partsCount = 2
    partValIdx = 1
  )
  const partsValSep = ":"
  var (
    formPartVal string
    formParts   []string
  )
  if strings.Contains(part, key) {
    if formParts = strings.Split(part, partsValSep); len(formParts) != partsCount {
      return "", fmt.Errorf("malformed. expected: %s:<v>", key)
    }
    formPartVal = formParts[partValIdx]
  }
  return formPartVal, nil
}

func hasMalformedFilterCompare(s string) bool {
  _, ok := map[string]struct{}{
    "eq":  {},
    "gt":  {},
    "gte": {},
    "lt":  {},
    "lte": {},
  }[s]
  return !ok
}

func hasMalformedSortOrder(s string) bool {
  return s != "asc" && s != "desc"
}
