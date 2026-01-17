package repository

import (
	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type GenreRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error)
	FindByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.Genre, error)
}

type genreRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewGenreRepository(db database.PgxIface, log *zap.Logger) GenreRepository {
	return &genreRepository{
		db:  db,
		log: log.With(zap.String("repository", "genre")),
	}
}

func (r *genreRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error) {
	query := `SELECT id, name, created_at FROM genres WHERE id = $1`

	var genre entity.Genre
	err := r.db.QueryRow(ctx, query, id).Scan(
		&genre.ID,
		&genre.Name,
		&genre.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find genre by ID",
			zap.Error(err),
			zap.String("genre_id", id.String()),
		)
		return nil, fmt.Errorf("find genre by id: %w", err)
	}

	return &genre, nil
}

func (r *genreRepository) FindByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.Genre, error) {
	query := `
		SELECT g.id, g.name, g.created_at
		FROM genres g
		INNER JOIN movie_genres mg ON g.id = mg.genre_id
		WHERE mg.movie_id = $1
		ORDER BY g.name
	`

	rows, err := r.db.Query(ctx, query, movieID)
	if err != nil {
		r.log.Error("Failed to find genres by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return nil, fmt.Errorf("find genres by movie id: %w", err)
	}
	defer rows.Close()

	var genres []*entity.Genre
	for rows.Next() {
		var genre entity.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan genre row", zap.Error(err))
			return nil, fmt.Errorf("scan genre row: %w", err)
		}
		genres = append(genres, &genre)
	}

	return genres, nil
}
