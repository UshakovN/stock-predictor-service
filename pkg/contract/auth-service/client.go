package authservice

import (
  "context"
  "errors"
  "fmt"
  "net/http"

  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/httpclient"
  "github.com/UshakovN/stock-predictor-service/utils"
)

const (
  ApiTokenHeader = "X-Service-Token"
  AuthHeader     = "X-Auth-Token"
)

var ErrAuthForbidden = errors.New("forbidden error")

type Client interface {
  AuthMiddleware(handler errs.HandlerErr) errs.HandlerErr
}

type client struct {
  apiClient httpclient.HttpClient
  apiToken  string
}

func NewClient(ctx context.Context, apiPrefix string, apiToken string) Client {
  return &client{
    apiClient: httpclient.NewClient(
      httpclient.WithContext(ctx),
      httpclient.WithApiPrefix(apiPrefix),
    ),
    apiToken: apiToken,
  }
}

func (c *client) checkUser(authToken string) (*CheckUserResponse, error) {
  const (
    checkRoute = "/check"
  )
  fullResp, err := c.apiClient.GetFullResp(checkRoute, &httpclient.RequestOptions{
    ApiAuth: &httpclient.RequestApiAuth{
      Key:   AuthHeader,
      Value: authToken,
      Type:  httpclient.AuthApiHeader,
    },
    RetryInternalOnly: true,
  })
  if err != nil {
    return nil, fmt.Errorf("cannot do get request to '%s'. api client error: %v", checkRoute, err)
  }
  if fullResp.Code == http.StatusUnauthorized || fullResp.Code == http.StatusForbidden {
    return nil, ErrAuthForbidden
  }
  resp := &CheckUserResponse{}

  if err = c.apiClient.ParseResponse(fullResp.Content, resp); err != nil {
    return nil, fmt.Errorf("cannot parse response from '%s'. error: %v", checkRoute, err)
  }
  return resp, nil
}

const ctxKeyDescUserId = "user_id"

type CtxKeyUserId struct{}

func (CtxKeyUserId) KeyDescription() string {
  return ctxKeyDescUserId
}

const ctxKeyDescServiceAccess = "service_access"

type CtxKeyServiceAccess struct{}

func (CtxKeyServiceAccess) KeyDescription() string {
  return ctxKeyDescServiceAccess
}

func (c *client) AuthMiddleware(handler errs.HandlerErr) errs.HandlerErr {
  return func(w http.ResponseWriter, r *http.Request) error {
    if apiToken := r.Header.Get(ApiTokenHeader); apiToken != "" {
      if apiToken != c.apiToken {
        return errs.NewError(errs.ErrTypeForbidden, nil)
      }
      ctx := utils.AddCtxValues(r.Context(), utils.CtxMap{
        CtxKeyServiceAccess{}: true,
      })
      return handler(w, r.WithContext(ctx))
    }
    authToken := r.Header.Get(AuthHeader)
    if authToken == "" {
      return errs.NewError(errs.ErrTypeNotFoundToken, nil)
    }
    resp, err := c.checkUser(authToken)
    if err != nil {
      if errs.ErrIs(err, ErrAuthForbidden) {
        return errs.NewError(errs.ErrTypeForbidden, nil)
      }
      return err
    }
    ctx := utils.AddCtxValues(r.Context(), utils.CtxMap{
      CtxKeyUserId{}: resp.UserId,
    })
    return handler(w, r.WithContext(ctx))
  }
}
