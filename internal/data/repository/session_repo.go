package repository

import (
	"context"
	"fmt"

	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindValidSession(ctx context.Context, token string) (*entity.Session, error)
	Revoke(ctx context.Context, token string) error
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
			zap.String("token", session.Token.String()),
		)
		return fmt.Errorf("create session for user %s: %w", session.UserID.String(), err)
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
		return nil, fmt.Errorf("find valid session for token %s: %w", token, err)
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
		return fmt.Errorf("revoke session token %s: %w", token, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session token %s not found or already revoked", token)
	}

	return nil
}
