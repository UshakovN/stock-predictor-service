package handler

import (
  "context"
  "fmt"
  "net/http"
  "os"
  "time"

  authservice "github.com/UshakovN/stock-predictor-service/contract/auth-service"
  clientservice "github.com/UshakovN/stock-predictor-service/contract/client-service"
  "github.com/UshakovN/stock-predictor-service/contract/common"
  searchservice "github.com/UshakovN/stock-predictor-service/contract/search-service"
  es "github.com/UshakovN/stock-predictor-service/elasticsearch"
  "github.com/UshakovN/stock-predictor-service/errs"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

type Handler struct {
  ctx                   context.Context
  serviceClient         clientservice.Client
  elasticClient         es.Client[*searchservice.Info]
  authClient            authservice.Client
  elasticIndexName      string
  suggestUpdateDuration time.Duration
}

func NewHandler(ctx context.Context, config *Config) (*Handler, error) {
  if config.SuggestUpdateHours <= 0 {
    return nil, fmt.Errorf("suggest update hours must be positive integer")
  }
  suggestUpdateDuration := time.Duration(config.SuggestUpdateHours) * time.Hour

  serviceClient := clientservice.NewClient(ctx, config.ClientServicePrefix, config.ClientServiceApiToken)

  elasticClient, err := es.NewClient[*searchservice.Info](ctx, &es.Config{
    Address: config.ElasticSearchPrefix,
  })
  if err != nil {
    return nil, err
  }
  authClient := authservice.NewClient(ctx, config.AuthServicePrefix, config.SuggestApiToken)

  return &Handler{
    ctx:                   ctx,
    serviceClient:         serviceClient,
    elasticClient:         elasticClient,
    authClient:            authClient,
    elasticIndexName:      config.ElasticIndexName,
    suggestUpdateDuration: suggestUpdateDuration,
  }, nil
}

func (h *Handler) BindRouter() {
  http.Handle("/suggest", errs.MiddlewareErr(h.authClient.AuthMiddleware(h.HandleSuggest)))
  http.Handle("/search", errs.MiddlewareErr(h.authClient.AuthMiddleware(h.HandleSearch)))
  http.Handle("/health", errs.MiddlewareErr(h.HandleHealth))
}

// HandleSuggest
//
// @Summary Suggest method for tickers suggesting
// @Description Suggest method provide tickers short info by query equal part of ticker id, company name, homepage url
// @Tags Suggesting
// @Produce            application/json
// @Param request body searchservice.SuggestRequest true "Request"
// @Success 200 {object} searchservice.SuggestResponse
// @Failure 400, 401, 403, 500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /suggest [post]
//
func (h *Handler) HandleSuggest(w http.ResponseWriter, r *http.Request) error {
  suggestReq := &searchservice.SuggestRequest{}

  if err := utils.ReadRequest(r, suggestReq); err != nil {
    return err
  }
  if err := suggestReq.Validate(); err != nil {
    return err
  }
  searchOptions := h.formSearchOptions(suggestReq.ResourceRequest)

  searchResults, err := h.elasticClient.Search(searchOptions)
  if err != nil {
    return err
  }
  suggestResp := h.formSuggestResponse(searchResults)

  if err = utils.WriteResponse(w, suggestResp, http.StatusOK); err != nil {
    return err
  }
  return nil
}

// HandleSearch
//
// @Summary Search method for tickers searching
// @Description Suggest method provide tickers models equal to /tickers response from Client Service by query
// @Tags Searching
// @Produce            application/json
// @Param request body searchservice.SearchRequest true "Request"
// @Success 200 {object} searchservice.SearchResponse
// @Failure 400, 401, 403, 500 {object} errs.Error
// @Security ApiKeyAuth
// @Router /search [post]
//
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) error {
  searchReq := &searchservice.SearchRequest{}

  if err := utils.ReadRequest(r, searchReq); err != nil {
    return err
  }
  if err := searchReq.Validate(); err != nil {
    return err
  }
  searchOptions := h.formSearchOptions(searchReq.ResourceRequest)

  searchResults, err := h.elasticClient.Search(searchOptions)
  if err != nil {
    return err
  }
  if len(searchResults.Parts) == 0 {
    if err = utils.WriteResponse(w, &searchservice.SearchResponse{
      ResourceResponse: &searchservice.ResourceResponse{
        Success: true,
        Count:   0,
        Total:   searchResults.Total,
      },
      Parts: []*clientservice.Ticker{},
    }, http.StatusOK); err != nil {
      return err
    }
    return nil
  }
  tickersReq := h.formTickersRequest(searchReq, searchResults)

  tickersResp, err := h.serviceClient.GetTickers(tickersReq)
  if err != nil {
    return err
  }
  if err = utils.WriteResponse(w, &searchservice.SearchResponse{
    ResourceResponse: &searchservice.ResourceResponse{
      Success: true,
      Count:   tickersResp.Count,
      Total:   searchResults.Total,
    },
    Parts: tickersResp.Tickers,
  }, http.StatusOK); err != nil {
    return err
  }

  return nil
}

// HandleHealth
//
// @Summary Health check method
// @Description Health method check http server health
// @Tags Health
// @Produce application/json
// @Success 200 {object} common.HealthResponse
// @Success 500 {object} errs.Error
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

func (h *Handler) ContinuouslyServeHttp(port string) {
  err := http.ListenAndServe(fmt.Sprint(":", port), nil)
  if err != nil {
    log.Fatalf("listen and serve error: %v", err)
  }
}

func (h *Handler) formTickersRequest(
  req *searchservice.SearchRequest,
  results *es.SearchResults[*searchservice.Info],
) *clientservice.TickersRequest {
  const (
    fieldTickerId = "ticker_id"
  )
  values := make([]any, 0, len(results.Parts))

  for _, part := range results.Parts {
    if part.DocId == "" {
      continue
    }
    values = append(values, part.DocId)
  }
  return &clientservice.TickersRequest{
    ResourceRequest: &clientservice.ResourceRequest{
      Filters: []*clientservice.Filter{
        {
          List: &clientservice.ListFilter{
            Field:  fieldTickerId,
            Values: values,
          },
        },
      },
    },
    With: req.With,
  }
}

func (h *Handler) formSuggestResponse(results *es.SearchResults[*searchservice.Info]) *searchservice.SuggestResponse {
  parts := make([]*searchservice.Part, 0, len(results.Parts))

  for _, part := range results.Parts {
    parts = append(parts, &searchservice.Part{
      Score: part.Score,
      Info:  part.Doc,
    })
  }
  return &searchservice.SuggestResponse{
    ResourceResponse: &searchservice.ResourceResponse{
      Success: true,
      Count:   len(parts),
      Total:   results.Total,
    },
    Parts: parts,
  }
}

func (h *Handler) formSearchOptions(req *searchservice.ResourceRequest) *es.SearchOptions {
  const (
    fieldTickerId    = "ticker_id"
    fieldCompanyName = "company_name"
    fieldHomepageUrl = "homepage_url"
  )
  req.Query = utils.StripString(req.Query)

  return &es.SearchOptions{
    Index: h.elasticIndexName,
    Query: req.Query,
    Fields: []string{
      fieldTickerId,
      fieldCompanyName,
      fieldHomepageUrl,
    },
    Page:  req.Page,
    Count: req.Count,
  }
}

func (h *Handler) UpdateSuggestScheduled() {
  const (
    startTimer = 0
    timeFormat = "2006-01-02 15:04:05"
  )
  timer := time.NewTimer(startTimer)
  defer timer.Stop()

  for timerTime := range timer.C {
    log.Infof("start scheduled update suggest: %s", timerTime.UTC().Format(timeFormat))
    if err := h.updateSuggest(); err != nil {
      log.Errorf("suggest update failed: %v", err)
      return
    }
    log.Infof("scheduled suggest update success: %s", utils.NotTimeUTC().Format(timeFormat))
    timer.Reset(h.suggestUpdateDuration)
  }
}

func (h *Handler) updateSuggest() error {
  resp, err := h.serviceClient.GetTickers(&clientservice.TickersRequest{})
  if err != nil {
    return fmt.Errorf("client service cannot get tickers: %v", err)
  }
  log.Infof("client service return %d tickers for suggest", resp.Count)

  var createdDocs int
  const createdCountForLog = 25

  for _, ticker := range resp.Tickers {
    if ticker.Fields == nil || ticker.Fields.TickerId == "" {
      log.Errorf("encountered ticker with nil fields")
      continue
    }
    partInfo := searchservice.Info{
      TickerId:           ticker.Fields.TickerId,
      CompanyName:        ticker.Fields.CompanyName,
      CompanyDescription: ticker.Fields.CompanyDescription,
      HomepageUrl:        ticker.Fields.HomepageUrl,
    }
    if err = h.elasticClient.CreateDoc(h.elasticIndexName, partInfo.TickerId, &partInfo); err != nil {
      return fmt.Errorf("es client cannot create doc in index %s: %v", h.elasticIndexName, err)
    }
    createdDocs++

    if createdDocs%createdCountForLog == 0 {
      log.Infof("es client create %d docs in index %s", createdDocs, h.elasticIndexName)
    }
  }
  return nil
}

func (h *Handler) createElasticIndexForSuggest(jsonConfigPath string) error {
  if jsonConfigPath == "" {
    return fmt.Errorf("json config path not specified")
  }
  jsonConfig, err := os.ReadFile(jsonConfigPath)
  if err != nil {
    return fmt.Errorf("cannot read json config from path '%s': %v", jsonConfigPath, err)
  }
  if err = h.elasticClient.CreateIndex(h.elasticIndexName, jsonConfig); err != nil {
    return fmt.Errorf("es client cannot create index: %v", err)
  }
  return nil
}
