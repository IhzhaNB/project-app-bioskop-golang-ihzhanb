package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func wireReview(
	r chi.Router,
	reviewHandler *adaptor.ReviewHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PUBLIC ROUTES ====================
	// GET /api/movies/{id}/reviews - View movie reviews (public)
	r.Get("/api/movies/{id}/reviews", reviewHandler.GetMovieReviews)

	// GET /api/movies/{id}/review-stats - View rating statistics (public)
	r.Get("/api/movies/{id}/review-stats", reviewHandler.GetMovieReviewStats)

	// ==================== PROTECTED ROUTES (require auth) ====================
	// Group routes that require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthSession(repo.Session, log))

		// POST /api/reviews - Create new review (authenticated users only)
		r.Post("/api/reviews", reviewHandler.CreateReview)

		// GET /api/user/reviews - View user's own reviews
		r.Get("/api/user/reviews", reviewHandler.GetUserReviews)

		// PUT /api/reviews/{id} - Update review (owner only)
		r.Put("/api/reviews/{id}", reviewHandler.UpdateReview)

		// DELETE /api/reviews/{id} - Delete review (owner only)
		r.Delete("/api/reviews/{id}", reviewHandler.DeleteReview)
	})
}
