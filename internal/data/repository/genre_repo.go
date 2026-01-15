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
	// CRUD Genre
	Create(ctx context.Context, genre *entity.Genre) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error)
	FindByName(ctx context.Context, name string) (*entity.Genre, error)
	FindAll(ctx context.Context) ([]*entity.Genre, error)
	Update(ctx context.Context, genre *entity.Genre) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Find genres by movie
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

func (r *genreRepository) Create(ctx context.Context, genre *entity.Genre) error {
	query := `INSERT INTO genres (id, name, created_at) VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, query,
		genre.ID,
		genre.Name,
		genre.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create genre",
			zap.Error(err),
			zap.String("name", genre.Name),
		)
		return fmt.Errorf("failed to create genre: %w", err)
	}

	return nil
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
		return nil, fmt.Errorf("failed to find genre: %w", err)
	}

	return &genre, nil
}

func (r *genreRepository) FindByName(ctx context.Context, name string) (*entity.Genre, error) {
	query := `SELECT id, name, created_at FROM genres WHERE name = $1`

	var genre entity.Genre
	err := r.db.QueryRow(ctx, query, name).Scan(
		&genre.ID,
		&genre.Name,
		&genre.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find genre by name",
			zap.Error(err),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to find genre: %w", err)
	}

	return &genre, nil
}

func (r *genreRepository) FindAll(ctx context.Context) ([]*entity.Genre, error) {
	query := `SELECT id, name, created_at FROM genres ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.log.Error("Failed to find all genres", zap.Error(err))
		return nil, fmt.Errorf("failed to find genres: %w", err)
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
			return nil, fmt.Errorf("failed to scan genre: %w", err)
		}
		genres = append(genres, &genre)
	}

	return genres, nil
}

func (r *genreRepository) Update(ctx context.Context, genre *entity.Genre) error {
	query := `UPDATE genres SET name = $2, created_at = $3 WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		genre.ID,
		genre.Name,
		genre.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update genre",
			zap.Error(err),
			zap.String("genre_id", genre.ID.String()),
		)
		return fmt.Errorf("failed to update genre: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("genre not found")
	}

	return nil
}

func (r *genreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM genres WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete genre",
			zap.Error(err),
			zap.String("genre_id", id.String()),
		)
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("genre not found")
	}

	r.log.Info("Genre deleted", zap.String("genre_id", id.String()))
	return nil
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
		return nil, fmt.Errorf("failed to find genres: %w", err)
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
			return nil, fmt.Errorf("failed to scan genre: %w", err)
		}
		genres = append(genres, &genre)
	}

	return genres, nil
}
