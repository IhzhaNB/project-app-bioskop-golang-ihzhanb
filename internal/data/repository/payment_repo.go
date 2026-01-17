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

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	FindByBookingID(ctx context.Context, bookingID uuid.UUID) (*entity.Payment, error)
	Update(ctx context.Context, payment *entity.Payment) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Business queries
	UpdateStatus(ctx context.Context, paymentID uuid.UUID, status entity.PaymentStatus, transactionID *string) error
}

type paymentRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewPaymentRepository(db database.PgxIface, log *zap.Logger) PaymentRepository {
	return &paymentRepository{
		db:  db,
		log: log.With(zap.String("repository", "payment")),
	}
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	query := `
		INSERT INTO payments (id, booking_id, payment_method_id, amount, status, transaction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.BookingID,
		payment.PaymentMethodID,
		payment.Amount,
		payment.Status,
		payment.TransactionID,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create payment",
			zap.Error(err),
			zap.String("booking_id", payment.BookingID.String()),
			zap.String("payment_method_id", payment.PaymentMethodID.String()),
		)
		return fmt.Errorf("create payment for booking %s: %w", payment.BookingID.String(), err)
	}

	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
	query := `
		SELECT id, booking_id, payment_method_id, amount, status, transaction_id, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	var payment entity.Payment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.PaymentMethodID,
		&payment.Amount,
		&payment.Status,
		&payment.TransactionID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find payment by ID",
			zap.Error(err),
			zap.String("payment_id", id.String()),
		)
		return nil, fmt.Errorf("find payment by ID %s: %w", id.String(), err)
	}

	return &payment, nil
}

func (r *paymentRepository) FindByBookingID(ctx context.Context, bookingID uuid.UUID) (*entity.Payment, error) {
	query := `
		SELECT id, booking_id, payment_method_id, amount, status, transaction_id, created_at, updated_at
		FROM payments
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var payment entity.Payment
	err := r.db.QueryRow(ctx, query, bookingID).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.PaymentMethodID,
		&payment.Amount,
		&payment.Status,
		&payment.TransactionID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find payment by booking ID",
			zap.Error(err),
			zap.String("booking_id", bookingID.String()),
		)
		return nil, fmt.Errorf("find payment by booking ID %s: %w", bookingID.String(), err)
	}

	return &payment, nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	query := `
		UPDATE payments
		SET booking_id = $2, payment_method_id = $3, amount = $4, 
		    status = $5, transaction_id = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.BookingID,
		payment.PaymentMethodID,
		payment.Amount,
		payment.Status,
		payment.TransactionID,
		payment.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update payment",
			zap.Error(err),
			zap.String("payment_id", payment.ID.String()),
		)
		return fmt.Errorf("update payment %s: %w", payment.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment %s not found", payment.ID.String())
	}

	return nil
}

func (r *paymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM payments WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete payment",
			zap.Error(err),
			zap.String("payment_id", id.String()),
		)
		return fmt.Errorf("delete payment %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment %s not found", id.String())
	}

	r.log.Info("Payment deleted", zap.String("payment_id", id.String()))
	return nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, paymentID uuid.UUID, status entity.PaymentStatus, transactionID *string) error {
	query := `
		UPDATE payments
		SET status = $2, transaction_id = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, paymentID, status, transactionID)
	if err != nil {
		r.log.Error("Failed to update payment status",
			zap.Error(err),
			zap.String("payment_id", paymentID.String()),
			zap.String("status", string(status)),
		)
		return fmt.Errorf("update payment %s status to %s: %w", paymentID.String(), string(status), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment %s not found", paymentID.String())
	}

	return nil
}
