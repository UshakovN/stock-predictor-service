package handler

import (
	"fmt"
	"time"

	"github.com/UshakovN/stock-predictor-service/errs"
)

type SignUpRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	Success      bool   `json:"status"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SingInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CheckUserRequest struct{}

type CheckUserResponse struct {
	Success   bool      `json:"success"`
	UserId    string    `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (r *SignUpRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("sign_up request is a nil")
	}
	err := errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
		" field must be specified", nil)

	if r.Email == "" {
		err.Message = fmt.Sprint("email", err.Message)
		return err
	}
	if r.FullName == "" {
		err.Message = fmt.Sprint("full_name", err.Message)
		return err
	}
	if r.Password == "" {
		err.Message = fmt.Sprint("password", err.Message)
		return err
	}
	return nil
}

func (r *SingInRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("sign_up request is a nil")
	}
	err := errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
		" field must be specified", nil)

	if r.Email == "" {
		err.Message = fmt.Sprint("email", err.Message)
		return err
	}
	if r.Password == "" {
		err.Message = fmt.Sprint("password", err.Message)
		return err
	}
	return nil
}

func (r *RefreshRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("sign_up request is a nil")
	}
	err := errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest,
		" field must be specified", nil)

	if r.RefreshToken == "" {
		err.Message = fmt.Sprint("refresh_token", err.Message)
		return err
	}
	return nil
}
