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
	Create(ctx context.Context, cinema *entity.Cinema) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Cinema, error)
	FindAll(ctx context.Context, limit, offset int, cityFilter *string) ([]*entity.Cinema, error)
	CountAll(ctx context.Context, cityFilter *string) (int64, error)
	Update(ctx context.Context, cinema *entity.Cinema) error
	Delete(ctx context.Context, id uuid.UUID) error
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
			zap.String("city", cinema.City),
		)
		return fmt.Errorf("create cinema %s: %w", cinema.Name, err)
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
		return nil, fmt.Errorf("find cinema by ID %s: %w", id.String(), err)
	}

	return &cinema, nil
}

func (r *cinemaRepository) FindAll(ctx context.Context, limit, offset int, cityFilter *string) ([]*entity.Cinema, error) {
	// Build query dengan optional filter
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT id, name, location, city, created_at, updated_at
		FROM cinemas
		WHERE deleted_at IS NULL
	`)

	args := []interface{}{}
	argCount := 1

	if cityFilter != nil && *cityFilter != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND city ILIKE $%d", argCount))
		args = append(args, "%"+*cityFilter+"%")
		argCount++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY city, name LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		r.log.Error("Failed to find all cinemas",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
			zap.Stringp("city_filter", cityFilter),
		)
		return nil, fmt.Errorf("find all cinemas limit %d offset %d: %w", limit, offset, err)
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
			return nil, fmt.Errorf("scan cinema row: %w", err)
		}
		cinemas = append(cinemas, &cinema)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("iterate cinema rows: %w", err)
	}

	return cinemas, nil
}

func (r *cinemaRepository) CountAll(ctx context.Context, cityFilter *string) (int64, error) {
	// Build count query
	query := `SELECT COUNT(*) FROM cinemas WHERE deleted_at IS NULL`
	args := []interface{}{}

	if cityFilter != nil && *cityFilter != "" {
		query += " AND city ILIKE $1"
		args = append(args, "%"+*cityFilter+"%")
	}

	var total int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		r.log.Error("Failed to count cinemas",
			zap.Error(err),
			zap.Stringp("city_filter", cityFilter),
		)
		return 0, fmt.Errorf("count all cinemas: %w", err)
	}

	return total, nil
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
		return fmt.Errorf("update cinema %s: %w", cinema.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cinema %s not found or already deleted", cinema.ID.String())
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
		return fmt.Errorf("delete cinema %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cinema %s not found", id.String())
	}

	r.log.Info("Cinema deleted", zap.String("cinema_id", id.String()))
	return nil
}
