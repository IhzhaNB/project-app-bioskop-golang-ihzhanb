// internal/wire/wire.go
package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/usecase"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// App menyimpan semua dependencies
type App struct {
	Router *chi.Mux
}

// Wiring menginisialisasi semua dependencies
func Wiring(repo *repository.Repository, config *utils.Config, logger *zap.Logger) *App {
	// Initialize services dan handlers
	service := usecase.NewService(repo, config, logger)
	handler := adaptor.NewHandler(service, logger)

	// Setup router
	router := setupRouter(handler, repo, config, logger)

	return &App{
		Router: router,
	}
}

// setupRouter konfigurasi Chi router
func setupRouter(
	handler *adaptor.Handler,
	repo *repository.Repository,
	config *utils.Config,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recover(logger))
	r.Use(middleware.CORS())

	// Apply routes
	wireAuth(r, handler.Auth, repo, config, logger)
	wireUser(r, handler.User, repo, config, logger)
	wireMovie(r, handler.Movie, repo, config, logger)
	wireCinema(r, handler.Cinema, repo, config, logger)
	wireBooking(r, handler.Booking, repo, config, logger)
	wireReview(r, handler.Review, repo, config, logger)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
