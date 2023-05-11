package searchservice

import (
  clientservice "github.com/UshakovN/stock-predictor-service/contract/client-service"
  "github.com/UshakovN/stock-predictor-service/errs"
)

type ResourceRequest struct {
  Query string `json:"query"`
  Page  int    `json:"page"`
  Count int    `json:"count"`
}

type ResourceResponse struct {
  Success bool `json:"success"`
  Count   int  `json:"count"`
  Total   int  `json:"total"`
}

type SuggestRequest struct {
  *ResourceRequest
}

type SuggestResponse struct {
  *ResourceResponse
  Parts []*Part `json:"parts"`
}

type Part struct {
  Score float64 `json:"score"`
  Info  *Info   `json:"info"`
}

type Info struct {
  TickerId           string `json:"ticker_id"`
  CompanyName        string `json:"company_name"`
  CompanyDescription string `json:"company_description"`
  HomepageUrl        string `json:"homepage_url"`
}

type SearchRequest struct {
  *ResourceRequest
  With *clientservice.WithFields `json:"with,omitempty"`
}

type SearchResponse struct {
  *ResourceResponse
  Parts []*clientservice.Ticker `json:"parts"`
}

func (r *ResourceRequest) Validate() error {
  if r.Query == "" {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "query must be specified", nil)
  }
  if r.Page < 0 {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "page must be positive", nil)
  }
  if r.Count < 0 {
    return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
      "count must be positive", nil)
  }
  return nil
}
