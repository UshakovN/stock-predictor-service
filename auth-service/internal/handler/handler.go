package handler

import (
  "context"
  "fmt"

  "main/internal/domain"
  "main/internal/service"
  "main/internal/storage"

  _ "main/docs"

  "net/http"
  "time"

  "github.com/UshakovN/stock-predictor-service/auth"
  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/hash"
  "github.com/UshakovN/stock-predictor-service/swagger"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type Handler struct {
  ctx     context.Context
  manager auth.TokenManager
  service service.UserAuthService
  swagger *swagger.Handler
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
    swagger: swagger.NewHandler(config.SwaggerConfig),
  }, nil
}

func (h *Handler) BindRouter() {
  http.Handle("/sign-up", errs.MiddlewareErr(h.HandleSignUp))
  http.Handle("/sign-in", errs.MiddlewareErr(h.HandleSignIn))
  http.Handle("/refresh", errs.MiddlewareErr(h.HandleRefresh))
  http.Handle("/check", errs.MiddlewareErr(h.MiddlewareAuth(h.HandleCheck)))
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))
  http.Handle("/swagger/", errs.MiddlewareErr(h.swagger.HandleSwagger()))
}

func (h *Handler) MiddlewareAuth(handler errs.HandlerErr) errs.HandlerErr {
  return func(w http.ResponseWriter, r *http.Request) error {
    authToken := r.Header.Get(authservice.AuthHeader)
    if authToken == "" {
      return errs.NewError(errs.ErrTypeNotFoundToken, nil)
    }
    userId, err := h.manager.Parse(authToken)
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

// HandleSignUp
//
// @Summary Sign Up method for service users
// @Description Sign Up method create service user model, put model in storage, create access and refresh tokens
// @Tags Authentication
// @Produce            application/json
// @Param request body authservice.SignUpRequest true "Request"
// @Success 200 {object} authservice.SignUpResponse
// @Failure 400,500 {object} errs.Error
// @Router /sign-up [post]
//
func (h *Handler) HandleSignUp(w http.ResponseWriter, r *http.Request) error {
  req := &authservice.SignUpRequest{}

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
  if err = utils.WriteResponse(w, &authservice.SignUpResponse{
    Success:      true,
    AccessToken:  tokens.Access,
    RefreshToken: tokens.Refresh,
  }, http.StatusCreated); err != nil {
    return err
  }

  return nil
}

// HandleSignIn
//
// @Summary Sign In method for service users
// @Description Sign In method search user model in storage, check provided credentials, generate access and refresh tokens
// @Tags Authentication
// @Produce            application/json
// @Param request body authservice.SignInRequest true "Request"
// @Success 200 {object} authservice.SignInResponse
// @Failure 400,500 {object} errs.Error
// @Router /sign-in [post]
//
func (h *Handler) HandleSignIn(w http.ResponseWriter, r *http.Request) error {
  req := &authservice.SignInRequest{}

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
  if err = utils.WriteResponse(w, &authservice.SignInResponse{
    Success:      true,
    AccessToken:  tokens.Access,
    RefreshToken: tokens.Refresh,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

// HandleRefresh
//
// @Summary Refresh tokens method for service users
// @Description Refresh method check provided refresh token and generate new access and refresh tokens
// @Tags Authentication
// @Produce            application/json
// @Param request body authservice.RefreshRequest true "Request"
// @Success 200 {object} authservice.RefreshResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Router /refresh [post]
//
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) error {
  req := &authservice.RefreshRequest{}

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
  if err = utils.WriteResponse(w, &authservice.RefreshResponse{
    Success:      true,
    AccessToken:  userTokens.Access,
    RefreshToken: userTokens.Refresh,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

// HandleCheck
//
// @Summary Check access token method for service users
// @Description Check method check user jwt access token from request header and collect user info
// @Tags Authorization
// @Produce application/json
// @Success 200 {object} authservice.CheckUserResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /check [get]
//
func (h *Handler) HandleCheck(w http.ResponseWriter, r *http.Request) error {
  req := &authservice.CheckUserRequest{}

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

  if err := utils.WriteResponse(w, &authservice.CheckUserResponse{
    Success:   true,
    UserId:    userInfo.UserId,
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

// HandleHealth
//
// @Summary Health check method
// @Description Health method check http server health
// @Tags Health
// @Produce application/json
// @Success 200 {object} common.HealthResponse
// @Router /health [get]
//
func (h *Handler) HandleHealth(w http.ResponseWriter, _ *http.Request) error {
  if err := utils.WriteResponse(w, &common.HealthResponse{
    Success: true,
  }, http.StatusOK); err != nil {
    return err
  }
  return nil
}
