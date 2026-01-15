package repository

import (
	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MovieGenreRepository interface {
	// Bridge table operations
	Create(ctx context.Context, movieGenre *entity.MovieGenre) error
	DeleteByMovieID(ctx context.Context, movieID uuid.UUID) error
	DeleteByGenreID(ctx context.Context, genreID uuid.UUID) error
	FindByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.MovieGenre, error)
	FindByGenreID(ctx context.Context, genreID uuid.UUID) ([]*entity.MovieGenre, error)

	// Batch operations
	CreateBatch(ctx context.Context, movieGenres []*entity.MovieGenre) error
}

type movieGenreRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewMovieGenreRepository(db database.PgxIface, log *zap.Logger) MovieGenreRepository {
	return &movieGenreRepository{
		db:  db,
		log: log.With(zap.String("repository", "movie_genre")),
	}
}

func (r *movieGenreRepository) Create(ctx context.Context, movieGenre *entity.MovieGenre) error {
	query := `INSERT INTO movie_genres (id, movie_id, genre_id, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(ctx, query,
		movieGenre.ID,
		movieGenre.MovieID,
		movieGenre.GenreID,
		movieGenre.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create movie_genre",
			zap.Error(err),
			zap.String("movie_id", movieGenre.MovieID.String()),
			zap.String("genre_id", movieGenre.GenreID.String()),
		)
		return fmt.Errorf("failed to create movie_genre: %w", err)
	}

	return nil
}

func (r *movieGenreRepository) DeleteByMovieID(ctx context.Context, movieID uuid.UUID) error {
	query := `DELETE FROM movie_genres WHERE movie_id = $1`

	_, err := r.db.Exec(ctx, query, movieID)
	if err != nil {
		r.log.Error("Failed to delete movie_genres by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return fmt.Errorf("failed to delete movie_genres: %w", err)
	}

	return nil
}

func (r *movieGenreRepository) DeleteByGenreID(ctx context.Context, genreID uuid.UUID) error {
	query := `DELETE FROM movie_genres WHERE genre_id = $1`

	_, err := r.db.Exec(ctx, query, genreID)
	if err != nil {
		r.log.Error("Failed to delete movie_genres by genre ID",
			zap.Error(err),
			zap.String("genre_id", genreID.String()),
		)
		return fmt.Errorf("failed to delete movie_genres: %w", err)
	}

	return nil
}

func (r *movieGenreRepository) FindByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.MovieGenre, error) {
	query := `SELECT id, movie_id, genre_id, created_at FROM movie_genres WHERE movie_id = $1`

	rows, err := r.db.Query(ctx, query, movieID)
	if err != nil {
		r.log.Error("Failed to find movie_genres by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return nil, fmt.Errorf("failed to find movie_genres: %w", err)
	}
	defer rows.Close()

	var movieGenres []*entity.MovieGenre
	for rows.Next() {
		var mg entity.MovieGenre
		err := rows.Scan(
			&mg.ID,
			&mg.MovieID,
			&mg.GenreID,
			&mg.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan movie_genre row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan movie_genre: %w", err)
		}
		movieGenres = append(movieGenres, &mg)
	}

	return movieGenres, nil
}

func (r *movieGenreRepository) FindByGenreID(ctx context.Context, genreID uuid.UUID) ([]*entity.MovieGenre, error) {
	query := `SELECT id, movie_id, genre_id, created_at FROM movie_genres WHERE genre_id = $1`

	rows, err := r.db.Query(ctx, query, genreID)
	if err != nil {
		r.log.Error("Failed to find movie_genres by genre ID",
			zap.Error(err),
			zap.String("genre_id", genreID.String()),
		)
		return nil, fmt.Errorf("failed to find movie_genres: %w", err)
	}
	defer rows.Close()

	var movieGenres []*entity.MovieGenre
	for rows.Next() {
		var mg entity.MovieGenre
		err := rows.Scan(
			&mg.ID,
			&mg.MovieID,
			&mg.GenreID,
			&mg.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan movie_genre row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan movie_genre: %w", err)
		}
		movieGenres = append(movieGenres, &mg)
	}

	return movieGenres, nil
}

func (r *movieGenreRepository) CreateBatch(ctx context.Context, movieGenres []*entity.MovieGenre) error {
	if len(movieGenres) == 0 {
		return nil
	}

	// Build batch insert
	query := `INSERT INTO movie_genres (id, movie_id, genre_id, created_at) VALUES `
	args := []interface{}{}

	for i, mg := range movieGenres {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)",
			i*4+1, i*4+2, i*4+3, i*4+4)

		args = append(args, mg.ID, mg.MovieID, mg.GenreID, mg.CreatedAt)
	}

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to create batch movie_genres",
			zap.Error(err),
			zap.Int("count", len(movieGenres)),
		)
		return fmt.Errorf("failed to create batch movie_genres: %w", err)
	}

	return nil
}
