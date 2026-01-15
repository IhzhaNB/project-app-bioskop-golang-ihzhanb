package repository

import (
	"context"
	"fmt"
	"strings"

	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type CinemaRepository interface {
	// CRUD Cinema
	Create(ctx context.Context, cinema *entity.Cinema) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Cinema, error)
	FindAll(ctx context.Context, page, limit int, city *string) ([]*entity.Cinema, error)
	CountAll(ctx context.Context, city *string) (int64, error)
	Update(ctx context.Context, cinema *entity.Cinema) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Find by city
	FindByCity(ctx context.Context, city string) ([]*entity.Cinema, error)
}

type cinemaRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewCinemaRepository(db database.PgxIface, log *zap.Logger) CinemaRepository {
	return &cinemaRepository{
		db:  db,
		log: log.With(zap.String("repository", "cinema")),
	}
}

func (r *cinemaRepository) Create(ctx context.Context, cinema *entity.Cinema) error {
	query := `
		INSERT INTO cinemas (id, name, location, city, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		cinema.ID,
		cinema.Name,
		cinema.Location,
		cinema.City,
		cinema.CreatedAt,
		cinema.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create cinema",
			zap.Error(err),
			zap.String("name", cinema.Name),
		)
		return fmt.Errorf("failed to create cinema: %w", err)
	}

	return nil
}

func (r *cinemaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Cinema, error) {
	query := `
		SELECT id, name, location, city, created_at, updated_at, deleted_at
		FROM cinemas
		WHERE id = $1 AND deleted_at IS NULL
	`

	var cinema entity.Cinema
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cinema.ID,
		&cinema.Name,
		&cinema.Location,
		&cinema.City,
		&cinema.CreatedAt,
		&cinema.UpdatedAt,
		&cinema.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find cinema by ID",
			zap.Error(err),
			zap.String("cinema_id", id.String()),
		)
		return nil, fmt.Errorf("failed to find cinema: %w", err)
	}

	return &cinema, nil
}

func (r *cinemaRepository) FindAll(ctx context.Context, page, limit int, city *string) ([]*entity.Cinema, error) {
	// Calculate offset
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Build query
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT id, name, location, city, created_at, updated_at
		FROM cinemas
		WHERE deleted_at IS NULL
	`)

	args := []interface{}{}
	argCount := 1

	if city != nil && *city != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND city = $%d", argCount))
		args = append(args, *city)
		argCount++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY city, name LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		r.log.Error("Failed to find all cinemas",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.Stringp("city", city),
		)
		return nil, fmt.Errorf("failed to find cinemas: %w", err)
	}
	defer rows.Close()

	var cinemas []*entity.Cinema
	for rows.Next() {
		var cinema entity.Cinema
		err := rows.Scan(
			&cinema.ID,
			&cinema.Name,
			&cinema.Location,
			&cinema.City,
			&cinema.CreatedAt,
			&cinema.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan cinema row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cinema: %w", err)
		}
		cinemas = append(cinemas, &cinema)
	}

	return cinemas, nil
}

func (r *cinemaRepository) CountAll(ctx context.Context, city *string) (int64, error) {
	query := `SELECT COUNT(*) FROM cinemas WHERE deleted_at IS NULL`
	args := []interface{}{}

	if city != nil && *city != "" {
		query += " AND city = $1"
		args = append(args, *city)
	}

	var total int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		r.log.Error("Failed to count cinemas",
			zap.Error(err),
			zap.Stringp("city", city),
		)
		return 0, fmt.Errorf("failed to count cinemas: %w", err)
	}

	return total, nil
}

func (r *cinemaRepository) FindByCity(ctx context.Context, city string) ([]*entity.Cinema, error) {
	query := `
		SELECT id, name, location, city, created_at, updated_at
		FROM cinemas
		WHERE city = $1 AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, city)
	if err != nil {
		r.log.Error("Failed to find cinemas by city",
			zap.Error(err),
			zap.String("city", city),
		)
		return nil, fmt.Errorf("failed to find cinemas: %w", err)
	}
	defer rows.Close()

	var cinemas []*entity.Cinema
	for rows.Next() {
		var cinema entity.Cinema
		err := rows.Scan(
			&cinema.ID,
			&cinema.Name,
			&cinema.Location,
			&cinema.City,
			&cinema.CreatedAt,
			&cinema.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan cinema row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cinema: %w", err)
		}
		cinemas = append(cinemas, &cinema)
	}

	return cinemas, nil
}

func (r *cinemaRepository) Update(ctx context.Context, cinema *entity.Cinema) error {
	query := `
		UPDATE cinemas
		SET name = $2, location = $3, city = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		cinema.ID,
		cinema.Name,
		cinema.Location,
		cinema.City,
		cinema.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update cinema",
			zap.Error(err),
			zap.String("cinema_id", cinema.ID.String()),
		)
		return fmt.Errorf("failed to update cinema: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cinema not found or already deleted")
	}

	return nil
}

func (r *cinemaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE cinemas SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete cinema",
			zap.Error(err),
			zap.String("cinema_id", id.String()),
		)
		return fmt.Errorf("failed to delete cinema: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cinema not found or already deleted")
	}

	r.log.Info("Cinema soft deleted", zap.String("cinema_id", id.String()))
	return nil
}
