package httpclient

import "fmt"

const errorPattern = "URL: %s. message: %v"

type Error struct {
  URL string
  Err error
}

func NewError(URL string, err error) *Error {
  return &Error{
    URL: URL,
    Err: err,
  }
}

func (e *Error) Error() string {
  if e == nil {
    return ""
  }
  return fmt.Sprintf(errorPattern, e.URL, e.Err)
}
