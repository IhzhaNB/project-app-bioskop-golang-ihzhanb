package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func wireCinema(
	r chi.Router,
	cinemaHandler *adaptor.CinemaHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PUBLIC ROUTES ====================
	// GET /api/cinemas - List all cinemas (public)
	r.Get("/api/cinemas", cinemaHandler.GetCinemas)

	// GET /api/cinemas/{id} - Get specific cinema details (public)
	r.Get("/api/cinemas/{id}", cinemaHandler.GetCinemaByID)

	// GET /api/cinemas/{id}/seats - Check seat availability (public)
	// Requires query params: ?date=2024-01-16&time=14:30
	r.Get("/api/cinemas/{id}/seats", cinemaHandler.GetSeatAvailability)

	// ==================== ADMIN ROUTES ====================
	// Group admin routes under /api/admin/cinemas
	r.Route("/api/admin/cinemas", func(r chi.Router) {
		// Apply middleware chain: AuthSession → Admin
		r.Use(middleware.AuthSession(repo.Session, log))
		r.Use(middleware.Admin(repo.User, log))

		// Cinema CRUD operations (admin only)
		r.Post("/", cinemaHandler.CreateCinema)       // Create new cinema
		r.Put("/{id}", cinemaHandler.UpdateCinema)    // Update existing cinema
		r.Delete("/{id}", cinemaHandler.DeleteCinema) // Delete cinema
	})
}
