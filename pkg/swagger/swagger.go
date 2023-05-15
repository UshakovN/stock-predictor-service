package swagger

import (
  "net/http"
  "strings"

  "github.com/UshakovN/stock-predictor-service/errs"
  httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Config struct {
  User     string `yaml:"user" required:"true"`
  Password string `yaml:"password" required:"true"`
}

type Handler struct {
  user     string
  password string
}

func NewHandler(config *Config) *Handler {
  return &Handler{
    user:     config.User,
    password: config.Password,
  }
}

func (h *Handler) BasicAuthMiddleware(handler errs.HandlerErr) errs.HandlerErr {
  const (
    headerKey   = "WWW-Authenticate"
    headerValue = "Basic realm=Swagger"
  )
  return func(w http.ResponseWriter, r *http.Request) error {
    user, password, ok := r.BasicAuth()
    if !ok {
      w.Header().Add(headerKey, headerValue)
      return errs.NewError(errs.ErrTypeNotFoundCredentials, nil)
    }
    if user != h.user || password != h.password {
      return errs.NewError(errs.ErrTypeWrongCredentials, nil)
    }
    return handler(w, r)
  }
}

func (h *Handler) HandleSwagger() errs.HandlerErr {
  return h.BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) error {
    setCorsHeaders(w)
    httpSwagger.Handler(httpSwagger.UIConfig(
      map[string]string{
        "defaultModelsExpandDepth": "-1",
      },
    ))(w, r)
    return nil
  })
}

func setCorsHeaders(w http.ResponseWriter) {
  corsHeaders := strings.Join([]string{
    "Content-Type",
    "Access-Control-Allow-Headers",
    "Authorization",
    "X-Requested-With",
  }, ", ")
  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,POST,GET")
  w.Header().Add("Access-Control-Allow-Headers", corsHeaders)
}
