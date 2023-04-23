package handler

import (
  "context"
  "fmt"

  "main/internal/domain"
  "main/internal/service"
  "main/internal/storage"

  "github.com/UshakovN/stock-predictor-service/auth"
  "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/hash"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"

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

  authService := service.NewUserAuthService(ctx, &service.Config{
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
  http.Handle("/sign-up", errs.MiddlewareErr(h.HandleSignUp))
  http.Handle("/sign-in", errs.MiddlewareErr(h.HandleSignIn))
  http.Handle("/refresh", errs.MiddlewareErr(h.HandleRefresh))
  http.Handle("/check", errs.MiddlewareErr(h.MiddlewareAuth(h.HandleCheck)))
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))
}

func (h *Handler) MiddlewareAuth(handler errs.HandlerErr) errs.HandlerErr {
  const (
    authHeader = "x-auth-token"
  )
  return func(w http.ResponseWriter, r *http.Request) error {
    accessToken := r.Header.Get(authHeader)

    userId, err := h.manager.Parse(accessToken)
    if err != nil {
      return errs.NewError(errs.ErrTypeMalformedToken, &errs.LogMessage{
        Err: err,
      })
    }
    ctx := utils.AddCtxValues(r.Context(), utils.CtxMap{
      ctxKeyUserId{}: userId,
    })

    return handler(w, r.WithContext(ctx))
  }
}

func (h *Handler) HandleSignUp(w http.ResponseWriter, r *http.Request) error {
  req := &auth_service.SignUpRequest{}

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
  if err = utils.WriteResponse(w, &auth_service.SignUpResponse{
    Success:      true,
    AccessToken:  tokens.Access,
    RefreshToken: tokens.Refresh,
  }, http.StatusCreated); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleSignIn(w http.ResponseWriter, r *http.Request) error {
  req := &auth_service.SingInRequest{}

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
  if err = utils.WriteResponse(w, &auth_service.SignInResponse{
    Success:      true,
    AccessToken:  tokens.Access,
    RefreshToken: tokens.Refresh,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) error {
  req := &auth_service.RefreshRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  if err := req.Validate(); err != nil {
    return err
  }
  userTokens, err := h.service.RefreshTokens(req.RefreshToken)
  if err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &auth_service.RefreshResponse{
    Success:      true,
    AccessToken:  userTokens.Access,
    RefreshToken: userTokens.Refresh,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func (h *Handler) HandleCheck(w http.ResponseWriter, r *http.Request) error {
  req := &auth_service.CheckUserRequest{}

  if err := utils.ReadRequest(r, req); err != nil {
    return err
  }
  userId, err := utils.GetCtxValue[string](r.Context(), ctxKeyUserId{})
  if err != nil {
    return errs.NewError(errs.ErrTypeForbidden, &errs.LogMessage{
      Err: err,
    })
  }
  userInfo, err := h.service.CheckUser(userId)
  if err != nil {
    return err
  }

  if err := utils.WriteResponse(w, &auth_service.CheckUserResponse{
    Success:   true,
    Email:     userInfo.Email,
    FullName:  userInfo.FullName,
    CreatedAt: userInfo.CreatedAt,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

func (h *Handler) ContinuouslyServeHttp(port string) {
  err := http.ListenAndServe(fmt.Sprint(":", port), nil)
  if err != nil {
    log.Fatalf("listen and serve error: %v", err)
  }
}

func (h *Handler) HandleHealth(w http.ResponseWriter, _ *http.Request) error {
  if err := utils.WriteResponse(w, &common.HealthResponse{
    Success: true,
  }, http.StatusOK); err != nil {
    return err
  }
  return nil
}
