package handler

import (
  "fmt"
  "main/pkg/apperror"
)

type PutRequest struct {
  Name          string `json:"name"`
  Section       string `json:"section"`
  Content       []byte `json:"content"`
  ContentType   string `json:"content_type"`
  ContentLength string `json:"content_length,omitempty"`
  Overwrite     bool   `json:"overwrite,omitempty"`
}

type PutResponse struct {
  Queued bool `json:"queued"`
}

type GetRequest struct {
  Name        string `json:"name"`
  Section     string `json:"section"`
  ContentType string `json:"content_type"`
}

type GetResponse struct {
  SourceUrl string `json:"source_url"`
}

func (r *PutRequest) Validate() error {
  if r == nil {
    return fmt.Errorf("put request is a nil")
  }

  err := apperror.NewErrorWithMessage(apperror.ErrTypeMalformedRequest,
    " field must be specified", nil)

  if len(r.Content) == 0 {
    err.Message = fmt.Sprint("content", err.Message)
    return err
  }
  if r.Name == "" {
    err.Message = fmt.Sprint("name", err.Message)
    return err
  }
  if r.Section == "" {
    err.Message = fmt.Sprint("section", err.Message)
    return err
  }
  if r.ContentType == "" {
    err.Message = fmt.Sprint("content_type", err.Message)
    return err
  }

  return nil
}

func (r *GetRequest) Validate() error {
  if r == nil {
    return fmt.Errorf("get request is a nil")
  }

  err := apperror.NewErrorWithMessage(apperror.ErrTypeMalformedRequest,
    " field must be specified", nil)

  if r.Name == "" {
    err.Message = fmt.Sprint("name", err.Message)
    return err
  }
  if r.Section == "" {
    err.Message = fmt.Sprint("section", err.Message)
    return err
  }
  if r.ContentType == "" {
    err.Message = fmt.Sprint("content_type", err.Message)
    return err
  }

  return nil
}
