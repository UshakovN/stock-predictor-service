package handler

import (
  "context"
  "fmt"
  _ "main/docs"
  "net/http"
  "regexp"

  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/swagger"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type Handler struct {
  ctx           context.Context
  servicesHosts map[string]string
  servicesNames []string
  swagger       *swagger.Handler
}

var regScheme = regexp.MustCompile(`https?://`)

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  servicesCount := len(config.SwaggerServices)

  servicesHosts := make(map[string]string, servicesCount)
  servicesNames := make([]string, 0, servicesCount)

  for _, service := range config.SwaggerServices {
    if schemeMatch := regScheme.MatchString(service.Prefix); !schemeMatch {
      return nil, fmt.Errorf("one of config swagger service has malformed prefix")
    }
    servicesHosts[service.Name] = service.Prefix
    servicesNames = append(servicesNames, service.Name)
  }

  return &Handler{
    ctx:           ctx,
    servicesHosts: servicesHosts,
    servicesNames: servicesNames,
    swagger:       swagger.NewHandler(config.SwaggerConfig),
  }, nil
}

func (h *Handler) BindRouter() {
  http.Handle("/redirect", errs.MiddlewareErr(h.swagger.BasicAuthMiddleware(h.HandleRedirect())))
  http.Handle("/services", errs.MiddlewareErr(h.swagger.BasicAuthMiddleware(h.HandleServices())))
  http.Handle("/health", h.HandleHealthWrapper())
  http.Handle("/swagger/", errs.MiddlewareErr(h.swagger.HandleSwagger()))
}

// HandleServices
//
// @Summary Services method for service names
// @Description Service method provide names of services support swagger
// @Tags Documentation
// @Produce            application/json
// @Success 200 {object} ServicesResponse
// @Failure 400,401,403,500 {object} errs.Error
// @Security HttpBasicAuth
// @Router /services [get]
//
func (h *Handler) HandleServices() errs.HandlerErr {
  resp := &ServicesResponse{
    Success:      true,
    ServiceNames: h.servicesNames,
  }
  return func(w http.ResponseWriter, r *http.Request) error {
    return utils.WriteResponse(w, resp, http.StatusOK)
  }
}

// HandleRedirect
//
// @Summary Redirect to swagger doc method
// @Description Redirect method made redirect to swagger documentation url for service
// @Tags Documentation
// @Produce            application/json
// @Param service query string true "Service Name"
// @Success 302
// @Failure 400,401,403,500 {object} errs.Error
// @Security HttpBasicAuth
// @Router /redirect [get]
//
func (h *Handler) HandleRedirect() errs.HandlerErr {
  const queryServiceName = "service"

  return func(w http.ResponseWriter, r *http.Request) error {
    if r.Method != http.MethodGet {
      return errs.NewError(errs.ErrTypeMethodNotSupported, nil)
    }
    serviceName := r.URL.Query().Get(queryServiceName)
    if serviceName == "" {
      msg := fmt.Sprintf("%s must be specified", queryServiceName)
      return errs.NewErrorWithMessage(errs.ErrTypeMalformedRequest, msg, nil)
    }
    servicePrefix, ok := h.servicesHosts[serviceName]
    if !ok {
      msg := fmt.Sprintf("%s service not found", serviceName)
      return errs.NewErrorWithMessage(errs.ErrTypeNotFoundContent, msg, nil)
    }
    swaggerRedirect := formSwaggerRedirect(servicePrefix)

    http.Redirect(w, r, swaggerRedirect, http.StatusFound)
    return nil
  }
}

func formSwaggerRedirect(servicePrefix string) string {
  return fmt.Sprintf("%s/swagger", servicePrefix)
}

// HandleHealthWrapper
//
// @Summary Health check method
// @Description Health method check http server health
// @Tags Health
// @Produce application/json
// @Success 200 {object} common.HealthResponse
// @Router /health [get]
//
func (h *Handler) HandleHealthWrapper() http.HandlerFunc {
  return utils.HandleHealth()
}

func (h *Handler) ContinuouslyServeHttp(port string) {
  err := http.ListenAndServe(fmt.Sprint(":", port), nil)
  if err != nil {
    log.Fatalf("listen and serve error: %v", err)
  }
}

type ServicesResponse struct {
  Success      bool     `json:"success"`
  ServiceNames []string `json:"service_names"`
}
