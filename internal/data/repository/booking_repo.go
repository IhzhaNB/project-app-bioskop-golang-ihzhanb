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

type BookingRepository interface {
	Create(ctx context.Context, booking *entity.Booking) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	FindByOrderID(ctx context.Context, orderID string) (*entity.Booking, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Booking, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Update(ctx context.Context, booking *entity.Booking) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Business queries
	FindByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*entity.Booking, error)
	FindConfirmedByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*entity.Booking, error)
	UpdateStatus(ctx context.Context, bookingID uuid.UUID, status entity.BookingStatus) error
}

type bookingRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewBookingRepository(db database.PgxIface, log *zap.Logger) BookingRepository {
	return &bookingRepository{
		db:  db,
		log: log.With(zap.String("repository", "booking")),
	}
}

func (r *bookingRepository) Create(ctx context.Context, booking *entity.Booking) error {
	query := `
		INSERT INTO bookings (id, order_id, user_id, schedule_id, total_seats, total_price, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		booking.ID,
		booking.OrderID,
		booking.UserID,
		booking.ScheduleID,
		booking.TotalSeats,
		booking.TotalPrice,
		booking.Status,
		booking.CreatedAt,
		booking.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create booking",
			zap.Error(err),
			zap.String("order_id", booking.OrderID),
			zap.String("user_id", booking.UserID.String()),
		)
		return fmt.Errorf("create booking %s: %w", booking.OrderID, err)
	}

	return nil
}

func (r *bookingRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error) {
	query := `
		SELECT id, order_id, user_id, schedule_id, total_seats, total_price, status, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`

	var booking entity.Booking
	err := r.db.QueryRow(ctx, query, id).Scan(
		&booking.ID,
		&booking.OrderID,
		&booking.UserID,
		&booking.ScheduleID,
		&booking.TotalSeats,
		&booking.TotalPrice,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find booking by ID",
			zap.Error(err),
			zap.String("booking_id", id.String()),
		)
		return nil, fmt.Errorf("find booking by ID %s: %w", id.String(), err)
	}

	return &booking, nil
}

func (r *bookingRepository) FindByOrderID(ctx context.Context, orderID string) (*entity.Booking, error) {
	query := `
		SELECT id, order_id, user_id, schedule_id, total_seats, total_price, status, created_at, updated_at
		FROM bookings
		WHERE order_id = $1
	`

	var booking entity.Booking
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&booking.ID,
		&booking.OrderID,
		&booking.UserID,
		&booking.ScheduleID,
		&booking.TotalSeats,
		&booking.TotalPrice,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find booking by order ID",
			zap.Error(err),
			zap.String("order_id", orderID),
		)
		return nil, fmt.Errorf("find booking by order ID %s: %w", orderID, err)
	}

	return &booking, nil
}

func (r *bookingRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Booking, error) {
	query := `
		SELECT id, order_id, user_id, schedule_id, total_seats, total_price, status, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.log.Error("Failed to find bookings by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
		)
		return nil, fmt.Errorf("find bookings by user ID %s: %w", userID.String(), err)
	}
	defer rows.Close()

	var bookings []*entity.Booking
	for rows.Next() {
		var booking entity.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.OrderID,
			&booking.UserID,
			&booking.ScheduleID,
			&booking.TotalSeats,
			&booking.TotalPrice,
			&booking.Status,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan booking row", zap.Error(err))
			return nil, fmt.Errorf("scan booking row: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	return bookings, nil
}

func (r *bookingRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE user_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		r.log.Error("Failed to count bookings by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return 0, fmt.Errorf("count bookings by user ID %s: %w", userID.String(), err)
	}

	return count, nil
}

func (r *bookingRepository) Update(ctx context.Context, booking *entity.Booking) error {
	query := `
		UPDATE bookings
		SET order_id = $2, user_id = $3, schedule_id = $4, total_seats = $5, 
		    total_price = $6, status = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		booking.ID,
		booking.OrderID,
		booking.UserID,
		booking.ScheduleID,
		booking.TotalSeats,
		booking.TotalPrice,
		booking.Status,
		booking.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update booking",
			zap.Error(err),
			zap.String("booking_id", booking.ID.String()),
		)
		return fmt.Errorf("update booking %s: %w", booking.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking %s not found", booking.ID.String())
	}

	return nil
}

func (r *bookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM bookings WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete booking",
			zap.Error(err),
			zap.String("booking_id", id.String()),
		)
		return fmt.Errorf("delete booking %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking %s not found", id.String())
	}

	r.log.Info("Booking deleted", zap.String("booking_id", id.String()))
	return nil
}

func (r *bookingRepository) FindByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*entity.Booking, error) {
	query := `
		SELECT id, order_id, user_id, schedule_id, total_seats, total_price, status, created_at, updated_at
		FROM bookings
		WHERE schedule_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		r.log.Error("Failed to find bookings by schedule ID",
			zap.Error(err),
			zap.String("schedule_id", scheduleID.String()),
		)
		return nil, fmt.Errorf("find bookings by schedule ID %s: %w", scheduleID.String(), err)
	}
	defer rows.Close()

	var bookings []*entity.Booking
	for rows.Next() {
		var booking entity.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.OrderID,
			&booking.UserID,
			&booking.ScheduleID,
			&booking.TotalSeats,
			&booking.TotalPrice,
			&booking.Status,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan booking row", zap.Error(err))
			return nil, fmt.Errorf("scan booking row: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	return bookings, nil
}

func (r *bookingRepository) FindConfirmedByScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]*entity.Booking, error) {
	query := `
		SELECT id, order_id, user_id, schedule_id, total_seats, total_price, status, created_at, updated_at
		FROM bookings
		WHERE schedule_id = $1 AND status = 'confirmed'
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		r.log.Error("Failed to find confirmed bookings by schedule ID",
			zap.Error(err),
			zap.String("schedule_id", scheduleID.String()),
		)
		return nil, fmt.Errorf("find confirmed bookings by schedule ID %s: %w", scheduleID.String(), err)
	}
	defer rows.Close()

	var bookings []*entity.Booking
	for rows.Next() {
		var booking entity.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.OrderID,
			&booking.UserID,
			&booking.ScheduleID,
			&booking.TotalSeats,
			&booking.TotalPrice,
			&booking.Status,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan booking row", zap.Error(err))
			return nil, fmt.Errorf("scan booking row: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	return bookings, nil
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, bookingID uuid.UUID, status entity.BookingStatus) error {
	query := `UPDATE bookings SET status = $2, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, bookingID, status)
	if err != nil {
		r.log.Error("Failed to update booking status",
			zap.Error(err),
			zap.String("booking_id", bookingID.String()),
			zap.String("status", string(status)),
		)
		return fmt.Errorf("update booking %s status to %s: %w", bookingID.String(), string(status), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking %s not found", bookingID.String())
	}

	return nil
}
