package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// wireUser configures user management routes with role-based access control
func wireUser(
	r chi.Router,
	userHandler *adaptor.UserHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PROTECTED USER ROUTES ====================
	// User profile - requires authentication
	r.With(middleware.AuthSession(repo.Session, log)).Get("/api/user/profile", userHandler.GetProfile)

	// ==================== ADMIN ROUTES ====================
	// Admin user management - requires both authentication AND admin role
	r.With(
		middleware.AuthSession(repo.Session, log), // Check valid session
		middleware.Admin(repo.User, log),          // Check admin role
	).Route("/api/admin/users", func(r chi.Router) {
		r.Get("/", userHandler.GetAllUsers)       // GET /api/admin/users?page=1&per_page=10
		r.Delete("/{id}", userHandler.DeleteUser) // DELETE /api/admin/users/{user-id}
	})
}
