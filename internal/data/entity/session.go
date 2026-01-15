package entity

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	BaseSimple
	UserID    uuid.UUID  `db:"user_id"`
	Token     uuid.UUID  `db:"token"`
	UserAgent *string    `db:"user_agent"`
	IPAddress *string    `db:"ip_address"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}
