package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/access-bot/internal/domain"
	"github.com/access-bot/internal/logging"
	"github.com/access-bot/internal/models"
	"github.com/access-bot/internal/services/accesslayer"
)

// AnswerHandler handles POST /v1/answer requests.
type AnswerHandler struct {
	service *accesslayer.Service
	logger  *logging.Logger
}

// NewAnswerHandler creates a new answer handler.
func NewAnswerHandler(service *accesslayer.Service, logger *logging.Logger) *AnswerHandler {
	return &AnswerHandler{
		service: service,
		logger:  logger,
	}
}

// ServeHTTP handles the answer request.
func (h *AnswerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only accept POST
	if r.Method != http.MethodPost {
		h.writeError(w, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var req models.AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "request_decode_failed", "error", err.Error())
		response := models.NewFailureResponse("", string(domain.StatusInvalidRequest), string(domain.ReasonValidationFailed))
		h.writeJSON(w, response, http.StatusBadRequest)
		return
	}

	// Process request through the service
	response := h.service.ProcessAnswerRequest(ctx, &req)

	// Determine HTTP status code based on response status
	statusCode := h.getHTTPStatus(response.Status)
	h.writeJSON(w, response, statusCode)
}

// getHTTPStatus maps response status to HTTP status code.
func (h *AnswerHandler) getHTTPStatus(status string) int {
	switch status {
	case string(domain.StatusOK):
		return http.StatusOK
	case string(domain.StatusInvalidRequest):
		return http.StatusBadRequest
	case string(domain.StatusPermissionDen):
		return http.StatusForbidden
	case string(domain.StatusCannotAnswer):
		return http.StatusOK // Still return 200 for cannot_answer as it's a valid response
	case string(domain.StatusInternalError):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// writeJSON writes a JSON response.
func (h *AnswerHandler) writeJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a simple error response.
func (h *AnswerHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// HealthHandler handles GET /health requests.
type HealthHandler struct{}

// NewHealthHandler creates a new health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServeHTTP handles the health check request.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "access-layer",
		"version": "v1",
	})
}
