package utils

import (
  "context"
  "errors"
  "fmt"
  "net/http"
)

var (
  errValueNotFound = errors.New("value not found")
  errTypeAssertion = errors.New("type assertion failed")
)

type CtxKey interface {
  KeyDescription() string
}

func GetReqCtxValue[T any](r *http.Request, key CtxKey) (T, error) {
  var defaultVal T
  ctxValue := r.Context().Value(key)
  if ctxValue == nil {
    return defaultVal, fmt.Errorf("%w: key '%s' in request context",
      errValueNotFound, key.KeyDescription())
  }
  value, ok := ctxValue.(T)
  if !ok {
    return defaultVal, fmt.Errorf("%w: expected type '%T', actual value: '%v'",
      errTypeAssertion, defaultVal, ctxValue)
  }
  return value, nil
}

type CtxMap map[CtxKey]any

func AddCtxValues(ctx context.Context, m CtxMap) context.Context {
  childCtx := ctx
  parentCtx := ctx
  for key, val := range m {
    childCtx = context.WithValue(parentCtx, key, val)
    parentCtx = childCtx
  }
  return childCtx
}
