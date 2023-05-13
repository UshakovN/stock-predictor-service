package handler

import (
  "context"
  "fmt"
  "net/http"
  "regexp"

  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/swagger"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type servicesHost map[string]string

type Handler struct {
  ctx      context.Context
  services servicesHost
  *swagger.Handler
}

var regScheme = regexp.MustCompile(`https?://`)

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  services := make(servicesHost, len(config.SwaggerServices))

  for _, service := range config.SwaggerServices {
    if schemeMatch := regScheme.MatchString(service.Prefix); !schemeMatch {
      return nil, fmt.Errorf("one of config swagger service has malformed prefix")
    }
    services[service.Name] = service.Prefix
  }

  return &Handler{
    ctx:      ctx,
    services: services,
    Handler:  swagger.NewHandler(config.SwaggerConfig),
  }, nil
}

func (h *Handler) BindRouter() {
  http.Handle("/swagger", errs.MiddlewareErr(h.HandleSwagger()))
  http.Handle("/health", utils.HandleHealth())
}

func (h *Handler) HandleSwagger() errs.HandlerErr {
  const query = "service"

  return h.BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) error {
    if r.Method != http.MethodGet {
      return errs.NewError(errs.ErrTypeMethodNotSupported, nil)
    }
    service := r.URL.Query().Get(query)
    if service == "" {
      msg := fmt.Sprintf("%s must be specified", query)
      return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest, msg, nil)
    }
    prefix, ok := h.services[service]
    if !ok {
      msg := fmt.Sprintf("%s service not found", service)
      return errs.NewErrorWithMessage(errs.ErrTypeNotFoundContent, msg, nil)
    }
    redirect := formSwaggerRedirect(prefix)

    http.Redirect(w, r, redirect, http.StatusFound)
    return nil
  })
}

func formSwaggerRedirect(prefix string) string {
  return fmt.Sprintf("%s/swagger", prefix)
}

func (h *Handler) ContinuouslyServeHttp(port string) {
  err := http.ListenAndServe(fmt.Sprint(":", port), nil)
  if err != nil {
    log.Fatalf("listen and serve error: %v", err)
  }
}
