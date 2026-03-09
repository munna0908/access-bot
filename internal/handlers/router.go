package handlers

import (
	"net/http"

	"github.com/access-bot/internal/logging"
	"github.com/access-bot/internal/services/accesslayer"
)

// Router sets up HTTP routing.
type Router struct {
	mux *http.ServeMux
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(service *accesslayer.Service, logger *logging.Logger) *Router {
	mux := http.NewServeMux()

	// Register handlers
	answerHandler := NewAnswerHandler(service, logger)
	healthHandler := NewHealthHandler()

	mux.Handle("POST /v1/answer", answerHandler)
	mux.Handle("GET /health", healthHandler)

	// Root path returns service info
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service": "access-layer", "version": "v1"}`))
	})

	return &Router{mux: mux}
}

// Handler returns the underlying HTTP handler.
func (r *Router) Handler() http.Handler {
	return r.mux
}
