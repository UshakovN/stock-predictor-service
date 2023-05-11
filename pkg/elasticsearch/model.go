package elasticsearch

import (
  "encoding/json"
  "fmt"
  "io"

  "github.com/elastic/go-elasticsearch/v7/esapi"
  log "github.com/sirupsen/logrus"
)

type searchQuery struct {
  From      int                   `json:"from,omitempty"`
  Size      int                   `json:"size,omitempty"`
  Query     searchQueryMatch      `json:"query"`
  Highlight *searchQueryHighlight `json:"highlight,omitempty"`
}

type searchQueryMultiMatch struct {
  Query  string   `json:"query"`
  Fields []string `json:"fields"`
}

type searchQueryMatch struct {
  MultiMatch searchQueryMultiMatch `json:"multi_match"`
}

type searchQueryHighlight struct {
  Fields map[string]*highlightField `json:"fields"`
}

type highlightField struct {
  PreTags      string `json:"pre_tags"`
  PostTags     string `json:"post_tags"`
  FragmentSize int    `json:"fragment_size"`
  Type         string `json:"type"`
}

func newSearchQuery(option *SearchOptions) *searchQuery {
  query := &searchQuery{
    Query: searchQueryMatch{
      MultiMatch: searchQueryMultiMatch{
        Query:  option.Query,
        Fields: option.Fields,
      },
    },
  }
  if option.Highlight {
    query.Highlight = newSearchQueryHighlight(option.Fields)
  }
  if option.Count > 0 {
    query.Size = option.Count
  }
  if option.Page > 0 {
    const (
      pageShift = 1
    )
    query.From = (option.Page - pageShift) * option.Count
  }
  return query
}

func newSearchQueryHighlight(fields []string) *searchQueryHighlight {
  highlightFields := make(map[string]*highlightField, len(fields))
  highlightField := newHighlightField()

  for _, field := range fields {
    highlightFields[field] = highlightField
  }
  return &searchQueryHighlight{
    Fields: highlightFields,
  }
}

func newHighlightField() *highlightField {
  const (
    noTags      = ""
    noFragments = 1
    noFeatures  = "plain"
  )
  return &highlightField{
    PreTags:      noTags,
    PostTags:     noTags,
    FragmentSize: noFragments,
    Type:         noFeatures,
  }
}

type searchResp[Doc any] struct {
  Hits struct {
    Total struct {
      Value int `json:"value"`
    } `json:"total"`
    Hits []struct {
      Index     string              `json:"_index"`
      Id        string              `json:"_id"`
      Score     float64             `json:"_score"`
      Source    Doc                 `json:"_source"`
      Highlight map[string][]string `json:"highlight"`
    } `json:"hits"`
  } `json:"hits"`
}

type SearchResult[Doc any] struct {
  Score      float64
  DocId      string
  Doc        Doc
  Highlights map[string][]string
}

type SearchResults[Doc any] struct {
  Total int
  Parts []*SearchResult[Doc]
}

func newSearchResults[Doc any](resp *searchResp[Doc]) *SearchResults[Doc] {
  results := &SearchResults[Doc]{}
  results.Total = resp.Hits.Total.Value

  for _, hit := range resp.Hits.Hits {
    results.Parts = append(results.Parts, &SearchResult[Doc]{
      Score:      hit.Score,
      DocId:      hit.Id,
      Doc:        hit.Source,
      Highlights: hit.Highlight,
    })
  }
  return results
}

func parseSearchResponse[Doc any](esResp *esapi.Response) (*searchResp[Doc], error) {
  if esResp.IsError() {
    return nil, fmt.Errorf("search error encountered")
  }
  buf, err := io.ReadAll(esResp.Body)
  if err != nil {
    return nil, fmt.Errorf("cannot read search response: %v", err)
  }
  defer func() {
    if err = esResp.Body.Close(); err != nil {
      log.Errorf("cannot close search response body: %v", err)
    }
  }()
  resp := &searchResp[Doc]{}
  if err := json.Unmarshal(buf, resp); err != nil {
    return nil, fmt.Errorf("cannot unmarshal json search response: %v", err)
  }
  return resp, nil
}
