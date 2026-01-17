package entity

import (
	"time"

	"github.com/google/uuid"
)

type OTPType string

const (
	OTPTypeEmailVerification OTPType = "email_verification"
	OTPTypePasswordReset     OTPType = "password_reset"
)

type OTP struct {
	BaseSimple
	UserID    uuid.UUID `db:"user_id"`
	Email     string    `db:"email"`
	OTPCode   string    `db:"otp_code"`
	OTPType   OTPType   `db:"otp_type"`
	ExpiresAt time.Time `db:"expires_at"`
	IsUsed    bool      `db:"is_used"`
}
