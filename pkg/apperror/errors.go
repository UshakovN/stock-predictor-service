package apperror

import (
  "encoding/json"

  log "github.com/sirupsen/logrus"
)

const (
  LogLevelErr  = 0
  LogLevelWarn = 1
)

const (
  ErrTypeUndefined           = 0
  ErrTypeMethodNotSupported  = 1
  ErrTypeBodyNotFound        = 2
  ErrTypeMalformedRequest    = 3
  ErrTypeInternalServerError = 4
  ErrTypeNotFoundContent     = 5
)

const (
  messageMethodNotSupported  = "method not supported"
  messageInternalServerError = "internal server error"
  messageBodyNotFound        = "body not found"
  messageInvalidBody         = "malformed request"
  messageNotFoundContent     = "not found content"
)

type Error struct {
  Err        error       `json:"-"`
  Type       int         `json:"-"`
  Message    string      `json:"message"`
  LogMessage *LogMessage `json:"-"`
}

type LogMessage struct {
  Level   int
  Message string
}

func NewErrorWithMessage(errType int, message string, logMessage *LogMessage) *Error {
  return &Error{
    Type:       errType,
    Message:    message,
    LogMessage: logMessage,
  }
}

func NewError(errType int, logMessage *LogMessage) *Error {
  err := &Error{
    Type:       errType,
    LogMessage: logMessage,
  }
  switch errType {
  case ErrTypeMethodNotSupported:
    err.Message = messageMethodNotSupported
  case ErrTypeBodyNotFound:
    err.Message = messageBodyNotFound
  case ErrTypeMalformedRequest:
    err.Message = messageInvalidBody
  case ErrTypeNotFoundContent:
    err.Message = messageNotFoundContent

  default:
    err.Message = messageInternalServerError
  }
  return err
}

func (e *Error) Error() string {
  return e.Message
}

func (e *Error) Marshal() []byte {
  bytesErr, err := json.Marshal(e)
  if err != nil {
    return []byte(e.Error())
  }
  return bytesErr
}

func (e *Error) Log() {
  if e.LogMessage != nil {
    switch e.LogMessage.Level {
    case LogLevelWarn:
      log.Warnln(e.LogMessage.Message)
    case LogLevelErr:
      log.Errorf(e.LogMessage.Message)
    }
  }
}
