package elasticsearch

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"

  es "github.com/elastic/go-elasticsearch/v7"
  "github.com/elastic/go-elasticsearch/v7/esapi"
  log "github.com/sirupsen/logrus"
)

const (
  prettyOption  = true
  refreshOption = "true"
  verboseResp   = true
)

type Client[Doc any] interface {
  CreateIndex(indexName string, jsonConfig []byte) error
  CreateDoc(indexName string, docId string, doc Doc) error
  Search(options *SearchOptions) (*SearchResults[Doc], error)
}

type client[Doc any] struct {
  ctx context.Context
  es  *es.Client
}

func NewClient[Doc any](ctx context.Context, config *Config) (Client[Doc], error) {
  esClient, err := es.NewClient(es.Config{
    Addresses: []string{
      config.Address,
    },
  })
  if err != nil {
    return nil, fmt.Errorf("create elasticsearch client failed: %v", err)
  }
  req := esapi.InfoRequest{
    Pretty: prettyOption,
  }
  resp, err := req.Do(ctx, esClient)
  if err != nil {
    return nil, fmt.Errorf("info request failed: %v", err)
  }
  if err = confirmResponse(resp, verboseResp); err != nil {
    return nil, err
  }
  return &client[Doc]{
    ctx: ctx,
    es:  esClient,
  }, nil
}

func (c *client[Doc]) CreateIndex(indexName string, jsonConfig []byte) error {
  req := esapi.IndicesCreateRequest{
    Index:  indexName,
    Body:   bytes.NewBuffer(jsonConfig),
    Pretty: prettyOption,
  }
  resp, err := req.Do(c.ctx, c.es)
  if err != nil {
    return fmt.Errorf("create index failed: %v", err)
  }
  if err = confirmResponse(resp, verboseResp); err != nil {
    return err
  }
  return nil
}

func (c *client[Doc]) CreateDoc(indexName, docId string, doc Doc) error {
  buf, err := json.Marshal(doc)
  if err != nil {
    return fmt.Errorf("marshal doc to json failed: %v", err)
  }
  req := esapi.IndexRequest{
    Index:      indexName,
    DocumentID: docId,
    Body:       bytes.NewBuffer(buf),
    Refresh:    refreshOption,
    Pretty:     prettyOption,
  }
  resp, err := req.Do(c.ctx, c.es)
  if err != nil {
    return fmt.Errorf("create doc failed: %v", err)
  }
  const (
    verboseResp = false
  )
  if err := confirmResponse(resp, verboseResp); err != nil {
    return err
  }
  return nil
}

type SearchOptions struct {
  Index     string
  Query     string
  Fields    []string
  Page      int
  Count     int
  Highlight bool
}

func (c *client[Doc]) Search(options *SearchOptions) (*SearchResults[Doc], error) {
  query := newSearchQuery(options)

  buf, err := json.Marshal(query)
  if err != nil {
    return nil, fmt.Errorf("marshal search query to json failed: %v", err)
  }
  if err != nil {
    return nil, fmt.Errorf("json buffer failed for search query: %v", err)
  }
  req := esapi.SearchRequest{
    Index: []string{
      options.Index,
    },
    Body:   bytes.NewBuffer(buf),
    Pretty: prettyOption,
  }
  esResp, err := req.Do(c.ctx, c.es)
  if err != nil {
    return nil, fmt.Errorf("search failed: %v", err)
  }
  resp, err := parseSearchResponse[Doc](esResp)
  if err != nil {
    return nil, fmt.Errorf("cannot parse search response: %v", err)
  }
  results := newSearchResults[Doc](resp)

  return results, nil
}

func confirmResponse(resp *esapi.Response, verbose bool) error {
  content := resp.String()
  if resp.IsError() {
    if verbose {
      log.Errorf("elasticsearch response error encountered")
      log.Errorf(content)
    }
    return fmt.Errorf("got error response")
  }
  if verbose {
    log.Infof("elasticsearch success response")
    log.Infof(content)
  }
  return nil
}
