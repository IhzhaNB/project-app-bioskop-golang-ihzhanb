package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func wireAuth(
	r chi.Router,
	authHandler *adaptor.AuthHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PUBLIC ROUTES ====================
	// Public routes (tanpa auth middleware)
	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)
	r.Post("/api/send-otp", authHandler.SendOTP)
	r.Post("/api/verify-email", authHandler.VerifyEmail)

	// ==================== PROTECTED ROUTES ====================
	// Logout - PROTECTED (butuh auth)
	r.With(middleware.AuthSession(repo.Session, log)).Post("/api/logout", authHandler.Logout)
}
