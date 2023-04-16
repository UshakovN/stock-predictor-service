package handler

import (
	"context"
	"net/http"
)

type Handler struct {
	ctx context.Context
	// TODO: service
}

// TODO: routes

func (h *Handler) BindRouter() {
	http.Handle("/tickers", nil)
	http.Handle("/stocks", nil)
	http.Handle("/subscriptions", nil)
	http.Handle("/subscribe", nil)
	http.Handle("/unsubscribe", nil)
}
