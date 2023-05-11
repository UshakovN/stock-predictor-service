package media_service

import (
  "context"
  "fmt"
  "time"

  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  "github.com/UshakovN/stock-predictor-service/httpclient"
)

type Client interface {
  GetBatch(req *GetBatchRequest) (*GetBatchResponse, error)
}

type client struct {
  apiClient httpclient.HttpClient
}

func NewClient(ctx context.Context, apiPrefix, apiToken string) Client {
  const (
    retryCount = 3
    retryWait  = 1 * time.Second
  )
  apiClient := httpclient.NewClient(
    httpclient.WithContext(ctx),
    httpclient.WithApiPrefix(apiPrefix),
    httpclient.WithHeaderApiToken(authservice.ApiTokenHeader, apiToken),
    httpclient.WithCustomRetries(retryCount, retryWait),
  )
  return &client{
    apiClient: apiClient,
  }
}

func (c *client) GetBatch(req *GetBatchRequest) (*GetBatchResponse, error) {
  const (
    getBatchRoute = "/get-batch"
  )
  content, err := c.apiClient.Post(getBatchRoute, req, &httpclient.RequestOptions{
    RetryInternalOnly: true,
  })
  if err != nil {
    return nil, fmt.Errorf("cannot do post request to '%s'. api client error: %v", getBatchRoute, err)
  }
  resp := &GetBatchResponse{}

  if err = c.apiClient.ParseResponse(content, resp); err != nil {
    return nil, fmt.Errorf("cannot parse response from '%s'. error: %v", getBatchRoute, err)
  }
  return resp, nil
}
