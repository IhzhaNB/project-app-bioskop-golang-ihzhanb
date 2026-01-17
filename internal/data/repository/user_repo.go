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

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error)
	CountAll(ctx context.Context) (int64, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewUserRepository(db database.PgxIface, log *zap.Logger) UserRepository {
	return &userRepository{
		db:  db,
		log: log,
	}
}

// Create inserts a new user record into the database
func (ur *userRepository) Create(ctx context.Context, user *entity.User) error {
	// SQL query
	query := `
		INSERT INTO users (id, username, email, password, phone, role,
		                  email_verified, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Execute query
	_, err := ur.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.Role,
		user.EmailVerified,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		ur.log.Error("Failed to create user",
			zap.Error(err),
			zap.String("email", user.Email),
			zap.String("username", user.Username),
		)
		return fmt.Errorf("create user %s: %w", user.Email, err)
	}

	return nil
}

func (ur *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, username, email, password, phone, role,
		       email_verified, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user entity.User
	// QueryRow returns at most one row
	err := ur.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.Role,
		&user.EmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		ur.log.Error("Failed to find user by ID",
			zap.Error(err),
			zap.String("user_id", id.String()),
		)
		return nil, fmt.Errorf("find user by ID %s: %w", id.String(), err)
	}

	return &user, nil
}

func (ur *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password, phone, role,
		       email_verified, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user entity.User
	// QueryRow returns at most one row
	err := ur.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.Role,
		&user.EmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		ur.log.Error("Failed to find user by email",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("find user by email %s: %w", email, err)
	}

	return &user, nil
}

func (ur *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password, phone, role,
		       email_verified, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	var user entity.User
	// QueryRow returns at most one row
	err := ur.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Phone,
		&user.Role,
		&user.EmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		ur.log.Error("Failed to find user by username",
			zap.Error(err),
			zap.String("username", username),
		)
		return nil, fmt.Errorf("find user by username %s: %w", username, err)
	}

	return &user, nil
}

// FindAll retrieves paginated list of users
func (ur *userRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT id, username, email, password, phone, role,
		       email_verified, is_active, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	// Query returns multiple rows
	rows, err := ur.db.Query(ctx, query, limit, offset)
	if err != nil {
		ur.log.Error("Failed to get all users",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
		)
		return nil, fmt.Errorf("find all users limit %d offset %d: %w", limit, offset, err)
	}
	defer rows.Close() // IMPORTANT: Close rows to release database connection

	var users []*entity.User
	// Iterate through each row
	for rows.Next() {
		var user entity.User
		// Scan each column into user struct fields
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Phone,
			&user.Role,
			&user.EmailVerified,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			ur.log.Error("Failed to scan user row", zap.Error(err))
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, &user)
	}

	// Check for errors during iteration (not just database errors)
	if err := rows.Err(); err != nil {
		ur.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("iterate users rows: %w", err)
	}

	return users, nil
}

func (ur *userRepository) CountAll(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var count int64
	err := ur.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		ur.log.Error("Database error counting users",
			zap.Error(err),
		)
		return 0, fmt.Errorf("count all users: %w", err)
	}

	return count, nil
}

func (ur *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password = $4, phone = $5,
		    role = $6, email_verified = $7, is_active = $8,
		    updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	// Execute query
	result, err := ur.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Phone,
		user.Role,
		user.EmailVerified,
		user.IsActive,
		user.UpdatedAt,
	)

	if err != nil {
		ur.log.Error("Failed to update user",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email),
		)
		return fmt.Errorf("update user %s: %w", user.ID.String(), err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user %s not found or already deleted", user.ID.String())
	}

	return nil
}

func (ur *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	// Execute query
	result, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		ur.log.Error("Failed to delete user",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("delete user %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user %s not found", id.String())
	}

	ur.log.Info("User deleted", zap.String("id", id.String()))
	return nil
}
