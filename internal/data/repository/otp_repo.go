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

type OTPRepository interface {
	Create(ctx context.Context, otp *entity.OTP) error
	FindValidOTP(ctx context.Context, email, otpCode, otpType string) (*entity.OTP, error)
	MarkAsUsed(ctx context.Context, otpID uuid.UUID) error
	CleanExpiredOTPs(ctx context.Context) error
}

type otpRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewOTPRepository(db database.PgxIface, log *zap.Logger) OTPRepository {
	return &otpRepository{
		db:  db,
		log: log.With(zap.String("repository", "otp")),
	}
}

func (r *otpRepository) Create(ctx context.Context, otp *entity.OTP) error {
	query := `
		INSERT INTO otps (id, user_id, email, otp_code, otp_type, 
		                  expires_at, is_used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		otp.ID,
		otp.UserID,
		otp.Email,
		otp.OTPCode,
		otp.OTPType,
		otp.ExpiresAt,
		otp.IsUsed,
		otp.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create OTP",
			zap.Error(err),
			zap.String("email", otp.Email),
		)
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	return nil
}

func (r *otpRepository) FindValidOTP(ctx context.Context, email, otpCode, otpType string) (*entity.OTP, error) {
	query := `
		SELECT id, user_id, email, otp_code, otp_type, 
		       expires_at, is_used, created_at
		FROM otps 
		WHERE email = $1 
		  AND otp_code = $2 
		  AND otp_type = $3
		  AND is_used = false
		  AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otp entity.OTP
	err := r.db.QueryRow(ctx, query, email, otpCode, otpType).Scan(
		&otp.ID,
		&otp.UserID,
		&otp.Email,
		&otp.OTPCode,
		&otp.OTPType,
		&otp.ExpiresAt,
		&otp.IsUsed,
		&otp.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find valid OTP",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("failed to find OTP: %w", err)
	}

	return &otp, nil
}

func (r *otpRepository) MarkAsUsed(ctx context.Context, otpID uuid.UUID) error {
	query := `
		UPDATE otps 
		SET is_used = true
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, otpID)
	if err != nil {
		r.log.Error("Failed to mark OTP as used",
			zap.Error(err),
			zap.String("otp_id", otpID.String()),
		)
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("OTP not found")
	}

	return nil
}

func (r *otpRepository) CleanExpiredOTPs(ctx context.Context) error {
	query := `
		DELETE FROM otps 
		WHERE expires_at < NOW() - INTERVAL '1 day'
	`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		r.log.Error("Failed to clean expired OTPs",
			zap.Error(err),
		)
		return fmt.Errorf("failed to clean OTPs: %w", err)
	}

	return nil
}
