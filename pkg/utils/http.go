package utils

import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "net/url"

  "github.com/UshakovN/stock-predictor-service/contract/common"
  "github.com/UshakovN/stock-predictor-service/errs"
  log "github.com/sirupsen/logrus"
)

func HandleHealth() http.HandlerFunc {
  return func(w http.ResponseWriter, _ *http.Request) {
    if err := WriteResponse(w, &common.HealthResponse{
      Success: true,
    }, http.StatusOK); err != nil {
      log.Errorf("cannot write response: %v", err)
    }
  }
}

func ContinuouslyServe(port string) {
  err := http.ListenAndServe(fmt.Sprint(":", port), nil)
  if err != nil {
    log.Fatalf("listen and serve error: %v", err)
  }
}

func ReadRequest(r *http.Request, req any) error {
  switch r.Method {

  case http.MethodGet:
    query := plainQuery(r.URL.Query())
    b, err := json.Marshal(query)
    if err != nil {
      return fmt.Errorf("cannot marshal request query: %v", err)
    }
    if err = json.Unmarshal(b, req); err != nil {
      return fmt.Errorf("cannot unmarshal request query: %v", err)
    }
    return nil

  case http.MethodPost:
    content, err := io.ReadAll(r.Body)
    if err != nil {
      return fmt.Errorf("cannot read request body: %v", err)
    }
    defer func() {
      if err := r.Body.Close(); err != nil {
        log.Errorf("cannot close request body: %v", err)
      }
    }()
    if len(content) == 0 {
      return errs.NewError(errs.ErrTypeBodyNotFound, nil)
    }

    if err = json.Unmarshal(content, req); err != nil {
      return errs.NewError(errs.ErrTypeMalformedRequest, nil)
    }
    return nil

  default:
    return errs.NewError(errs.ErrTypeMethodNotSupported, nil)
  }
}

func plainQuery(query url.Values) map[string]string {
  plain := make(map[string]string, len(query))

  for key, values := range query {
    if len(values) == 0 {
      continue
    }
    plain[key] = values[0]
  }
  return plain
}

func WriteResponse(w http.ResponseWriter, resp any, statusCode int) error {
  respBytes, err := json.Marshal(resp)
  if err != nil {
    return fmt.Errorf("cannot marshal response: %v", err)
  }
  w.Header().Add("Content-Type", "application/json")
  w.WriteHeader(statusCode)
  if _, err = w.Write(respBytes); err != nil {
    return fmt.Errorf("cannot write response: %v", err)
  }
  return nil
}
