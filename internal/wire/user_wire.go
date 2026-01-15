package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func wireUser(
	r chi.Router,
	userHandler *adaptor.UserHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PROTECTED USER ROUTES ====================
	// User profile - butuh auth
	r.With(middleware.AuthSession(repo.Session, log)).Get("/api/user/profile", userHandler.GetProfile)

	// ==================== ADMIN ROUTES ====================
	// Admin routes - butuh auth + admin role
	r.With(
		middleware.AuthSession(repo.Session, log),
		middleware.Admin(repo.User, log),
	).Route("/api/admin/users", func(r chi.Router) {
		r.Get("/", userHandler.GetAllUsers)
		r.Delete("/{id}", userHandler.DeleteUser)
	})
}
