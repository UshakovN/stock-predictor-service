package main

import (
  "context"
  "flag"
  "fmt"
  "main/internal/handler"
  "os"
  "os/signal"
  "syscall"

  log "github.com/sirupsen/logrus"
)

func main() {
  ctx := context.Background()

  serveHost := flag.String("host", "localhost", "host prefix for serve media content")
  servePort := flag.String("port", "8080", "serving port")
  configPath := flag.String("path", "", "path to service config file")
  flag.Parse()

  cfg := handler.NewConfig()
  if err := cfg.Parse(*configPath); err != nil {
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
