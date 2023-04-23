package utils

import (
  "context"
  "errors"
  "fmt"
)

var (
  ErrCtxValueNotFound = errors.New("value not found")
  ErrCtxTypeAssertion = errors.New("type assertion failed")
)

type CtxKey interface {
  KeyDescription() string
}

func GetCtxValue[T any](ctx context.Context, key CtxKey) (T, error) {
  var defaultVal T
  ctxValue := ctx.Value(key)
  if ctxValue == nil {
    return defaultVal, fmt.Errorf("%w: key '%s' in request context",
      ErrCtxValueNotFound, key.KeyDescription())
  }
  value, ok := ctxValue.(T)
  if !ok {
    return defaultVal, fmt.Errorf("%w: expected type '%T', actual value: '%v'",
      ErrCtxTypeAssertion, defaultVal, ctxValue)
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
