package httpclient

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "io"
  "main/pkg/utils"
  "net/http"
  "net/url"
  "strings"
  "time"

  limiter "github.com/UshakovN/token-bucket"
  log "github.com/sirupsen/logrus"
)

type Client struct {
  ctx     context.Context
  client  http.Client
  limiter *rateLimiter
  token   *apiToken
}

type apiToken struct {
  key   string
  value string
}

type rateLimiter struct {
  limiter     *limiter.TokenBucket
  reqsCount   int
  perDur      time.Duration
  waitDur     time.Duration
  deadlineDur time.Duration
}

type Options func(c *Client)

func NewClient(options ...Options) *Client {
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
    c.token = &apiToken{
      key:   tokenKey,
      value: tokenValue,
    }
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

func (c *Client) GetFullResp(requestURL string, headers ...Header) (*FullResp, error) {
  var (
    resp *http.Response
    err  error
  )

  err = utils.DoWithRetry(func() error {
    resp, err = c.getOnce(requestURL, headers)
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

func (c *Client) Get(requestURL string, headers ...Header) ([]byte, error) {
  fullResp, err := c.GetFullResp(requestURL, headers...)
  if err != nil {
    return nil, err
  }
  return fullResp.Content, nil
}

func (c *Client) getOnce(requestURL string, headers []Header) (*http.Response, error) {
  if err := c.limiter.Wait(c.ctx); err != nil {
    return nil, fmt.Errorf("limiter wait failed: %v", err)
  }

  req, err := c.formRequest(requestURL, nil)
  if err != nil {
    return nil, err
  }
  resp, err := c.doRequest(req, headers)
  if err != nil {
    return nil, err
  }

  return resp, nil
}

func (t *apiToken) addToRequestURL(reqURL string) (string, error) {
  reqURL = strings.TrimSpace(reqURL)
  // try parse url
  parsedUrl, err := url.Parse(reqURL)
  if err != nil {
    return "", fmt.Errorf("malformed request url '%s': %v", reqURL, err)
  }
  // parse query
  query := parsedUrl.Query()
  // set token value
  query.Set(t.key, t.value)
  parsedUrl.RawQuery = query.Encode()

  reqURL = parsedUrl.String()

  return reqURL, nil
}

func (c *Client) formRequest(reqURL string, body io.Reader) (*http.Request, error) {
  var (
    err error
  )
  if c.token != nil {
    reqURL, err = c.token.addToRequestURL(reqURL)
    if err != nil {
      return nil, fmt.Errorf("cannot add token to request url: %v", err)
    }
  }
  req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, reqURL, body)
  if err != nil {
    return nil, NewError(reqURL, fmt.Errorf("cannot create request: %v", err))
  }
  return req, nil
}

func (c *Client) doRequest(req *http.Request, headers []Header) (*http.Response, error) {
  req.Header = utils.ExtractOptional(headers...).toHttpHeaders()

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
  switch t := payload.(type) {
  case string:
    reader = strings.NewReader(payload.(string))
  case []byte:
    reader = bytes.NewBuffer(payload.([]byte))
  default:
    return nil, NewError(requestURL, fmt.Errorf("unsupported payload type: %T", t))
  }
  return reader, nil
}

func (c *Client) Post(requestURL string, payload any, headers ...Header) ([]byte, error) {
  var (
    resp *http.Response
    err  error
  )

  err = utils.DoWithRetry(func() error {
    resp, err = c.postOnce(requestURL, payload, headers)
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

func (c *Client) postOnce(requestURL string, payload any, headers []Header) (*http.Response, error) {
  if err := c.limiter.Wait(c.ctx); err != nil {
    return nil, fmt.Errorf("limiter wait failed: %v", err)
  }

  body, err := preparePostPayload(requestURL, payload)
  if err != nil {
    return nil, NewError(requestURL, fmt.Errorf("cannot prepare post payload: %v", err))
  }
  req, err := c.formRequest(requestURL, body)
  if err != nil {
    return nil, err
  }
  resp, err := c.doRequest(req, headers)
  if err != nil {
    return nil, err
  }

  return resp, err
}

func (c *Client) ParseResponse(bytes []byte, resp any) error {
  return json.Unmarshal(bytes, resp)
}

func (l *rateLimiter) Wait(ctx context.Context) error {
  if l.limiter == nil {
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
