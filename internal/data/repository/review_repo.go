package repository

import (
	"context"
	"fmt"

	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *entity.Review) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Review, error)
	FindByMovieID(ctx context.Context, movieID uuid.UUID, limit, offset int) ([]*entity.Review, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Review, error)
	FindByUserAndMovie(ctx context.Context, userID, movieID uuid.UUID) (*entity.Review, error)
	CountByMovieID(ctx context.Context, movieID uuid.UUID) (int64, error)
	Update(ctx context.Context, review *entity.Review) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Business queries
	GetMovieAverageRating(ctx context.Context, movieID uuid.UUID) (float64, error)
	GetMovieReviewStats(ctx context.Context, movieID uuid.UUID) (float64, int64, error) // rating, count
}

type reviewRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewReviewRepository(db database.PgxIface, log *zap.Logger) ReviewRepository {
	return &reviewRepository{
		db:  db,
		log: log.With(zap.String("repository", "review")),
	}
}

func (r *reviewRepository) Create(ctx context.Context, review *entity.Review) error {
	query := `
		INSERT INTO reviews (id, user_id, movie_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		review.ID,
		review.UserID,
		review.MovieID,
		review.Rating,
		review.Comment,
		review.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create review",
			zap.Error(err),
			zap.String("user_id", review.UserID.String()),
			zap.String("movie_id", review.MovieID.String()),
		)
		return fmt.Errorf("create review for movie %s by user %s: %w",
			review.MovieID.String(), review.UserID.String(), err)
	}

	return nil
}

func (r *reviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	query := `
		SELECT id, user_id, movie_id, rating, comment, created_at
		FROM reviews
		WHERE id = $1
	`

	var review entity.Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ID,
		&review.UserID,
		&review.MovieID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find review by ID",
			zap.Error(err),
			zap.String("review_id", id.String()),
		)
		return nil, fmt.Errorf("find review by ID %s: %w", id.String(), err)
	}

	return &review, nil
}

func (r *reviewRepository) FindByMovieID(ctx context.Context, movieID uuid.UUID, limit, offset int) ([]*entity.Review, error) {
	query := `
		SELECT id, user_id, movie_id, rating, comment, created_at
		FROM reviews
		WHERE movie_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, movieID, limit, offset)
	if err != nil {
		r.log.Error("Failed to find reviews by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
		)
		return nil, fmt.Errorf("find reviews by movie ID %s: %w", movieID.String(), err)
	}
	defer rows.Close()

	var reviews []*entity.Review
	for rows.Next() {
		var review entity.Review
		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.MovieID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan review row", zap.Error(err))
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

func (r *reviewRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Review, error) {
	query := `
		SELECT id, user_id, movie_id, rating, comment, created_at
		FROM reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.log.Error("Failed to find reviews by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
		)
		return nil, fmt.Errorf("find reviews by user ID %s: %w", userID.String(), err)
	}
	defer rows.Close()

	var reviews []*entity.Review
	for rows.Next() {
		var review entity.Review
		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.MovieID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan review row", zap.Error(err))
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

func (r *reviewRepository) FindByUserAndMovie(ctx context.Context, userID, movieID uuid.UUID) (*entity.Review, error) {
	query := `
		SELECT id, user_id, movie_id, rating, comment, created_at
		FROM reviews
		WHERE user_id = $1 AND movie_id = $2
		LIMIT 1
	`

	var review entity.Review
	err := r.db.QueryRow(ctx, query, userID, movieID).Scan(
		&review.ID,
		&review.UserID,
		&review.MovieID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find review by user and movie",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("movie_id", movieID.String()),
		)
		return nil, fmt.Errorf("find review by user %s and movie %s: %w",
			userID.String(), movieID.String(), err)
	}

	return &review, nil
}

func (r *reviewRepository) CountByMovieID(ctx context.Context, movieID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE movie_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, movieID).Scan(&count)
	if err != nil {
		r.log.Error("Failed to count reviews by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return 0, fmt.Errorf("count reviews by movie ID %s: %w", movieID.String(), err)
	}

	return count, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *entity.Review) error {
	query := `
		UPDATE reviews
		SET rating = $2, comment = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		review.ID,
		review.Rating,
		review.Comment,
	)

	if err != nil {
		r.log.Error("Failed to update review",
			zap.Error(err),
			zap.String("review_id", review.ID.String()),
		)
		return fmt.Errorf("update review %s: %w", review.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review %s not found", review.ID.String())
	}

	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reviews WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete review",
			zap.Error(err),
			zap.String("review_id", id.String()),
		)
		return fmt.Errorf("delete review %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("review %s not found", id.String())
	}

	r.log.Info("Review deleted", zap.String("review_id", id.String()))
	return nil
}

func (r *reviewRepository) GetMovieAverageRating(ctx context.Context, movieID uuid.UUID) (float64, error) {
	query := `SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE movie_id = $1`

	var avgRating float64
	err := r.db.QueryRow(ctx, query, movieID).Scan(&avgRating)
	if err != nil {
		r.log.Error("Failed to get movie average rating",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return 0, fmt.Errorf("get movie average rating for %s: %w", movieID.String(), err)
	}

	return avgRating, nil
}

func (r *reviewRepository) GetMovieReviewStats(ctx context.Context, movieID uuid.UUID) (float64, int64, error) {
	query := `
		SELECT 
			COALESCE(AVG(rating), 0) as avg_rating,
			COUNT(*) as review_count
		FROM reviews 
		WHERE movie_id = $1
	`

	var avgRating float64
	var reviewCount int64
	err := r.db.QueryRow(ctx, query, movieID).Scan(&avgRating, &reviewCount)
	if err != nil {
		r.log.Error("Failed to get movie review stats",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return 0, 0, fmt.Errorf("get movie review stats for %s: %w", movieID.String(), err)
	}

	return avgRating, reviewCount, nil
}
