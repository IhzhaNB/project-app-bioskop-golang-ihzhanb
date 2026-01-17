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

type CinemaHandler struct {
	service usecase.CinemaService
	log     *zap.Logger
}

func NewCinemaHandler(service usecase.CinemaService, log *zap.Logger) *CinemaHandler {
	return &CinemaHandler{
		service: service,
		log:     log.With(zap.String("handler", "cinema")),
	}
}

// GetCinemas handles GET /api/cinemas (public)
func (h *CinemaHandler) GetCinemas(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req := &request.PaginatedRequest{
		Page:    1,
		PerPage: 10,
	}

	query := r.URL.Query()
	req.Page = utils.ParseInt(query.Get("page"), 1)
	req.PerPage = utils.ParseInt(query.Get("per_page"), 10)

	// Filter by city (optional)
	var cityFilter *string
	if city := query.Get("city"); city != "" {
		cityFilter = &city
	}

	// Call service
	cinemas, err := h.service.GetCinemas(r.Context(), req, cityFilter)
	if err != nil {
		h.handleServiceError(w, err, "get cinemas")
		return
	}

	utils.ResponseSuccess(w, "success", cinemas)
}

// GetCinemaByID handles GET /api/cinemas/{id} (public)
func (h *CinemaHandler) GetCinemaByID(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "id")
	if cinemaID == "" {
		utils.ResponseBadRequest(w, "Cinema ID is required", nil)
		return
	}

	cinema, err := h.service.GetCinemaByID(r.Context(), cinemaID)
	if err != nil {
		h.handleServiceError(w, err, "get cinema by ID")
		return
	}

	utils.ResponseSuccess(w, "success", cinema)
}

// GetSeatAvailability handles GET /api/cinemas/{id}/seats (public)
func (h *CinemaHandler) GetSeatAvailability(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "id")
	if cinemaID == "" {
		utils.ResponseBadRequest(w, "Cinema ID is required", nil)
		return
	}

	// Get date and time from query parameters
	query := r.URL.Query()
	date := query.Get("date")
	time := query.Get("time")

	if date == "" || time == "" {
		utils.ResponseBadRequest(w, "Both date and time query parameters are required", nil)
		return
	}

	// Call service
	seatAvailability, err := h.service.GetSeatAvailability(r.Context(), cinemaID, date, time)
	if err != nil {
		h.handleServiceError(w, err, "get seat availability")
		return
	}

	utils.ResponseSuccess(w, "success", seatAvailability)
}

// CreateCinema handles POST /api/admin/cinemas
func (h *CinemaHandler) CreateCinema(w http.ResponseWriter, r *http.Request) {
	var req request.CinemaRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	cinema, err := h.service.CreateCinema(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err, "create cinema")
		return
	}

	utils.ResponseCreated(w, "success", cinema)
}

// UpdateCinema handles PUT /api/admin/cinemas/{id}
func (h *CinemaHandler) UpdateCinema(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "id")
	if cinemaID == "" {
		utils.ResponseBadRequest(w, "Cinema ID is required", nil)
		return
	}

	var req request.CinemaUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	cinema, err := h.service.UpdateCinema(r.Context(), cinemaID, &req)
	if err != nil {
		h.handleServiceError(w, err, "update cinema")
		return
	}

	utils.ResponseSuccess(w, "success", cinema)
}

// DeleteCinema handles DELETE /api/admin/cinemas/{id}
func (h *CinemaHandler) DeleteCinema(w http.ResponseWriter, r *http.Request) {
	cinemaID := chi.URLParam(r, "id")
	if cinemaID == "" {
		utils.ResponseBadRequest(w, "Cinema ID is required", nil)
		return
	}

	if err := h.service.DeleteCinema(r.Context(), cinemaID); err != nil {
		h.handleServiceError(w, err, "delete cinema")
		return
	}

	utils.ResponseSuccess(w, "success", nil)
}

// handleServiceError handles errors untuk cinema operations
func (h *CinemaHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
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
