package handler

import (
	"fmt"
	"net/http"

	"insider-challenge/internal/service"
	"insider-challenge/pkg/config"

	httpSwagger "github.com/swaggo/http-swagger"
)

// Handler represents the http handler
type Handler struct {
	service *service.Service
	cfg     *config.Config
	mux     *http.ServeMux
}

// New creates a new handler instance
func New(service *service.Service, cfg *config.Config) *Handler {
	h := &Handler{
		service: service,
		cfg:     cfg,
		mux:     http.NewServeMux(),
	}

	// Swagger
	h.mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.ServerPort)),
	))

	h.mux.HandleFunc("/start", h.handleStart)
	h.mux.HandleFunc("/stop", h.handleStop)
	h.mux.HandleFunc("/sent", h.handleSent)

	return h
}

// Start the http server
func (h *Handler) Start(port string) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", port), h.mux)
}
