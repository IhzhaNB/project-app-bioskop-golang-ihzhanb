package repository

import (
	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type MovieRepository interface {
	// CRUD Movie
	Create(ctx context.Context, movie *entity.Movie) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Movie, error)
	Update(ctx context.Context, movie *entity.Movie) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context, offset, limit int, releaseStatus *string) ([]*entity.Movie, error)
	CountAll(ctx context.Context, releaseStatus *string) (int64, error)

	// Update rating
	UpdateRating(ctx context.Context, movieID uuid.UUID, newRating float64) error
}

type movieRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewMovieRepository(db database.PgxIface, log *zap.Logger) MovieRepository {
	return &movieRepository{
		db:  db,
		log: log.With(zap.String("repository", "movie")),
	}
}

func (r *movieRepository) Create(ctx context.Context, movie *entity.Movie) error {
	query := `
		INSERT INTO movies (id, title, description, poster_url, rating, 
		                   release_date, duration_in_minutes, release_status,
		                   created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		movie.ID,
		movie.Title,
		movie.Description,
		movie.PosterURL,
		movie.Rating,
		movie.ReleaseDate,
		movie.DurationInMinutes,
		movie.ReleaseStatus,
		movie.CreatedAt,
		movie.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create movie",
			zap.Error(err),
			zap.String("title", movie.Title),
		)
		return fmt.Errorf("failed to create movie: %w", err)
	}

	return nil
}

func (r *movieRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Movie, error) {
	query := `
		SELECT id, title, description, poster_url, rating, release_date,
		       duration_in_minutes, release_status, created_at, updated_at, deleted_at
		FROM movies
		WHERE id = $1 AND deleted_at IS NULL
	`

	var movie entity.Movie
	err := r.db.QueryRow(ctx, query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.PosterURL,
		&movie.Rating,
		&movie.ReleaseDate,
		&movie.DurationInMinutes,
		&movie.ReleaseStatus,
		&movie.CreatedAt,
		&movie.UpdatedAt,
		&movie.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find movie by ID",
			zap.Error(err),
			zap.String("movie_id", id.String()),
		)
		return nil, fmt.Errorf("failed to find movie: %w", err)
	}

	return &movie, nil
}

func (r *movieRepository) FindAll(ctx context.Context, offset, limit int, releaseStatus *string) ([]*entity.Movie, error) {
	// Build query dengan optional filter
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT id, title, description, poster_url, rating, release_date,
		       duration_in_minutes, release_status, created_at, updated_at
		FROM movies
		WHERE deleted_at IS NULL
	`)

	args := []interface{}{}
	argCount := 1

	if releaseStatus != nil && *releaseStatus != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND release_status = $%d", argCount))
		args = append(args, *releaseStatus)
		argCount++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY release_date DESC LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		r.log.Error("Failed to find all movies",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
			zap.Stringp("release_status", releaseStatus),
		)
		return nil, fmt.Errorf("failed to find movies: %w", err)
	}
	defer rows.Close()

	var movies []*entity.Movie
	for rows.Next() {
		var movie entity.Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.PosterURL,
			&movie.Rating,
			&movie.ReleaseDate,
			&movie.DurationInMinutes,
			&movie.ReleaseStatus,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan movie row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan movie: %w", err)
		}
		movies = append(movies, &movie)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	r.log.Debug("Movies found",
		zap.Int("count", len(movies)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return movies, nil
}

func (r *movieRepository) CountAll(ctx context.Context, releaseStatus *string) (int64, error) {
	// Build count query
	query := `SELECT COUNT(*) FROM movies WHERE deleted_at IS NULL`
	args := []interface{}{}

	if releaseStatus != nil && *releaseStatus != "" {
		query += " AND release_status = $1"
		args = append(args, *releaseStatus)
	}

	var total int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		r.log.Error("Failed to count movies",
			zap.Error(err),
			zap.Stringp("release_status", releaseStatus),
		)
		return 0, fmt.Errorf("failed to count movies: %w", err)
	}

	r.log.Debug("Movies counted",
		zap.Int64("total", total),
		zap.Stringp("release_status", releaseStatus),
	)

	return total, nil
}

func (r *movieRepository) Update(ctx context.Context, movie *entity.Movie) error {
	query := `
		UPDATE movies
		SET title = $2, description = $3, poster_url = $4, rating = $5,
		    release_date = $6, duration_in_minutes = $7, release_status = $8,
		    updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		movie.ID,
		movie.Title,
		movie.Description,
		movie.PosterURL,
		movie.Rating,
		movie.ReleaseDate,
		movie.DurationInMinutes,
		movie.ReleaseStatus,
		movie.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update movie",
			zap.Error(err),
			zap.String("movie_id", movie.ID.String()),
		)
		return fmt.Errorf("failed to update movie: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("movie not found or already deleted")
	}

	return nil
}

func (r *movieRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE movies SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete movie",
			zap.Error(err),
			zap.String("movie_id", id.String()),
		)
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("movie not found or already deleted")
	}

	r.log.Info("Movie soft deleted", zap.String("movie_id", id.String()))
	return nil
}

func (r *movieRepository) UpdateRating(ctx context.Context, movieID uuid.UUID, newRating float64) error {
	query := `UPDATE movies SET rating = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, movieID, newRating)
	if err != nil {
		r.log.Error("Failed to update movie rating",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
			zap.Float64("new_rating", newRating),
		)
		return fmt.Errorf("failed to update rating: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("movie not found")
	}

	return nil
}
