package usecase

import (
	"context"
	"fmt"
	"time"

	"cinema-booking/internal/data/entity"
	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/dto/response"
	"cinema-booking/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReviewService interface {
	// Public endpoints
	CreateReview(ctx context.Context, userID string, req *request.CreateReviewRequest) (*response.ReviewResponse, error)
	GetMovieReviews(ctx context.Context, movieID string, req *request.PaginatedRequest) (*response.PaginatedResponse[response.ReviewResponse], error)
	GetUserReviews(ctx context.Context, userID string, req *request.PaginatedRequest) (*response.PaginatedResponse[response.ReviewResponse], error)
	UpdateReview(ctx context.Context, reviewID, userID string, req *request.UpdateReviewRequest) (*response.ReviewResponse, error)
	DeleteReview(ctx context.Context, reviewID, userID string) error

	// Stats
	GetMovieReviewStats(ctx context.Context, movieID string) (*response.MovieReviewStats, error)
}

type reviewService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewReviewService(repo *repository.Repository, log *zap.Logger) ReviewService {
	return &reviewService{
		repo: repo,
		log:  log.With(zap.String("service", "review")),
	}
}

func (s *reviewService) CreateReview(ctx context.Context, userID string, req *request.CreateReviewRequest) (*response.ReviewResponse, error) {
	// Validate request
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Create review validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Parse IDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	movieID, err := uuid.Parse(req.MovieID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie ID format %s: %w", req.MovieID, err)
	}

	// Check if movie exists
	movie, err := s.repo.Movie.FindByID(ctx, movieID)
	if err != nil || movie == nil {
		return nil, fmt.Errorf("movie %s not found", req.MovieID)
	}

	// Check if user has already reviewed this movie
	existingReview, err := s.repo.Review.FindByUserAndMovie(ctx, userUUID, movieID)
	if err != nil {
		s.log.Error("Failed to check existing review", zap.Error(err))
		return nil, fmt.Errorf("check existing review: %w", err)
	}

	if existingReview != nil {
		return nil, fmt.Errorf("user already reviewed this movie")
	}

	// Create review entity
	now := time.Now()
	review := &entity.Review{
		BaseSimple: entity.BaseSimple{
			ID:        uuid.New(),
			CreatedAt: now,
		},
		UserID:  userUUID,
		MovieID: movieID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	// Save review
	if err := s.repo.Review.Create(ctx, review); err != nil {
		s.log.Error("Failed to create review",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("movie_id", req.MovieID),
		)
		return nil, fmt.Errorf("create review: %w", err)
	}

	// Update movie rating
	if err := s.updateMovieRating(ctx, movieID); err != nil {
		s.log.Warn("Failed to update movie rating",
			zap.Error(err),
			zap.String("movie_id", req.MovieID),
		)
		// Continue anyway
	}

	// Get user and movie info for response
	user, _ := s.repo.User.FindByID(ctx, userUUID)
	username := ""
	if user != nil {
		username = user.Username
	}

	s.log.Info("Review created",
		zap.String("review_id", review.ID.String()),
		zap.String("user_id", userID),
		zap.String("movie_id", req.MovieID),
		zap.Int("rating", req.Rating),
	)

	reviewResp := response.ReviewToResponse(review, username, movie.Title)
	return &reviewResp, nil
}

func (s *reviewService) GetMovieReviews(ctx context.Context, movieID string, req *request.PaginatedRequest) (*response.PaginatedResponse[response.ReviewResponse], error) {
	// Parse movie ID
	movieUUID, err := uuid.Parse(movieID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie ID format %s: %w", movieID, err)
	}

	limit := req.Limit()
	offset := req.Offset()

	// Get reviews
	reviews, err := s.repo.Review.FindByMovieID(ctx, movieUUID, limit, offset)
	if err != nil {
		s.log.Error("Failed to get movie reviews",
			zap.Error(err),
			zap.String("movie_id", movieID),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
		)
		return nil, fmt.Errorf("get movie reviews: %w", err)
	}

	// Get total count
	total, err := s.repo.Review.CountByMovieID(ctx, movieUUID)
	if err != nil {
		s.log.Error("Failed to count movie reviews", zap.Error(err))
		return nil, fmt.Errorf("count movie reviews: %w", err)
	}

	// Get movie info
	movie, _ := s.repo.Movie.FindByID(ctx, movieUUID)
	movieTitle := ""
	if movie != nil {
		movieTitle = movie.Title
	}

	// Convert to response
	reviewResponses := make([]response.ReviewResponse, len(reviews))
	for i, review := range reviews {
		// Get user info
		user, _ := s.repo.User.FindByID(ctx, review.UserID)
		username := ""
		if user != nil {
			username = user.Username
		}

		reviewResponses[i] = response.ReviewToResponse(review, username, movieTitle)
	}

	s.log.Info("Movie reviews retrieved",
		zap.String("movie_id", movieID),
		zap.Int("count", len(reviews)),
		zap.Int64("total", total),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
	)

	return response.NewPaginatedResponse(reviewResponses, req.Page, req.PerPage, total), nil
}

func (s *reviewService) GetUserReviews(ctx context.Context, userID string, req *request.PaginatedRequest) (*response.PaginatedResponse[response.ReviewResponse], error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	limit := req.Limit()
	offset := req.Offset()

	// Get reviews
	reviews, err := s.repo.Review.FindByUserID(ctx, userUUID, limit, offset)
	if err != nil {
		s.log.Error("Failed to get user reviews",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
		)
		return nil, fmt.Errorf("get user reviews: %w", err)
	}

	// Get total count (simplified - bisa pakai CountByUserID kalau ada)
	total := int64(len(reviews)) // Simplified

	// Get user info
	user, _ := s.repo.User.FindByID(ctx, userUUID)
	username := ""
	if user != nil {
		username = user.Username
	}

	// Convert to response
	reviewResponses := make([]response.ReviewResponse, len(reviews))
	for i, review := range reviews {
		// Get movie info
		movie, _ := s.repo.Movie.FindByID(ctx, review.MovieID)
		movieTitle := ""
		if movie != nil {
			movieTitle = movie.Title
		}

		reviewResponses[i] = response.ReviewToResponse(review, username, movieTitle)
	}

	s.log.Info("User reviews retrieved",
		zap.String("user_id", userID),
		zap.Int("count", len(reviews)),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
	)

	return response.NewPaginatedResponse(reviewResponses, req.Page, req.PerPage, total), nil
}

func (s *reviewService) UpdateReview(ctx context.Context, reviewID, userID string, req *request.UpdateReviewRequest) (*response.ReviewResponse, error) {
	// Parse IDs
	reviewUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return nil, fmt.Errorf("invalid review ID format %s: %w", reviewID, err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	// Get existing review
	review, err := s.repo.Review.FindByID(ctx, reviewUUID)
	if err != nil || review == nil {
		return nil, fmt.Errorf("review %s not found", reviewID)
	}

	// Check if review belongs to user
	if review.UserID != userUUID {
		return nil, fmt.Errorf("unauthorized to update this review")
	}

	// Update fields if provided
	updated := false

	if req.Rating != nil && *req.Rating != review.Rating {
		review.Rating = *req.Rating
		updated = true
	}

	if req.Comment != nil {
		review.Comment = req.Comment
		updated = true
	}

	if !updated {
		// No changes
		return s.buildReviewResponse(ctx, review), nil
	}

	// Save updated review
	if err := s.repo.Review.Update(ctx, review); err != nil {
		s.log.Error("Failed to update review",
			zap.Error(err),
			zap.String("review_id", reviewID),
		)
		return nil, fmt.Errorf("update review: %w", err)
	}

	// Update movie rating
	if err := s.updateMovieRating(ctx, review.MovieID); err != nil {
		s.log.Warn("Failed to update movie rating",
			zap.Error(err),
			zap.String("movie_id", review.MovieID.String()),
		)
		// Continue anyway
	}

	s.log.Info("Review updated",
		zap.String("review_id", reviewID),
		zap.String("user_id", userID),
		zap.Bool("was_updated", updated),
	)

	return s.buildReviewResponse(ctx, review), nil
}

func (s *reviewService) DeleteReview(ctx context.Context, reviewID, userID string) error {
	// Parse IDs
	reviewUUID, err := uuid.Parse(reviewID)
	if err != nil {
		return fmt.Errorf("invalid review ID format %s: %w", reviewID, err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	// Get existing review
	review, err := s.repo.Review.FindByID(ctx, reviewUUID)
	if err != nil || review == nil {
		return fmt.Errorf("review %s not found", reviewID)
	}

	// Check if review belongs to user
	if review.UserID != userUUID {
		return fmt.Errorf("unauthorized to delete this review")
	}

	// Delete review
	if err := s.repo.Review.Delete(ctx, reviewUUID); err != nil {
		s.log.Error("Failed to delete review",
			zap.Error(err),
			zap.String("review_id", reviewID),
		)
		return fmt.Errorf("delete review: %w", err)
	}

	// Update movie rating
	if err := s.updateMovieRating(ctx, review.MovieID); err != nil {
		s.log.Warn("Failed to update movie rating",
			zap.Error(err),
			zap.String("movie_id", review.MovieID.String()),
		)
		// Continue anyway
	}

	s.log.Info("Review deleted",
		zap.String("review_id", reviewID),
		zap.String("user_id", userID),
		zap.String("movie_id", review.MovieID.String()),
	)

	return nil
}

func (s *reviewService) GetMovieReviewStats(ctx context.Context, movieID string) (*response.MovieReviewStats, error) {
	// Parse movie ID
	movieUUID, err := uuid.Parse(movieID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie ID format %s: %w", movieID, err)
	}

	avgRating, reviewCount, err := s.repo.Review.GetMovieReviewStats(ctx, movieUUID)
	if err != nil {
		s.log.Error("Failed to get movie review stats",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		return nil, fmt.Errorf("get movie review stats: %w", err)
	}

	return &response.MovieReviewStats{
		AverageRating: avgRating,
		ReviewCount:   reviewCount,
	}, nil
}

// ==================== HELPER METHODS ====================

func (s *reviewService) updateMovieRating(ctx context.Context, movieID uuid.UUID) error {
	avgRating, err := s.repo.Review.GetMovieAverageRating(ctx, movieID)
	if err != nil {
		return fmt.Errorf("get average rating: %w", err)
	}

	// Update movie rating in movies table
	if err := s.repo.Movie.UpdateRating(ctx, movieID, avgRating); err != nil {
		return fmt.Errorf("update movie rating: %w", err)
	}

	s.log.Debug("Movie rating updated",
		zap.String("movie_id", movieID.String()),
		zap.Float64("new_rating", avgRating),
	)

	return nil
}

func (s *reviewService) buildReviewResponse(ctx context.Context, review *entity.Review) *response.ReviewResponse {
	// Get user info
	user, _ := s.repo.User.FindByID(ctx, review.UserID)
	username := ""
	if user != nil {
		username = user.Username
	}

	// Get movie info
	movie, _ := s.repo.Movie.FindByID(ctx, review.MovieID)
	movieTitle := ""
	if movie != nil {
		movieTitle = movie.Title
	}

	reviewResp := response.ReviewToResponse(review, username, movieTitle)
	return &reviewResp
}
