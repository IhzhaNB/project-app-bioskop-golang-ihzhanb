package database

import (
	"cinema-booking/pkg/utils"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxIface interface untuk abstraction database
type PgxIface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
}

// DB wrapper struct
type DB struct {
	pool *pgxpool.Pool
}

// Query implements PgxIface
func (db *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// QueryRow implements PgxIface
func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Exec implements PgxIface - FIXED
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}

// Begin implements PgxIface
func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// Ping implements PgxIface
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Close implements PgxIface
func (db *DB) Close() {
	db.pool.Close()
}

// InitDB membuat koneksi database pool
func InitDB(config utils.DatabaseConfig) (PgxIface, error) {
	// Build connection string
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s",
		config.User, config.Password, config.Name, config.Host)

	// Parse config
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	// Pool configuration
	poolConfig.MaxConns = int32(config.MaxConns)
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Test connection
	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database failed: %w", err)
	}

	return &DB{pool: pool}, nil
}
