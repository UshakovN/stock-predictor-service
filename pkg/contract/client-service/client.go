package clientservice

import (
  "context"
  "fmt"

  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  "github.com/UshakovN/stock-predictor-service/httpclient"
)

type Client interface {
  GetTickers(req *TickersRequest) (*TickersResponse, error)
}

type client struct {
  apiClient httpclient.HttpClient
}

func NewClient(ctx context.Context, apiPrefix, apiToken string) Client {
  return &client{
    apiClient: httpclient.NewClient(
      httpclient.WithContext(ctx),
      httpclient.WithApiPrefix(apiPrefix),
      httpclient.WithHeaderApiToken(authservice.ApiTokenHeader, apiToken),
    ),
  }
}

func (c *client) GetTickers(req *TickersRequest) (*TickersResponse, error) {
  const (
    getTickersRoute = "/tickers"
  )
  content, err := c.apiClient.Post(getTickersRoute, req, nil)
  if err != nil {
    return nil, fmt.Errorf("cannot do post request to '%s': %v", getTickersRoute, err)
  }
  resp := &TickersResponse{}

  if err = c.apiClient.ParseResponse(content, resp); err != nil {
    return nil, fmt.Errorf("cannot parse response from '%s': %v", getTickersRoute, err)
  }
  return resp, nil
}
