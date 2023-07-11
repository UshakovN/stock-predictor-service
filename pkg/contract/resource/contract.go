package resource

type Request struct {
  Pagination *Pagination `json:"pagination,omitempty"`
  Sort       *Sort       `json:"sort,omitempty"`
  Filters    []*Filter   `json:"filters,omitempty"`
}

type Response struct {
  Success bool `json:"success"`
  Count   int  `json:"count"`
}

type Pagination struct {
  Page  int `json:"page"`
  Count int `json:"count"`
}

type Sort struct {
  Field string `json:"field"`
  Order string `json:"order"`
}

type Filter struct {
  Border  *BorderFilter  `json:"border,omitempty"`
  Between *BetweenFilter `json:"between,omitempty"`
  List    *ListFilter    `json:"list,omitempty"`
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
