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

type ReviewHandler struct {
	service usecase.ReviewService
	log     *zap.Logger
}

func NewReviewHandler(service usecase.ReviewService, log *zap.Logger) *ReviewHandler {
	return &ReviewHandler{
		service: service,
		log:     log.With(zap.String("handler", "review")),
	}
}

// CreateReview handles POST /api/reviews (protected)
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	var req request.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	review, err := h.service.CreateReview(r.Context(), userID.String(), &req)
	if err != nil {
		h.handleServiceError(w, err, "create review")
		return
	}

	utils.ResponseCreated(w, "success", review)
}

// GetMovieReviews handles GET /api/movies/{id}/reviews (public)
func (h *ReviewHandler) GetMovieReviews(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "id")
	if movieID == "" {
		utils.ResponseBadRequest(w, "Movie ID is required", nil)
		return
	}

	req := &request.PaginatedRequest{
		Page:    1,
		PerPage: 10,
	}

	// Parse query parameters
	query := r.URL.Query()
	req.Page = utils.ParseInt(query.Get("page"), 1)
	req.PerPage = utils.ParseInt(query.Get("per_page"), 10)

	reviews, err := h.service.GetMovieReviews(r.Context(), movieID, req)
	if err != nil {
		h.handleServiceError(w, err, "get movie reviews")
		return
	}

	utils.ResponseSuccess(w, "success", reviews)
}

// GetUserReviews handles GET /api/user/reviews (protected)
func (h *ReviewHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	req := &request.PaginatedRequest{
		Page:    1,
		PerPage: 10,
	}

	// Parse query parameters
	query := r.URL.Query()
	req.Page = utils.ParseInt(query.Get("page"), 1)
	req.PerPage = utils.ParseInt(query.Get("per_page"), 10)

	reviews, err := h.service.GetUserReviews(r.Context(), userID.String(), req)
	if err != nil {
		h.handleServiceError(w, err, "get user reviews")
		return
	}

	utils.ResponseSuccess(w, "success", reviews)
}

// UpdateReview handles PUT /api/reviews/{id} (protected)
func (h *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		utils.ResponseBadRequest(w, "Review ID is required", nil)
		return
	}

	var req request.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	review, err := h.service.UpdateReview(r.Context(), reviewID, userID.String(), &req)
	if err != nil {
		h.handleServiceError(w, err, "update review")
		return
	}

	utils.ResponseSuccess(w, "success", review)
}

// DeleteReview handles DELETE /api/reviews/{id} (protected)
func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		utils.ResponseBadRequest(w, "Review ID is required", nil)
		return
	}

	if err := h.service.DeleteReview(r.Context(), reviewID, userID.String()); err != nil {
		h.handleServiceError(w, err, "delete review")
		return
	}

	utils.ResponseSuccess(w, "success", nil)
}

// GetMovieReviewStats handles GET /api/movies/{id}/review-stats (public)
func (h *ReviewHandler) GetMovieReviewStats(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "id")
	if movieID == "" {
		utils.ResponseBadRequest(w, "Movie ID is required", nil)
		return
	}

	stats, err := h.service.GetMovieReviewStats(r.Context(), movieID)
	if err != nil {
		h.handleServiceError(w, err, "get movie review stats")
		return
	}

	utils.ResponseSuccess(w, "success", stats)
}

// handleServiceError handles errors untuk review operations
func (h *ReviewHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
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

	case strings.Contains(errMsg, "already reviewed"):
		h.log.Warn(operation+" failed - already reviewed",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseBadRequest(w, errMsg, nil)

	case strings.Contains(errMsg, "unauthorized"):
		h.log.Warn(operation+" failed - unauthorized",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseUnauthorized(w, errMsg)

	default:
		h.log.Error("Failed to "+operation,
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseInternalError(w, "Internal server error")
	}
}
