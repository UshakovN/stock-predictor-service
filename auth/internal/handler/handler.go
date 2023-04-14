package handler

import (
  "context"
  "fmt"

  "main/internal/domain"
  "main/internal/service"
  "main/internal/storage"

  "github.com/UshakovN/stock-predictor-service"

  "net/http"
  "time"
)

type Handler struct {
  ctx     context.Context
  manager auth.TokenManager
  service service.UserAuthService
}

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  serviceStorage, err := storage.NewStorage(ctx, config.StorageConfig)
  if err != nil {
    return nil, fmt.Errorf("cannot create storage: %v", err)
  }

  tokenManager, err := auth.NewManager(config.TokenSignKey)
  if err != nil {
    return nil, fmt.Errorf("cannot create token manager: %v", err)
  }
  passwordManager := hash.NewManager(config.PasswordSalt)
  tokenTtl := time.Duration(config.TokenTtlMinutes) * time.Minute

  authService := service.NewService(ctx, &service.Config{
    TokenTtl:        tokenTtl,
    TokenManager:    tokenManager,
    PasswordManager: passwordManager,
    Storage:         serviceStorage,
  })

  return &Handler{
    ctx:     ctx,
    manager: tokenManager,
    service: authService,
  }, nil
}

func (h *Handler) BindRouter() {
  http.Handle("/sign-up", nil)

  // TODO: complete this

}

func (h *Handler) MiddlewareAuth(w http.ResponseWriter, r *http.Request) error {
  const (
    authTokenHeader = "x-auth-token"
  )
  authToken := r.Header.Get(authTokenHeader)

  claimUserId, err := h.manager.Parse(authToken)
  if err != nil {
    return apperror.NewError(apperror.ErrTypeMethodNotSupported)
  }

}

func (h *Handler) HandleSignUp(w http.ResponseWriter, r *http.Request) error {
  req := SignUpRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := req.Validate(); err != nil {
    return err
  }
  tokens, err := h.service.SignUp(&domain.SignUpInput{
    Email:    req.Email,
    FullName: req.FullName,
    Password: req.Password,
  })
  if err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &SignUpResponse{
    Success:      true,
    AccessToken:  tokens.Access,
    RefreshToken: tokens.Refresh,
  }, http.StatusCreated); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleSignIn(w http.ResponseWriter, r *http.Request) error {
  req := SingInRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := req.Validate(); err != nil {
    return err
  }
  tokens, err := h.service.SignIn(&domain.SignInInput{
    Email:    req.Email,
    Password: req.Password,
  })
  if err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &SignInResponse{
    Success:      true,
    AccessToken:  tokens.Access,
    RefreshToken: tokens.Refresh,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func HandleCheckUser(w http.ResponseWriter, r *http.Request) error {
  req := CheckUserRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }

  return nil
}

// TODO: complete this
//
// handle routes
//
// POST: /sign-up
// GET:  /sign-in
// GET:  /check-user
// GET:  /deactivate-user
// POST: /refresh-token
