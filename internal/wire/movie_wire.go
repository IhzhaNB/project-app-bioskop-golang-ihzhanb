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

	// GET /api/movies (public - sesuai requirement)
	r.Get("/api/movies", movieHandler.GetMovies)

	// GET /api/movies/{id} (public - optional)
	r.Get("/api/movies/{id}", movieHandler.GetMovieByID)

	// ==================== ADMIN ROUTES ====================

	r.Route("/api/admin/movies", func(r chi.Router) {
		// Apply auth + admin middleware
		r.Use(middleware.AuthSession(repo.Session, log))
		r.Use(middleware.Admin(repo.User, log))

		r.Post("/", movieHandler.CreateMovie)
		r.Put("/{id}", movieHandler.UpdateMovie)
		r.Delete("/{id}", movieHandler.DeleteMovie)
	})
}
