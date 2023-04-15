package handler

import (
	"fmt"

	"github.com/UshakovN/stock-predictor-service/errs"
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
	Name    string `json:"name"`
	Section string `json:"section"`
}

type GetResponse struct {
	SourceUrl string `json:"source_url"`
}

type GetBatchRequest struct {
	Parts []*GetRequest `json:"parts"`
}

func (r *PutRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("put request is a nil")
	}
	err := errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
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
	err := errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
		" field must be specified", nil)

	if r.Name == "" {
		err.Message = fmt.Sprint("name", err.Message)
		return err
	}
	if r.Section == "" {
		err.Message = fmt.Sprint("section", err.Message)
		return err
	}
	return nil
}

func (r *GetBatchRequest) Validate() error {
	err := errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
		" field must be specified", nil)

	if len(r.Parts) == 0 {
		err.Message = fmt.Sprint("parts", err.Message)
		return err
	}
	for _, part := range r.Parts {
		if err := part.Validate(); err != nil {
			return err
		}
	}
	return nil
}
