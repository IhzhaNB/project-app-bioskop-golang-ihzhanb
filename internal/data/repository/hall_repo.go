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

type HallRepository interface {
	Create(ctx context.Context, hall *entity.Hall) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Hall, error)
	FindByCinemaID(ctx context.Context, cinemaID uuid.UUID) ([]*entity.Hall, error)
	Update(ctx context.Context, hall *entity.Hall) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type hallRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewHallRepository(db database.PgxIface, log *zap.Logger) HallRepository {
	return &hallRepository{
		db:  db,
		log: log.With(zap.String("repository", "hall")),
	}
}

func (r *hallRepository) Create(ctx context.Context, hall *entity.Hall) error {
	query := `
		INSERT INTO halls (id, cinema_id, hall_number, total_seats, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		hall.ID,
		hall.CinemaID,
		hall.HallNumber,
		hall.TotalSeats,
		hall.CreatedAt,
		hall.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create hall",
			zap.Error(err),
			zap.String("cinema_id", hall.CinemaID.String()),
			zap.Int("hall_number", hall.HallNumber),
		)
		return fmt.Errorf("create hall %d in cinema %s: %w", hall.HallNumber, hall.CinemaID.String(), err)
	}

	return nil
}

func (r *hallRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Hall, error) {
	query := `
		SELECT id, cinema_id, hall_number, total_seats, created_at, updated_at, deleted_at
		FROM halls
		WHERE id = $1 AND deleted_at IS NULL
	`

	var hall entity.Hall
	err := r.db.QueryRow(ctx, query, id).Scan(
		&hall.ID,
		&hall.CinemaID,
		&hall.HallNumber,
		&hall.TotalSeats,
		&hall.CreatedAt,
		&hall.UpdatedAt,
		&hall.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find hall by ID",
			zap.Error(err),
			zap.String("hall_id", id.String()),
		)
		return nil, fmt.Errorf("find hall by ID %s: %w", id.String(), err)
	}

	return &hall, nil
}

func (r *hallRepository) FindByCinemaID(ctx context.Context, cinemaID uuid.UUID) ([]*entity.Hall, error) {
	query := `
		SELECT id, cinema_id, hall_number, total_seats, created_at, updated_at
		FROM halls
		WHERE cinema_id = $1 AND deleted_at IS NULL
		ORDER BY hall_number
	`

	rows, err := r.db.Query(ctx, query, cinemaID)
	if err != nil {
		r.log.Error("Failed to find halls by cinema ID",
			zap.Error(err),
			zap.String("cinema_id", cinemaID.String()),
		)
		return nil, fmt.Errorf("find halls by cinema ID %s: %w", cinemaID.String(), err)
	}
	defer rows.Close()

	var halls []*entity.Hall
	for rows.Next() {
		var hall entity.Hall
		err := rows.Scan(
			&hall.ID,
			&hall.CinemaID,
			&hall.HallNumber,
			&hall.TotalSeats,
			&hall.CreatedAt,
			&hall.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan hall row", zap.Error(err))
			return nil, fmt.Errorf("scan hall row: %w", err)
		}
		halls = append(halls, &hall)
	}

	return halls, nil
}

func (r *hallRepository) Update(ctx context.Context, hall *entity.Hall) error {
	query := `
		UPDATE halls
		SET cinema_id = $2, hall_number = $3, total_seats = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		hall.ID,
		hall.CinemaID,
		hall.HallNumber,
		hall.TotalSeats,
		hall.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update hall",
			zap.Error(err),
			zap.String("hall_id", hall.ID.String()),
		)
		return fmt.Errorf("update hall %s: %w", hall.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("hall %s not found or already deleted", hall.ID.String())
	}

	return nil
}

func (r *hallRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE halls SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete hall",
			zap.Error(err),
			zap.String("hall_id", id.String()),
		)
		return fmt.Errorf("delete hall %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("hall %s not found", id.String())
	}

	r.log.Info("Hall deleted", zap.String("hall_id", id.String()))
	return nil
}
