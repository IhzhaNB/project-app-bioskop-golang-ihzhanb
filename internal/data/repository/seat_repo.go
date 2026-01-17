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

type SeatRepository interface {
	Create(ctx context.Context, seat *entity.Seat) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Seat, error)
	FindByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error)
	FindAvailableByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error)
	Update(ctx context.Context, seat *entity.Seat) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, seats []*entity.Seat) error
}

type seatRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewSeatRepository(db database.PgxIface, log *zap.Logger) SeatRepository {
	return &seatRepository{
		db:  db,
		log: log.With(zap.String("repository", "seat")),
	}
}

func (r *seatRepository) Create(ctx context.Context, seat *entity.Seat) error {
	query := `
		INSERT INTO seats (id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		seat.ID,
		seat.HallID,
		seat.SeatNumber,
		seat.SeatRow,
		seat.SeatColumn,
		seat.IsAvailable,
		seat.CreatedAt,
		seat.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create seat",
			zap.Error(err),
			zap.String("hall_id", seat.HallID.String()),
			zap.String("seat_number", seat.SeatNumber),
		)
		return fmt.Errorf("create seat %s in hall %s: %w", seat.SeatNumber, seat.HallID.String(), err)
	}

	return nil
}

func (r *seatRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Seat, error) {
	query := `
		SELECT id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at, deleted_at
		FROM seats
		WHERE id = $1 AND deleted_at IS NULL
	`

	var seat entity.Seat
	err := r.db.QueryRow(ctx, query, id).Scan(
		&seat.ID,
		&seat.HallID,
		&seat.SeatNumber,
		&seat.SeatRow,
		&seat.SeatColumn,
		&seat.IsAvailable,
		&seat.CreatedAt,
		&seat.UpdatedAt,
		&seat.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find seat by ID",
			zap.Error(err),
			zap.String("seat_id", id.String()),
		)
		return nil, fmt.Errorf("find seat by ID %s: %w", id.String(), err)
	}

	return &seat, nil
}

func (r *seatRepository) FindByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error) {
	query := `
		SELECT id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at
		FROM seats
		WHERE hall_id = $1 AND deleted_at IS NULL
		ORDER BY seat_row, seat_column
	`

	rows, err := r.db.Query(ctx, query, hallID)
	if err != nil {
		r.log.Error("Failed to find seats by hall ID",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
		)
		return nil, fmt.Errorf("find seats by hall ID %s: %w", hallID.String(), err)
	}
	defer rows.Close()

	var seats []*entity.Seat
	for rows.Next() {
		var seat entity.Seat
		err := rows.Scan(
			&seat.ID,
			&seat.HallID,
			&seat.SeatNumber,
			&seat.SeatRow,
			&seat.SeatColumn,
			&seat.IsAvailable,
			&seat.CreatedAt,
			&seat.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan seat row", zap.Error(err))
			return nil, fmt.Errorf("scan seat row: %w", err)
		}
		seats = append(seats, &seat)
	}

	return seats, nil
}

func (r *seatRepository) FindAvailableByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error) {
	query := `
		SELECT id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at
		FROM seats
		WHERE hall_id = $1 AND is_available = true AND deleted_at IS NULL
		ORDER BY seat_row, seat_column
	`

	rows, err := r.db.Query(ctx, query, hallID)
	if err != nil {
		r.log.Error("Failed to find available seats by hall ID",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
		)
		return nil, fmt.Errorf("find available seats by hall ID %s: %w", hallID.String(), err)
	}
	defer rows.Close()

	var seats []*entity.Seat
	for rows.Next() {
		var seat entity.Seat
		err := rows.Scan(
			&seat.ID,
			&seat.HallID,
			&seat.SeatNumber,
			&seat.SeatRow,
			&seat.SeatColumn,
			&seat.IsAvailable,
			&seat.CreatedAt,
			&seat.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan seat row", zap.Error(err))
			return nil, fmt.Errorf("scan seat row: %w", err)
		}
		seats = append(seats, &seat)
	}

	return seats, nil
}

func (r *seatRepository) Update(ctx context.Context, seat *entity.Seat) error {
	query := `
		UPDATE seats
		SET hall_id = $2, seat_number = $3, seat_row = $4, seat_column = $5, 
		    is_available = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		seat.ID,
		seat.HallID,
		seat.SeatNumber,
		seat.SeatRow,
		seat.SeatColumn,
		seat.IsAvailable,
		seat.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update seat",
			zap.Error(err),
			zap.String("seat_id", seat.ID.String()),
		)
		return fmt.Errorf("update seat %s: %w", seat.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("seat %s not found or already deleted", seat.ID.String())
	}

	return nil
}

func (r *seatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE seats SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete seat",
			zap.Error(err),
			zap.String("seat_id", id.String()),
		)
		return fmt.Errorf("delete seat %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("seat %s not found", id.String())
	}

	r.log.Info("Seat deleted", zap.String("seat_id", id.String()))
	return nil
}

func (r *seatRepository) CreateBatch(ctx context.Context, seats []*entity.Seat) error {
	if len(seats) == 0 {
		return nil
	}

	// Simple loop untuk sekarang, bisa optimize dengan batch insert nanti
	for _, seat := range seats {
		if err := r.Create(ctx, seat); err != nil {
			return err
		}
	}

	return nil
}
