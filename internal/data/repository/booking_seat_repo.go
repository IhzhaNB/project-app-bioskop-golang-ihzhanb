package repository

import (
	"context"
	"fmt"

	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BookingSeatRepository interface {
	Create(ctx context.Context, bookingSeat *entity.BookingSeat) error
	FindByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*entity.BookingSeat, error)
	FindBySeatID(ctx context.Context, seatID uuid.UUID) ([]*entity.BookingSeat, error)
	DeleteByBookingID(ctx context.Context, bookingID uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, bookingSeats []*entity.BookingSeat) error

	// Business queries
	FindBookedSeatsBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]uuid.UUID, error)
}

type bookingSeatRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewBookingSeatRepository(db database.PgxIface, log *zap.Logger) BookingSeatRepository {
	return &bookingSeatRepository{
		db:  db,
		log: log.With(zap.String("repository", "booking_seat")),
	}
}

func (r *bookingSeatRepository) Create(ctx context.Context, bookingSeat *entity.BookingSeat) error {
	query := `
		INSERT INTO booking_seats (id, booking_id, seat_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query,
		bookingSeat.ID,
		bookingSeat.BookingID,
		bookingSeat.SeatID,
		bookingSeat.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create booking seat",
			zap.Error(err),
			zap.String("booking_id", bookingSeat.BookingID.String()),
			zap.String("seat_id", bookingSeat.SeatID.String()),
		)
		return fmt.Errorf("create booking seat for booking %s seat %s: %w",
			bookingSeat.BookingID.String(), bookingSeat.SeatID.String(), err)
	}

	return nil
}

func (r *bookingSeatRepository) CreateBatch(ctx context.Context, bookingSeats []*entity.BookingSeat) error {
	if len(bookingSeats) == 0 {
		return nil
	}

	// Simple loop untuk sekarang
	for _, bs := range bookingSeats {
		if err := r.Create(ctx, bs); err != nil {
			return err
		}
	}

	return nil
}

func (r *bookingSeatRepository) FindByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*entity.BookingSeat, error) {
	query := `
		SELECT id, booking_id, seat_id, created_at
		FROM booking_seats
		WHERE booking_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, bookingID)
	if err != nil {
		r.log.Error("Failed to find booking seats by booking ID",
			zap.Error(err),
			zap.String("booking_id", bookingID.String()),
		)
		return nil, fmt.Errorf("find booking seats by booking ID %s: %w", bookingID.String(), err)
	}
	defer rows.Close()

	var bookingSeats []*entity.BookingSeat
	for rows.Next() {
		var bs entity.BookingSeat
		err := rows.Scan(
			&bs.ID,
			&bs.BookingID,
			&bs.SeatID,
			&bs.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan booking seat row", zap.Error(err))
			return nil, fmt.Errorf("scan booking seat row: %w", err)
		}
		bookingSeats = append(bookingSeats, &bs)
	}

	return bookingSeats, nil
}

func (r *bookingSeatRepository) FindBySeatID(ctx context.Context, seatID uuid.UUID) ([]*entity.BookingSeat, error) {
	query := `
		SELECT id, booking_id, seat_id, created_at
		FROM booking_seats
		WHERE seat_id = $1
	`

	rows, err := r.db.Query(ctx, query, seatID)
	if err != nil {
		r.log.Error("Failed to find booking seats by seat ID",
			zap.Error(err),
			zap.String("seat_id", seatID.String()),
		)
		return nil, fmt.Errorf("find booking seats by seat ID %s: %w", seatID.String(), err)
	}
	defer rows.Close()

	var bookingSeats []*entity.BookingSeat
	for rows.Next() {
		var bs entity.BookingSeat
		err := rows.Scan(
			&bs.ID,
			&bs.BookingID,
			&bs.SeatID,
			&bs.CreatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan booking seat row", zap.Error(err))
			return nil, fmt.Errorf("scan booking seat row: %w", err)
		}
		bookingSeats = append(bookingSeats, &bs)
	}

	return bookingSeats, nil
}

func (r *bookingSeatRepository) DeleteByBookingID(ctx context.Context, bookingID uuid.UUID) error {
	query := `DELETE FROM booking_seats WHERE booking_id = $1`

	_, err := r.db.Exec(ctx, query, bookingID)
	if err != nil {
		r.log.Error("Failed to delete booking seats by booking ID",
			zap.Error(err),
			zap.String("booking_id", bookingID.String()),
		)
		return fmt.Errorf("delete booking seats by booking ID %s: %w", bookingID.String(), err)
	}

	return nil
}

func (r *bookingSeatRepository) FindBookedSeatsBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT bs.seat_id
		FROM booking_seats bs
		INNER JOIN bookings b ON bs.booking_id = b.id
		WHERE b.schedule_id = $1 AND b.status IN ('confirmed', 'pending')
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		r.log.Error("Failed to find booked seats by schedule",
			zap.Error(err),
			zap.String("schedule_id", scheduleID.String()),
		)
		return nil, fmt.Errorf("find booked seats by schedule %s: %w", scheduleID.String(), err)
	}
	defer rows.Close()

	var seatIDs []uuid.UUID
	for rows.Next() {
		var seatID uuid.UUID
		err := rows.Scan(&seatID)
		if err != nil {
			r.log.Error("Failed to scan seat ID row", zap.Error(err))
			return nil, fmt.Errorf("scan seat ID row: %w", err)
		}
		seatIDs = append(seatIDs, seatID)
	}

	return seatIDs, nil
}
