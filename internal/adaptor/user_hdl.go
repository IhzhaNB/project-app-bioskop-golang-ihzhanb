package adaptor

import (
	"net/http"
	"strconv"
	"strings"

	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/usecase"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type UserHandler struct {
	service usecase.UserService
	log     *zap.Logger
}

func NewUserHandler(service usecase.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		log:     log,
	}
}

// GetProfile handles GET /api/users/profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), userID.String())
	if err != nil {
		h.handleServiceError(w, err, "get profile")
		return
	}

	utils.ResponseSuccess(w, "Profile retrieved successfully", profile)
}

// GetAllUsers handles GET /api/admin/users (admin only)
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	req := &request.PaginatedRequest{
		Page:    1,
		PerPage: 10,
	}

	// Parse query parameters
	query := r.URL.Query()
	req.Page = h.parseInt(query.Get("page"), 1)
	req.PerPage = h.parseInt(query.Get("per_page"), 10)

	// Validate per_page max
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	users, err := h.service.GetAllUsers(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err, "get all users")
		return
	}

	utils.ResponseSuccess(w, "Users retrieved successfully", users)
}

// DeleteUser handles DELETE /api/admin/users/{id} (admin only)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		utils.ResponseBadRequest(w, "User ID is required", nil)
		return
	}

	if err := h.service.DeleteUser(r.Context(), userID); err != nil {
		h.handleServiceError(w, err, "delete user")
		return
	}

	utils.ResponseSuccess(w, "User deleted successfully", nil)
}

// handleServiceError handles errors for user operations
func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "not found"):
		h.log.Warn(operation+" failed - not found", zap.Error(err))
		utils.ResponseNotFound(w, errMsg)

	case strings.Contains(errMsg, "validation failed"):
		h.log.Warn(operation+" validation failed", zap.Error(err))
		utils.ResponseBadRequest(w, errMsg, err)

	case strings.Contains(errMsg, "invalid"):
		h.log.Warn("Invalid input for "+operation, zap.Error(err))
		utils.ResponseBadRequest(w, errMsg, err)

	case strings.Contains(errMsg, "unauthorized"),
		strings.Contains(errMsg, "authentication"):
		h.log.Warn(operation+" failed - unauthorized", zap.Error(err))
		utils.ResponseUnauthorized(w, errMsg)

	case strings.Contains(errMsg, "forbidden"):
		h.log.Warn(operation+" failed - forbidden", zap.Error(err))
		utils.ResponseForbidden(w, errMsg)

	default:
		h.log.Error("Failed to "+operation,
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseInternalError(w, "Internal server error")
	}
}

// parseInt helper untuk parse query parameters
func (h *UserHandler) parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if result < 1 {
		return defaultValue
	}

	return result
}
