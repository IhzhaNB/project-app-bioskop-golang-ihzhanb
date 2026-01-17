package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// wireAuth configures authentication-related routes with appropriate middleware
func wireAuth(
	r chi.Router,
	authHandler *adaptor.AuthHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PUBLIC ROUTES ====================
	// These endpoints don't require authentication
	r.Post("/api/register", authHandler.Register)        // User registration
	r.Post("/api/login", authHandler.Login)              // User login
	r.Post("/api/send-otp", authHandler.SendOTP)         // Request OTP for verification
	r.Post("/api/verify-email", authHandler.VerifyEmail) // Verify email with OTP

	// ==================== PROTECTED ROUTES ====================
	// Logout requires valid session (can't logout without being logged in)
	r.With(middleware.AuthSession(repo.Session, log)).Post("/api/logout", authHandler.Logout)
}
