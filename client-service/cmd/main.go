package main

import (
  "context"
  "flag"
  "main/internal/handler"
  "os"
  "os/signal"
  "syscall"

  "github.com/UshakovN/stock-predictor-service/config"
  log "github.com/sirupsen/logrus"
)

// @title Client Service API
// @version 1.0.0
// @description API for client resource service
//
// @host localhost:8081
// @BasePath /
// @schemes http
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Auth-Token
//
func main() {
  ctx := context.Background()

  servePort := flag.String("port", "8081", "serving port")
  configPath := flag.String("path", "", "path to service config file")
  flag.Parse()

  cfg := handler.NewConfig()
  if err := config.Parse(*configPath, cfg); err != nil {
    log.Fatalf("cannot parse fetcher config: %v", err)
  }

  h, err := handler.NewHandler(ctx, cfg)
  if err != nil {
    log.Fatalf("cannot create new handler: %v", err)
  }
  h.BindRouter()

  go h.ContinuouslyServeHttp(*servePort)
  log.Infof("ready for serve http on port: %s", *servePort)

  serviceShutdown()
}

func serviceShutdown() {
  exitSignal := make(chan os.Signal)
  signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

  <-exitSignal
}
