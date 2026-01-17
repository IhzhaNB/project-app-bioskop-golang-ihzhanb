package adaptor

import (
	"encoding/json"
	"net/http"
	"strings"

	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/usecase"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MovieHandler struct {
	service usecase.MovieService
	log     *zap.Logger
}

func NewMovieHandler(service usecase.MovieService, log *zap.Logger) *MovieHandler {
	return &MovieHandler{
		service: service,
		log:     log.With(zap.String("handler", "movie")),
	}
}

// GetMovies handles GET /api/movies (sesuai requirement)
func (h *MovieHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := &request.PaginatedRequest{
		Page:    1,
		PerPage: 10,
	}

	query := r.URL.Query()
	req.Page = utils.ParseInt(query.Get("page"), 1)
	req.PerPage = utils.ParseInt(query.Get("per_page"), 10)

	// Parse optional filter parameter
	var releaseStatus *string
	if status := query.Get("release_status"); status != "" {
		// Map "now" to "now_playing" for compatibility
		if status == "now_playing" || status == "coming_soon" || status == "now" {
			if status == "now" {
				status = "now_playing"
			}
			releaseStatus = &status
		} else {
			h.log.Warn("Invalid release_status filter", zap.String("status", status))
		}
	}

	// Call service
	movies, err := h.service.GetMovies(r.Context(), req, releaseStatus)
	if err != nil {
		h.handleServiceError(w, err, "get movies")
		return
	}

	utils.ResponseSuccess(w, "success", movies)
}

// GetMovieByID handles GET /api/movies/{id} (optional)
func (h *MovieHandler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "id")
	if movieID == "" {
		utils.ResponseBadRequest(w, "Movie ID is required", nil)
		return
	}

	movie, err := h.service.GetMovieByID(r.Context(), movieID)
	if err != nil {
		h.handleServiceError(w, err, "get movie by ID")
		return
	}

	utils.ResponseSuccess(w, "success", movie)
}

// CreateMovie handles POST /api/admin/movies (admin only - optional)
func (h *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var req request.MovieRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	movie, err := h.service.CreateMovie(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err, "create movie")
		return
	}

	utils.ResponseCreated(w, "Movie created successfully", movie)
}

// UpdateMovie handles PUT /api/admin/movies/{id} (admin only - optional)
func (h *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "id")
	if movieID == "" {
		utils.ResponseBadRequest(w, "Movie ID is required", nil)
		return
	}

	var req request.MovieUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// ✅ FIX: Tambah validation untuk update (optional fields)
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	movie, err := h.service.UpdateMovie(r.Context(), movieID, &req)
	if err != nil {
		h.handleServiceError(w, err, "update movie")
		return
	}

	utils.ResponseSuccess(w, "Movie updated successfully", movie)
}

// DeleteMovie handles DELETE /api/admin/movies/{id} (admin only - optional)
func (h *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "id")
	if movieID == "" {
		utils.ResponseBadRequest(w, "Movie ID is required", nil)
		return
	}

	if err := h.service.DeleteMovie(r.Context(), movieID); err != nil {
		h.handleServiceError(w, err, "delete movie")
		return
	}

	utils.ResponseSuccess(w, "Movie deleted successfully", nil)
}

// handleServiceError handles errors untuk movie operations
func (h *MovieHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "not found"):
		h.log.Warn(operation+" failed - not found",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseNotFound(w, errMsg)

	case strings.Contains(errMsg, "validation failed"):
		h.log.Warn(operation+" validation failed",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseBadRequest(w, errMsg, nil)

	case strings.Contains(errMsg, "invalid"):
		h.log.Warn("Invalid input for "+operation,
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseBadRequest(w, errMsg, nil)

	case strings.Contains(errMsg, "already exists"):
		h.log.Warn(operation+" failed - already exists",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseBadRequest(w, errMsg, err)

	default:
		h.log.Error("Failed to "+operation,
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseInternalError(w, "Internal server error")
	}
}
