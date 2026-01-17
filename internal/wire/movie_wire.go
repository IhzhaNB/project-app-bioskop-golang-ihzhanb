package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func wireMovie(
	r chi.Router,
	movieHandler *adaptor.MovieHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PUBLIC ROUTES ====================
	// GET /api/movies - List movies (public, anyone can view)
	r.Get("/api/movies", movieHandler.GetMovies)

	// GET /api/movies/{id} - Movie details (public)
	r.Get("/api/movies/{id}", movieHandler.GetMovieByID)

	// ==================== ADMIN ROUTES ====================
	// Group admin routes with middleware chain
	r.Route("/api/admin/movies", func(r chi.Router) {
		// Apply middleware to all routes in this group
		r.Use(middleware.AuthSession(repo.Session, log)) // Must be authenticated
		r.Use(middleware.Admin(repo.User, log))          // Must be admin

		// Admin movie management endpoints
		r.Post("/", movieHandler.CreateMovie)       // POST /api/admin/movies
		r.Put("/{id}", movieHandler.UpdateMovie)    // PUT /api/admin/movies/{id}
		r.Delete("/{id}", movieHandler.DeleteMovie) // DELETE /api/admin/movies/{id}
	})
}
