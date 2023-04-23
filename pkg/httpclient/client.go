package httpclient

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "strings"
  "time"

  "github.com/UshakovN/stock-predictor-service/utils"
  limiter "github.com/UshakovN/token-bucket"
  log "github.com/sirupsen/logrus"
)

type ApiAuthType int

const (
  AuthApiToken  ApiAuthType = 0
  AuthApiHeader ApiAuthType = 1
)

type RequestApiAuth struct {
  Key   string
  Value string
  Type  ApiAuthType
}

type RequestOptions struct {
  Headers Header
  ApiAuth *RequestApiAuth
}

func (o *RequestOptions) stuffRequest(req *http.Request) error {
  if auth := o.ApiAuth; auth != nil {
    token := &apiAuth{
      key:   auth.Key,
      value: auth.Value,
    }
    if auth.Type != AuthApiToken && auth.Type != AuthApiHeader {
      return fmt.Errorf("unrecognized auth type")
    }
    if auth.Type == AuthApiToken {
      token.query = true
    } else
    if auth.Type == AuthApiHeader {
      token.header = true
    }
    token.stuffRequest(req)
  }
  if headers := o.Headers; headers != nil {
    req.Header = headers.toHttpHeaders()
  }
  return nil
}

type HttpClient interface {
  GetFullResp(requestURL string, options *RequestOptions) (*FullResp, error)
  Get(requestURL string, options *RequestOptions) ([]byte, error)
  Post(requestURL string, payload any, options *RequestOptions) ([]byte, error)
  ParseResponse(bytes []byte, resp any) error
}

type Client struct {
  ctx     context.Context
  client  http.Client
  limiter *rateLimiter
  token   *apiAuth
  prefix  string
}

type apiAuth struct {
  key    string
  value  string
  query  bool
  header bool
}

type rateLimiter struct {
  limiter     *limiter.TokenBucket
  reqsCount   int
  perDur      time.Duration
  waitDur     time.Duration
  deadlineDur time.Duration
}

type Options func(c *Client)

func NewClient(options ...Options) HttpClient {
  client := &Client{
    ctx:    context.Background(),
    client: http.Client{},
  }
  for _, opt := range options {
    opt(client)
  }
  return client
}

func WithContext(ctx context.Context) Options {
  return func(c *Client) {
    if ctx == nil {
      return
    }
    c.ctx = ctx
  }
}

func WithApiToken(tokenKey, tokenValue string) Options {
  return func(c *Client) {
    c.token = &apiAuth{
      key:   tokenKey,
      value: tokenValue,
      query: true,
    }
  }
}

func WithAuthHeader(authHeader, accessToken string) Options {
  return func(c *Client) {
    c.token = &apiAuth{
      key:    authHeader,
      value:  accessToken,
      header: true,
    }
  }
}

func WithApiPrefix(prefixURL string) Options {
  return func(c *Client) {
    c.prefix = prefixURL
  }
}

func WithRequestsLimit(reqsCount int, perDur, waitDur, deadlineDur time.Duration) Options {
  return func(c *Client) {
    if reqsCount <= 0 {
      return
    }
    c.limiter = &rateLimiter{
      limiter: limiter.NewTokenBucket(
        reqsCount,
        reqsCount,
        limiter.SetRefillDuration(perDur),
      ),
      reqsCount:   reqsCount,
      perDur:      perDur,
      waitDur:     waitDur,
      deadlineDur: deadlineDur,
    }
  }
}

type Header map[string]string

func (h Header) GetOrDefault(key string) string {
  return h[key]
}

func (h Header) Get(key string) (string, bool) {
  val, ok := h[key]
  return val, ok
}

func (h Header) toHttpHeaders() http.Header {
  httpHeaders := http.Header{}

  for key, value := range h {
    if key == "" {
      continue
    }
    httpHeaders.Add(key, value)
  }

  return httpHeaders
}

func toHeaders(httpHeaders http.Header) Header {
  headers := Header{}

  for key, values := range httpHeaders {
    if len(values) == 0 {
      continue
    }
    headers[key] = values[0]
  }

  return headers
}

type FullResp struct {
  Content []byte
  Headers Header
  Code    int
}

func (c *Client) GetFullResp(requestURL string, options *RequestOptions) (*FullResp, error) {
  var (
    resp *http.Response
    err  error
  )

  err = utils.DoWithRetry(func() error {
    resp, err = c.getOnce(requestURL, options)
    return err
  })
  if err != nil {
    return nil, NewError(requestURL,
      fmt.Errorf("%s request failed: %v", http.MethodGet, err),
    )
  }

  content, err := readResponse(requestURL, resp)
  if err != nil {
    return nil, err
  }
  respHeaders := toHeaders(resp.Header)
  statusCode := resp.StatusCode

  return &FullResp{
    Content: content,
    Headers: respHeaders,
    Code:    statusCode,
  }, nil
}

func (c *Client) Get(requestURL string, options *RequestOptions) ([]byte, error) {
  fullResp, err := c.GetFullResp(requestURL, options)
  if err != nil {
    return nil, err
  }
  return fullResp.Content, nil
}

func (c *Client) getOnce(requestURL string, options *RequestOptions) (*http.Response, error) {
  if err := c.limiter.Wait(c.ctx); err != nil {
    return nil, fmt.Errorf("limiter wait failed: %v", err)
  }

  req, err := c.formRequest(http.MethodGet, requestURL, nil, options)
  if err != nil {
    return nil, err
  }
  resp, err := c.doRequest(req)
  if err != nil {
    return nil, err
  }

  return resp, nil
}

func (a *apiAuth) stuffRequest(req *http.Request) {
  if a.query {
    a.stuffReqURL(req)
  } else
  if a.header {
    a.stuffReqHeader(req)
  }
}

func (a *apiAuth) stuffReqHeader(req *http.Request) {
  if req != nil {
    req.Header.Set(a.key, a.value)
  }
}

func (a *apiAuth) stuffReqURL(req *http.Request) {
  if req != nil {
    query := req.URL.Query()
    query.Set(a.key, a.value)
    req.URL.RawQuery = query.Encode()
  }
}

func (c *Client) formRequest(method, reqURL string, body io.Reader, options *RequestOptions) (*http.Request, error) {
  if c.prefix != "" {
    reqURL = fmt.Sprint(c.prefix, reqURL)
  }

  req, err := http.NewRequestWithContext(c.ctx, method, reqURL, body)
  if err != nil {
    return nil, NewError(reqURL, fmt.Errorf("cannot create request: %v", err))
  }

  if c.token != nil {
    c.token.stuffRequest(req)
  }
  if options != nil {
    if err := options.stuffRequest(req); err != nil {
      return nil, fmt.Errorf("cannot stuff request options: %v", err)
    }
  }

  return req, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
  resp, err := c.client.Do(req)
  if err != nil {
    return nil, NewError(req.URL.String(), fmt.Errorf("do request failed: %v", err))
  }
  statusCode := resp.StatusCode

  if statusCode >= http.StatusBadRequest {
    return nil, NewError(req.URL.String(), fmt.Errorf("bad response. got status code: %d", statusCode))
  }
  return resp, nil
}

func readResponse(requestURL string, resp *http.Response) ([]byte, error) {
  content, err := io.ReadAll(resp.Body)
  if err != nil {
    return nil, NewError(requestURL, fmt.Errorf("cannot read response: %v", err))
  }

  if err = resp.Body.Close(); err != nil {
    return nil, NewError(requestURL, fmt.Errorf("cannot close response reader: %v", err))
  }

  return content, nil
}

func preparePostPayload(requestURL string, payload any) (io.Reader, error) {
  var reader io.Reader
  switch p := payload.(type) {
  case string:
    reader = strings.NewReader(p)
  case []byte:
    reader = bytes.NewBuffer(p)
  default:
    buf, err := json.Marshal(p)
    if err != nil {
      return nil, NewError(requestURL, fmt.Errorf("cannot marshal payload to json: %v", err))
    }
    reader = bytes.NewBuffer(buf)
  }
  return reader, nil
}

func (c *Client) Post(requestURL string, payload any, options *RequestOptions) ([]byte, error) {
  var (
    resp *http.Response
    err  error
  )

  err = utils.DoWithRetry(func() error {
    resp, err = c.postOnce(requestURL, payload, options)
    return err
  })
  if err != nil {
    return nil, NewError(requestURL, fmt.Errorf("post request failed: %v", err))
  }

  content, err := readResponse(requestURL, resp)
  if err != nil {
    return nil, err
  }

  return content, nil
}

func (c *Client) postOnce(requestURL string, payload any, options *RequestOptions) (*http.Response, error) {
  if err := c.limiter.Wait(c.ctx); err != nil {
    return nil, fmt.Errorf("limiter wait failed: %v", err)
  }

  body, err := preparePostPayload(requestURL, payload)
  if err != nil {
    return nil, NewError(requestURL, fmt.Errorf("cannot prepare post payload: %v", err))
  }
  req, err := c.formRequest(http.MethodPost, requestURL, body, options)
  if err != nil {
    return nil, err
  }
  resp, err := c.doRequest(req)
  if err != nil {
    return nil, err
  }

  return resp, err
}

func (c *Client) ParseResponse(bytes []byte, resp any) error {
  return json.Unmarshal(bytes, resp)
}

func (l *rateLimiter) Wait(ctx context.Context) error {
  if l == nil || l.limiter == nil {
    return nil
  }
  deadlineTime := time.Now().Add(l.deadlineDur)

  ctx, cancel := context.WithDeadline(ctx, deadlineTime)
  defer cancel()

  for {
    select {
    case <-ctx.Done():
      return fmt.Errorf("limiter deadline %s exceeded", l.deadlineDur)

    default:
      if l.limiter.Allow() {
        return nil
      }
      untilDeadlineDur := deadlineTime.Sub(time.Now()).Round(time.Second)

      log.Infof("limiter: sent %d requests in %s. limit reached. sleep on %s. until waiting deadline: %s",
        l.reqsCount, l.perDur, l.waitDur, untilDeadlineDur)

      time.Sleep(l.waitDur)
    }
  }
}
