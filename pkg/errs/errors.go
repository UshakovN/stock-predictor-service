package errs

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

const (
	LogLevelErr  = 0
	LogLevelWarn = 1
)

const (
	ErrTypeUndefined = iota
	ErrTypeMethodNotSupported
	ErrTypeBodyNotFound
	ErrTypeMalformedRequest
	ErrTypeInternalServerError
	ErrTypeNotFoundContent
	ErrTypeMalformedToken
	ErrTypeForbidden
	ErrTypeExpiredToken
)

const (
	messageMethodNotSupported  = "method not supported"
	messageInternalServerError = "internal server error"
	messageBodyNotFound        = "body not found"
	messageInvalidBody         = "malformed request"
	messageNotFoundContent     = "not found content"
	messageMalformedToken      = "malformed token"
	messageForbidden           = "forbidden"
	messageExpiredToken        = "expired token"
)

type Error struct {
	Err        error       `json:"-"`
	Type       int         `json:"-"`
	Message    string      `json:"message"`
	LogMessage *LogMessage `json:"-"`
}

type LogMessage struct {
	Level int
	Err   error
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
	case ErrTypeMalformedToken:
		err.Message = messageMalformedToken
	case ErrTypeForbidden:
		err.Message = messageForbidden
	case ErrTypeExpiredToken:
		err.Message = messageExpiredToken
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
			log.Warnln(e.LogMessage.Err.Error())
		case LogLevelErr:
			log.Errorln(e.LogMessage.Err.Error())
		}
	}
}
