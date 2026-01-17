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
	DeleteByMovieID(ctx context.Context, movieID uuid.UUID) error
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

func (r *movieGenreRepository) DeleteByMovieID(ctx context.Context, movieID uuid.UUID) error {
	query := `DELETE FROM movie_genres WHERE movie_id = $1`

	_, err := r.db.Exec(ctx, query, movieID)
	if err != nil {
		r.log.Error("Failed to delete movie_genres by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return fmt.Errorf("delete movie_genres by movie id: %w", err)
	}

	return nil
}

func (r *movieGenreRepository) CreateBatch(ctx context.Context, movieGenres []*entity.MovieGenre) error {
	if len(movieGenres) == 0 {
		return nil
	}

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
		return fmt.Errorf("create batch movie_genres: %w", err)
	}

	return nil
}
