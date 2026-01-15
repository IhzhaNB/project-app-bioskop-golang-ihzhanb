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
	// CRUD Seat
	Create(ctx context.Context, seat *entity.Seat) error
	CreateBatch(ctx context.Context, seats []*entity.Seat) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Seat, error)
	FindByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error)
	FindAvailableSeats(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error)
	Update(ctx context.Context, seat *entity.Seat) error
	UpdateAvailability(ctx context.Context, seatID uuid.UUID, isAvailable bool) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Special queries for booking
	FindSeatsForBooking(ctx context.Context, hallID uuid.UUID, seatIDs []uuid.UUID) ([]*entity.Seat, error)
	UpdateSeatsAvailability(ctx context.Context, seatIDs []uuid.UUID, isAvailable bool) error
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
		return fmt.Errorf("failed to create seat: %w", err)
	}

	return nil
}

func (r *seatRepository) CreateBatch(ctx context.Context, seats []*entity.Seat) error {
	if len(seats) == 0 {
		return nil
	}

	// Build batch insert
	query := `INSERT INTO seats (id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at) VALUES `
	args := []interface{}{}

	for i, seat := range seats {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8)

		args = append(args,
			seat.ID,
			seat.HallID,
			seat.SeatNumber,
			seat.SeatRow,
			seat.SeatColumn,
			seat.IsAvailable,
			seat.CreatedAt,
			seat.UpdatedAt,
		)
	}

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to create batch seats",
			zap.Error(err),
			zap.Int("count", len(seats)),
		)
		return fmt.Errorf("failed to create batch seats: %w", err)
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
		return nil, fmt.Errorf("failed to find seat: %w", err)
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
		return nil, fmt.Errorf("failed to find seats: %w", err)
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
			return nil, fmt.Errorf("failed to scan seat: %w", err)
		}
		seats = append(seats, &seat)
	}

	return seats, nil
}

func (r *seatRepository) FindAvailableSeats(ctx context.Context, hallID uuid.UUID) ([]*entity.Seat, error) {
	query := `
		SELECT id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at
		FROM seats
		WHERE hall_id = $1 AND is_available = true AND deleted_at IS NULL
		ORDER BY seat_row, seat_column
	`

	rows, err := r.db.Query(ctx, query, hallID)
	if err != nil {
		r.log.Error("Failed to find available seats",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
		)
		return nil, fmt.Errorf("failed to find available seats: %w", err)
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
			return nil, fmt.Errorf("failed to scan seat: %w", err)
		}
		seats = append(seats, &seat)
	}

	return seats, nil
}

func (r *seatRepository) FindSeatsForBooking(ctx context.Context, hallID uuid.UUID, seatIDs []uuid.UUID) ([]*entity.Seat, error) {
	if len(seatIDs) == 0 {
		return []*entity.Seat{}, nil
	}

	// Build query dengan IN clause
	query := `
		SELECT id, hall_id, seat_number, seat_row, seat_column, is_available, created_at, updated_at
		FROM seats
		WHERE hall_id = $1 AND id = ANY($2) AND deleted_at IS NULL
		ORDER BY seat_row, seat_column
	`

	rows, err := r.db.Query(ctx, query, hallID, seatIDs)
	if err != nil {
		r.log.Error("Failed to find seats for booking",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
			zap.Int("seat_count", len(seatIDs)),
		)
		return nil, fmt.Errorf("failed to find seats: %w", err)
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
			return nil, fmt.Errorf("failed to scan seat: %w", err)
		}
		seats = append(seats, &seat)
	}

	return seats, nil
}

func (r *seatRepository) Update(ctx context.Context, seat *entity.Seat) error {
	query := `
		UPDATE seats
		SET seat_number = $2, seat_row = $3, seat_column = $4, is_available = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		seat.ID,
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
		return fmt.Errorf("failed to update seat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("seat not found or already deleted")
	}

	return nil
}

func (r *seatRepository) UpdateAvailability(ctx context.Context, seatID uuid.UUID, isAvailable bool) error {
	query := `UPDATE seats SET is_available = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, seatID, isAvailable)
	if err != nil {
		r.log.Error("Failed to update seat availability",
			zap.Error(err),
			zap.String("seat_id", seatID.String()),
			zap.Bool("is_available", isAvailable),
		)
		return fmt.Errorf("failed to update seat availability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("seat not found")
	}

	return nil
}

func (r *seatRepository) UpdateSeatsAvailability(ctx context.Context, seatIDs []uuid.UUID, isAvailable bool) error {
	if len(seatIDs) == 0 {
		return nil
	}

	query := `UPDATE seats SET is_available = $2, updated_at = NOW() WHERE id = ANY($1) AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, seatIDs, isAvailable)
	if err != nil {
		r.log.Error("Failed to update seats availability",
			zap.Error(err),
			zap.Int("seat_count", len(seatIDs)),
			zap.Bool("is_available", isAvailable),
		)
		return fmt.Errorf("failed to update seats availability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no seats updated")
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
		return fmt.Errorf("failed to delete seat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("seat not found or already deleted")
	}

	r.log.Info("Seat soft deleted", zap.String("seat_id", id.String()))
	return nil
}
