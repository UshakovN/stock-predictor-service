package main

import (
  "context"
  "flag"
  "main/internal/fetcher/polygon"
  "net/http"
  "os"
  "os/signal"
  "syscall"

  "github.com/UshakovN/stock-predictor-service/config"
  "github.com/UshakovN/stock-predictor-service/utils"
  log "github.com/sirupsen/logrus"
)

// @title Data Fetcher API
// @version 1.0.0
// @description API for stock market data fetcher
//
// @host localhost:8082
// @BasePath /
// @schemes http
//
func main() {
  ctx := context.Background()

  servePort := flag.String("port", "8082", "serving port")
  configPath := flag.String("path", "", "path to service config file")
  tickerId := flag.String("ticker", "", "specified ticker id for fetching")
  flag.Parse()

  cfg := polygon.NewConfig()
  if err := config.Parse(*configPath, cfg); err != nil {
    log.Fatalf("cannot parse fetcher config: %v", err)
  }

  fetcher, err := polygon.NewFetcher(ctx, cfg)
  if err != nil {
    log.Fatalf("cannot create new fetcher: %v", err)
  }
  if *tickerId != "" {
    fetcher.SetTickerId(*tickerId)
  }

  http.Handle("/health", HandleHealthWrapper())

  go utils.ContinuouslyServe(*servePort)
  log.Infof("ready for serve http on port: %s", *servePort)

  go fetcher.ContinuouslyFetch()
  defer fetcher.SaveFetcherState()

  serviceShutdown()
}

func serviceShutdown() {
  exitSignal := make(chan os.Signal)
  signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

  <-exitSignal
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
func HandleHealthWrapper() http.HandlerFunc {
  return utils.HandleHealth()
}
