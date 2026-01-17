package adaptor

import (
	"encoding/json"
	"net/http"
	"strings"

	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/usecase"
	"cinema-booking/pkg/utils"

	"go.uber.org/zap"
)

type AuthHandler struct {
	service usecase.AuthService
	log     *zap.Logger
}

func NewAuthHandler(service usecase.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     log,
	}
}

// Register handles POST /api/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	// Call service
	response, err := h.service.Register(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err, "register")
		return
	}

	utils.ResponseCreated(w, "success", response)
}

// Login handles POST /api/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	response, err := h.service.Login(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err, "login")
		return
	}

	utils.ResponseSuccess(w, "success", response)
}

// Logout handles POST /api/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token dari Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.ResponseBadRequest(w, "No token provided", nil)
		return
	}

	// Format: "Bearer <token-uuid>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.ResponseBadRequest(w, "Invalid token format. Use: Bearer <token>", nil)
		return
	}

	token := parts[1]

	if err := h.service.Logout(r.Context(), token); err != nil {
		h.handleServiceError(w, err, "logout")
		return
	}

	utils.ResponseSuccess(w, "success", nil)
}

// SendOTP handles POST /api/send-otp
func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req request.SendOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	if err := h.service.SendOTP(r.Context(), req.Email, req.Type); err != nil {
		h.handleServiceError(w, err, "send OTP")
		return
	}

	utils.ResponseSuccess(w, "success", nil)
}

// VerifyEmail handles POST /api/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req request.VerifyEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseBadRequest(w, "Invalid request body", nil)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(req); len(validationErrors) > 0 {
		utils.ResponseBadRequest(w, "Validation failed", validationErrors)
		return
	}

	if err := h.service.VerifyEmail(r.Context(), &req); err != nil {
		h.handleServiceError(w, err, "verify email")
		return
	}

	utils.ResponseSuccess(w, "success", nil)
}

// handleServiceError categorizes service errors and returns appropriate HTTP responses
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error, operation string) {
	errMsg := err.Error()

	// Check error message patterns to determine error type
	switch {
	case strings.Contains(errMsg, "not found"):
		h.log.Warn(operation+" failed - not found", zap.Error(err))
		utils.ResponseNotFound(w, errMsg)

	case strings.Contains(errMsg, "already registered"),
		strings.Contains(errMsg, "already taken"),
		strings.Contains(errMsg, "already verified"):
		h.log.Warn(operation+" failed - already exists", zap.Error(err))
		utils.ResponseBadRequest(w, errMsg, err)

	case strings.Contains(errMsg, "invalid credentials"),
		strings.Contains(errMsg, "incorrect"),
		strings.Contains(errMsg, "invalid password"):
		h.log.Warn(operation+" failed - invalid credentials", zap.Error(err))
		utils.ResponseUnauthorized(w, errMsg)

	case strings.Contains(errMsg, "deactivated"):
		h.log.Warn(operation+" failed - account deactivated", zap.Error(err))
		utils.ResponseForbidden(w, errMsg)

	case strings.Contains(errMsg, "validation failed"):
		h.log.Warn(operation+" validation failed", zap.Error(err))
		utils.ResponseBadRequest(w, errMsg, err)

	case strings.Contains(errMsg, "invalid or expired"):
		h.log.Warn(operation+" failed - invalid OTP", zap.Error(err))
		utils.ResponseBadRequest(w, errMsg, err)

	default:
		h.log.Error("Failed to "+operation, zap.Error(err), zap.String("operation", operation))
		utils.ResponseInternalError(w, "Internal server error")
	}
}
