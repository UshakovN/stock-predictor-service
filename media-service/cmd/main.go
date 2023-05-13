package main

import (
  "context"
  "flag"
  "fmt"
  "main/internal/handler"
  "os"
  "os/signal"
  "syscall"

  "github.com/UshakovN/stock-predictor-service/config"
  log "github.com/sirupsen/logrus"
)

// @title Media Service API
// @version 1.0.0
// @description API for media content serving service
//
// @host localhost:8083
// @BasePath /
// @schemes http
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Auth-Token
//
func main() {
  ctx := context.Background()

  servePort := flag.String("port", "8083", "serving port")
  serveHost := flag.String("host", "", "host prefix for serve media content")
  configPath := flag.String("path", "", "path to service config file")
  flag.Parse()

  cfg := handler.NewConfig()
  if err := config.Parse(*configPath, cfg); err != nil {
    log.Fatalf("cannot parse fetcher config: %v", err)
  }
  hostPrefix := fmt.Sprintf("%s:%s", *serveHost, *servePort)

  h, err := handler.NewHandler(ctx, hostPrefix, cfg)
  if err != nil {
    log.Fatalf("cannot create new handler: %v", err)
  }
  h.BindRouter()

  go h.ContinuouslyServeHttp(*servePort)
  log.Infof("ready for serve http on port: %s", *servePort)

  go h.ContinuouslyServeQueue()
  log.Println("ready for serve message queue")

  serviceShutdown()
}

func serviceShutdown() {
  exitSignal := make(chan os.Signal)
  signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

  <-exitSignal
}
