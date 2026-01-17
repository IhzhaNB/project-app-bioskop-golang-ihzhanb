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

type PaymentMethodRepository interface {
	Create(ctx context.Context, paymentMethod *entity.PaymentMethod) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.PaymentMethod, error)
	FindAllActive(ctx context.Context) ([]*entity.PaymentMethod, error)
	Update(ctx context.Context, paymentMethod *entity.PaymentMethod) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type paymentMethodRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewPaymentMethodRepository(db database.PgxIface, log *zap.Logger) PaymentMethodRepository {
	return &paymentMethodRepository{
		db:  db,
		log: log.With(zap.String("repository", "payment_method")),
	}
}

func (r *paymentMethodRepository) Create(ctx context.Context, paymentMethod *entity.PaymentMethod) error {
	query := `
		INSERT INTO payment_methods (id, name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		paymentMethod.ID,
		paymentMethod.Name,
		paymentMethod.IsActive,
		paymentMethod.CreatedAt,
		paymentMethod.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create payment method",
			zap.Error(err),
			zap.String("name", paymentMethod.Name),
		)
		return fmt.Errorf("create payment method %s: %w", paymentMethod.Name, err)
	}

	return nil
}

func (r *paymentMethodRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.PaymentMethod, error) {
	query := `
		SELECT id, name, is_active, created_at, updated_at, deleted_at
		FROM payment_methods
		WHERE id = $1 AND deleted_at IS NULL
	`

	var paymentMethod entity.PaymentMethod
	err := r.db.QueryRow(ctx, query, id).Scan(
		&paymentMethod.ID,
		&paymentMethod.Name,
		&paymentMethod.IsActive,
		&paymentMethod.CreatedAt,
		&paymentMethod.UpdatedAt,
		&paymentMethod.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find payment method by ID",
			zap.Error(err),
			zap.String("payment_method_id", id.String()),
		)
		return nil, fmt.Errorf("find payment method by ID %s: %w", id.String(), err)
	}

	return &paymentMethod, nil
}

func (r *paymentMethodRepository) FindAllActive(ctx context.Context) ([]*entity.PaymentMethod, error) {
	query := `
		SELECT id, name, is_active, created_at, updated_at
		FROM payment_methods
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.log.Error("Failed to find all active payment methods", zap.Error(err))
		return nil, fmt.Errorf("find all active payment methods: %w", err)
	}
	defer rows.Close()

	var paymentMethods []*entity.PaymentMethod
	for rows.Next() {
		var pm entity.PaymentMethod
		err := rows.Scan(
			&pm.ID,
			&pm.Name,
			&pm.IsActive,
			&pm.CreatedAt,
			&pm.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan payment method row", zap.Error(err))
			return nil, fmt.Errorf("scan payment method row: %w", err)
		}
		paymentMethods = append(paymentMethods, &pm)
	}

	return paymentMethods, nil
}

func (r *paymentMethodRepository) Update(ctx context.Context, paymentMethod *entity.PaymentMethod) error {
	query := `
		UPDATE payment_methods
		SET name = $2, is_active = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		paymentMethod.ID,
		paymentMethod.Name,
		paymentMethod.IsActive,
		paymentMethod.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update payment method",
			zap.Error(err),
			zap.String("payment_method_id", paymentMethod.ID.String()),
		)
		return fmt.Errorf("update payment method %s: %w", paymentMethod.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment method %s not found or already deleted", paymentMethod.ID.String())
	}

	return nil
}

func (r *paymentMethodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE payment_methods SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete payment method",
			zap.Error(err),
			zap.String("payment_method_id", id.String()),
		)
		return fmt.Errorf("delete payment method %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment method %s not found", id.String())
	}

	r.log.Info("Payment method deleted", zap.String("payment_method_id", id.String()))
	return nil
}
