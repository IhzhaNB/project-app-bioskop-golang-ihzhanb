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

type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindValidSession(ctx context.Context, token string) (*entity.Session, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanExpiredSessions(ctx context.Context) error
}

type sessionRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewSessionRepository(db database.PgxIface, log *zap.Logger) SessionRepository {
	return &sessionRepository{
		db:  db,
		log: log.With(zap.String("repository", "session")),
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token, user_agent, ip_address, 
		                     expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.Token,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create session",
			zap.Error(err),
			zap.String("user_id", session.UserID.String()),
		)
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *sessionRepository) FindValidSession(ctx context.Context, token string) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token, user_agent, ip_address, 
		       expires_at, revoked_at, created_at
		FROM sessions 
		WHERE token = $1 
		  AND revoked_at IS NULL 
		  AND expires_at > NOW()
	`

	var session entity.Session
	err := r.db.QueryRow(ctx, query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find valid session",
			zap.Error(err),
			zap.String("token", token),
		)
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return &session, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, token string) error {
	query := `
		UPDATE sessions 
		SET revoked_at = NOW()
		WHERE token = $1 AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, token)
	if err != nil {
		r.log.Error("Failed to revoke session",
			zap.Error(err),
			zap.String("token", token),
		)
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	return nil
}

func (r *sessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE sessions 
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		r.log.Error("Failed to revoke all user sessions",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	return nil
}

func (r *sessionRepository) CleanExpiredSessions(ctx context.Context) error {
	query := `
		DELETE FROM sessions 
		WHERE expires_at < NOW() - INTERVAL '7 days'
	`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		r.log.Error("Failed to clean expired sessions",
			zap.Error(err),
		)
		return fmt.Errorf("failed to clean sessions: %w", err)
	}

	return nil
}
