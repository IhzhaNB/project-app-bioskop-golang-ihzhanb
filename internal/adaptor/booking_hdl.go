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

type BookingHandler struct {
	service usecase.BookingService
	log     *zap.Logger
}

func NewBookingHandler(service usecase.BookingService, log *zap.Logger) *BookingHandler {
	return &BookingHandler{
		service: service,
		log:     log.With(zap.String("handler", "booking")),
	}
}

// CreateBooking handles POST /api/booking (protected)
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	var req request.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	booking, err := h.service.CreateBooking(r.Context(), userID.String(), &req)
	if err != nil {
		h.handleServiceError(w, err, "create booking")
		return
	}

	utils.ResponseCreated(w, "success", booking)
}

// GetUserBookings handles GET /api/user/bookings (protected)
func (h *BookingHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
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

	bookings, err := h.service.GetUserBookings(r.Context(), userID.String(), req)
	if err != nil {
		h.handleServiceError(w, err, "get user bookings")
		return
	}

	utils.ResponseSuccess(w, "success", bookings)
}

// ProcessPayment handles POST /api/pay (protected)
func (h *BookingHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		utils.ResponseUnauthorized(w, "Authentication required")
		return
	}

	var req request.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	payment, err := h.service.ProcessPayment(r.Context(), userID.String(), &req)
	if err != nil {
		h.handleServiceError(w, err, "process payment")
		return
	}

	utils.ResponseSuccess(w, "success", payment)
}

// GetPaymentMethods handles GET /api/payment-methods (public)
func (h *BookingHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	paymentMethods, err := h.service.GetPaymentMethods(r.Context())
	if err != nil {
		h.handleServiceError(w, err, "get payment methods")
		return
	}

	utils.ResponseSuccess(w, "success", paymentMethods)
}

// ==================== ADMIN METHODS ====================

// GetBookingByID handles GET /api/admin/bookings/{id} (admin only)
func (h *BookingHandler) GetBookingByID(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		utils.ResponseBadRequest(w, "Booking ID is required", nil)
		return
	}

	booking, err := h.service.GetBookingByID(r.Context(), bookingID)
	if err != nil {
		h.handleServiceError(w, err, "get booking by ID")
		return
	}

	utils.ResponseSuccess(w, "success", booking)
}

// CancelBooking handles PUT /api/admin/bookings/{id}/cancel (admin only)
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	if bookingID == "" {
		utils.ResponseBadRequest(w, "Booking ID is required", nil)
		return
	}

	if err := h.service.CancelBooking(r.Context(), bookingID); err != nil {
		h.handleServiceError(w, err, "cancel booking")
		return
	}

	utils.ResponseSuccess(w, "success", nil)
}

// handleServiceError handles errors untuk booking operations
func (h *BookingHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
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

	case strings.Contains(errMsg, "already booked"):
		h.log.Warn(operation+" failed - seat already booked",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseBadRequest(w, errMsg, nil)

	case strings.Contains(errMsg, "unauthorized"):
		h.log.Warn(operation+" failed - unauthorized",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseUnauthorized(w, errMsg)

	case strings.Contains(errMsg, "cannot"):
		h.log.Warn(operation+" failed - invalid state",
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseBadRequest(w, errMsg, nil)

	default:
		h.log.Error("Failed to "+operation,
			zap.Error(err),
			zap.String("operation", operation))
		utils.ResponseInternalError(w, "Internal server error")
	}
}
