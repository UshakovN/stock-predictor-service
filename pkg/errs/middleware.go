package errs

import (
  "errors"
  "net/http"

  log "github.com/sirupsen/logrus"
)

type HandlerErr func(w http.ResponseWriter, r *http.Request) error

func MiddlewareErr(handler HandlerErr) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    handlerErr := handler(w, r)
    if handlerErr == nil {
      return
    }
    var (
      appErr     *Error
      statusCode int
    )
    if errors.As(handlerErr, &appErr) {
      switch appErr.Type {
      // define another app errors this
      case
        ErrTypeMethodNotSupported:
        statusCode = http.StatusMethodNotAllowed
      case
        ErrTypeBodyNotFound, ErrTypeMalformedRequest:
        statusCode = http.StatusBadRequest
      case
        ErrTypeNotFoundContent:
        statusCode = http.StatusNotFound
      case
        ErrTypeNotFoundToken, ErrTypeNotFoundCredentials:
        statusCode = http.StatusUnauthorized
      case
        ErrTypeMalformedToken, ErrTypeExpiredToken, ErrTypeForbidden, ErrTypeWrongCredentials:
        statusCode = http.StatusForbidden
      default:
        statusCode = http.StatusInternalServerError
      }
      writeError(w, appErr, statusCode)
      return

    }
    // if error not as app error
    statusCode = http.StatusInternalServerError
    appErr = NewError(ErrTypeUndefined, &LogMessage{
      Level: LogLevelErr,
      Err:   handlerErr,
    })

    writeError(w, appErr, statusCode)
  }
}

func writeError(w http.ResponseWriter, err *Error, statusCode int) {
  err.Log()
  w.Header().Add("Content-Type", "application/json")
  w.WriteHeader(statusCode)
  if _, err := w.Write(err.Marshal()); err != nil {
    log.Errorf("response writer error: %v", err)
  }
}
