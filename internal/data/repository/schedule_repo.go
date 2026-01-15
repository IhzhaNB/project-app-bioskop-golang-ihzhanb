package repository

import (
	"context"
	"fmt"
	"strings"

	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ScheduleRepository interface {
	// CRUD Schedule
	Create(ctx context.Context, schedule *entity.Schedule) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error)
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*entity.Schedule, error)
	CountAll(ctx context.Context, filters map[string]interface{}) (int64, error)
	Update(ctx context.Context, schedule *entity.Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Special queries
	FindByMovieID(ctx context.Context, movieID uuid.UUID, date *string) ([]*entity.Schedule, error)
	FindByHallID(ctx context.Context, hallID uuid.UUID, date *string) ([]*entity.Schedule, error)
	FindByCinemaID(ctx context.Context, cinemaID uuid.UUID, date *string) ([]*entity.Schedule, error)
	FindAvailableSchedules(ctx context.Context, movieID uuid.UUID, date string) ([]*entity.Schedule, error)

	// Check seat availability
	CheckSeatAvailability(ctx context.Context, scheduleID uuid.UUID) (int, error)
}

type scheduleRepository struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewScheduleRepository(db database.PgxIface, log *zap.Logger) ScheduleRepository {
	return &scheduleRepository{
		db:  db,
		log: log.With(zap.String("repository", "schedule")),
	}
}

func (r *scheduleRepository) Create(ctx context.Context, schedule *entity.Schedule) error {
	query := `
		INSERT INTO schedules (id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		schedule.ID,
		schedule.MovieID,
		schedule.HallID,
		schedule.ShowDate,
		schedule.ShowTime,
		schedule.Price,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to create schedule",
			zap.Error(err),
			zap.String("movie_id", schedule.MovieID.String()),
			zap.String("hall_id", schedule.HallID.String()),
		)
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}

func (r *scheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at, deleted_at
		FROM schedules
		WHERE id = $1 AND deleted_at IS NULL
	`

	var schedule entity.Schedule
	err := r.db.QueryRow(ctx, query, id).Scan(
		&schedule.ID,
		&schedule.MovieID,
		&schedule.HallID,
		&schedule.ShowDate,
		&schedule.ShowTime,
		&schedule.Price,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&schedule.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find schedule by ID",
			zap.Error(err),
			zap.String("schedule_id", id.String()),
		)
		return nil, fmt.Errorf("failed to find schedule: %w", err)
	}

	return &schedule, nil
}

func (r *scheduleRepository) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*entity.Schedule, error) {
	// Calculate offset
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Build query dengan filters
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE deleted_at IS NULL
	`)

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if movieID, ok := filters["movie_id"].(uuid.UUID); ok {
		queryBuilder.WriteString(fmt.Sprintf(" AND movie_id = $%d", argCount))
		args = append(args, movieID)
		argCount++
	}

	if hallID, ok := filters["hall_id"].(uuid.UUID); ok {
		queryBuilder.WriteString(fmt.Sprintf(" AND hall_id = $%d", argCount))
		args = append(args, hallID)
		argCount++
	}

	if cinemaID, ok := filters["cinema_id"].(uuid.UUID); ok {
		// Need to join with halls table
		queryBuilder.WriteString(fmt.Sprintf(" AND hall_id IN (SELECT id FROM halls WHERE cinema_id = $%d)", argCount))
		args = append(args, cinemaID)
		argCount++
	}

	if date, ok := filters["show_date"].(string); ok && date != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND show_date = $%d::DATE", argCount))
		args = append(args, date)
		argCount++
	}

	if fromDate, ok := filters["from_date"].(string); ok && fromDate != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND show_date >= $%d::DATE", argCount))
		args = append(args, fromDate)
		argCount++
	}

	if toDate, ok := filters["to_date"].(string); ok && toDate != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND show_date <= $%d::DATE", argCount))
		args = append(args, toDate)
		argCount++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY show_date, show_time LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		r.log.Error("Failed to find all schedules",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("limit", limit),
			zap.Any("filters", filters),
		)
		return nil, fmt.Errorf("failed to find schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*entity.Schedule
	for rows.Next() {
		var schedule entity.Schedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.MovieID,
			&schedule.HallID,
			&schedule.ShowDate,
			&schedule.ShowTime,
			&schedule.Price,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan schedule row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) CountAll(ctx context.Context, filters map[string]interface{}) (int64, error) {
	// Build count query
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`SELECT COUNT(*) FROM schedules WHERE deleted_at IS NULL`)

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if movieID, ok := filters["movie_id"].(uuid.UUID); ok {
		queryBuilder.WriteString(fmt.Sprintf(" AND movie_id = $%d", argCount))
		args = append(args, movieID)
		argCount++
	}

	if hallID, ok := filters["hall_id"].(uuid.UUID); ok {
		queryBuilder.WriteString(fmt.Sprintf(" AND hall_id = $%d", argCount))
		args = append(args, hallID)
		argCount++
	}

	if cinemaID, ok := filters["cinema_id"].(uuid.UUID); ok {
		queryBuilder.WriteString(fmt.Sprintf(" AND hall_id IN (SELECT id FROM halls WHERE cinema_id = $%d)", argCount))
		args = append(args, cinemaID)
		argCount++
	}

	if date, ok := filters["show_date"].(string); ok && date != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND show_date = $%d::DATE", argCount))
		args = append(args, date)
		argCount++
	}

	var total int64
	query := queryBuilder.String()

	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		r.log.Error("Failed to count schedules",
			zap.Error(err),
			zap.Any("filters", filters),
		)
		return 0, fmt.Errorf("failed to count schedules: %w", err)
	}

	return total, nil
}

func (r *scheduleRepository) FindByMovieID(ctx context.Context, movieID uuid.UUID, date *string) ([]*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE movie_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{movieID}

	if date != nil && *date != "" {
		query += " AND show_date = $2::DATE"
		args = append(args, *date)
	}

	query += " ORDER BY show_date, show_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to find schedules by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
			zap.Stringp("date", date),
		)
		return nil, fmt.Errorf("failed to find schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*entity.Schedule
	for rows.Next() {
		var schedule entity.Schedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.MovieID,
			&schedule.HallID,
			&schedule.ShowDate,
			&schedule.ShowTime,
			&schedule.Price,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan schedule row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) FindByHallID(ctx context.Context, hallID uuid.UUID, date *string) ([]*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE hall_id = $1 AND deleted_at IS NULL
	`

	args := []interface{}{hallID}

	if date != nil && *date != "" {
		query += " AND show_date = $2::DATE"
		args = append(args, *date)
	}

	query += " ORDER BY show_date, show_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to find schedules by hall ID",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
			zap.Stringp("date", date),
		)
		return nil, fmt.Errorf("failed to find schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*entity.Schedule
	for rows.Next() {
		var schedule entity.Schedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.MovieID,
			&schedule.HallID,
			&schedule.ShowDate,
			&schedule.ShowTime,
			&schedule.Price,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan schedule row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) FindByCinemaID(ctx context.Context, cinemaID uuid.UUID, date *string) ([]*entity.Schedule, error) {
	query := `
		SELECT s.id, s.movie_id, s.hall_id, s.show_date, s.show_time, s.price, s.created_at, s.updated_at
		FROM schedules s
		INNER JOIN halls h ON s.hall_id = h.id
		WHERE h.cinema_id = $1 AND s.deleted_at IS NULL AND h.deleted_at IS NULL
	`

	args := []interface{}{cinemaID}

	if date != nil && *date != "" {
		query += " AND s.show_date = $2::DATE"
		args = append(args, *date)
	}

	query += " ORDER BY s.show_date, s.show_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("Failed to find schedules by cinema ID",
			zap.Error(err),
			zap.String("cinema_id", cinemaID.String()),
			zap.Stringp("date", date),
		)
		return nil, fmt.Errorf("failed to find schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*entity.Schedule
	for rows.Next() {
		var schedule entity.Schedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.MovieID,
			&schedule.HallID,
			&schedule.ShowDate,
			&schedule.ShowTime,
			&schedule.Price,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan schedule row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) FindAvailableSchedules(ctx context.Context, movieID uuid.UUID, date string) ([]*entity.Schedule, error) {
	query := `
		SELECT s.id, s.movie_id, s.hall_id, s.show_date, s.show_time, s.price, s.created_at, s.updated_at
		FROM schedules s
		WHERE s.movie_id = $1 
		  AND s.show_date = $2::DATE
		  AND s.deleted_at IS NULL
		  AND EXISTS (
			SELECT 1 FROM seats st
			WHERE st.hall_id = s.hall_id 
			  AND st.is_available = true
			  AND st.deleted_at IS NULL
		  )
		ORDER BY s.show_time
	`

	rows, err := r.db.Query(ctx, query, movieID, date)
	if err != nil {
		r.log.Error("Failed to find available schedules",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
			zap.String("date", date),
		)
		return nil, fmt.Errorf("failed to find available schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*entity.Schedule
	for rows.Next() {
		var schedule entity.Schedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.MovieID,
			&schedule.HallID,
			&schedule.ShowDate,
			&schedule.ShowTime,
			&schedule.Price,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			r.log.Error("Failed to scan schedule row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) Update(ctx context.Context, schedule *entity.Schedule) error {
	query := `
		UPDATE schedules
		SET movie_id = $2, hall_id = $3, show_date = $4, show_time = $5, price = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		schedule.ID,
		schedule.MovieID,
		schedule.HallID,
		schedule.ShowDate,
		schedule.ShowTime,
		schedule.Price,
		schedule.UpdatedAt,
	)

	if err != nil {
		r.log.Error("Failed to update schedule",
			zap.Error(err),
			zap.String("schedule_id", schedule.ID.String()),
		)
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("schedule not found or already deleted")
	}

	return nil
}

func (r *scheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE schedules SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete schedule",
			zap.Error(err),
			zap.String("schedule_id", id.String()),
		)
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("schedule not found or already deleted")
	}

	r.log.Info("Schedule soft deleted", zap.String("schedule_id", id.String()))
	return nil
}

func (r *scheduleRepository) CheckSeatAvailability(ctx context.Context, scheduleID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) as available_seats
		FROM seats s
		WHERE s.hall_id = (
			SELECT hall_id FROM schedules WHERE id = $1 AND deleted_at IS NULL
		)
		AND s.is_available = true
		AND s.deleted_at IS NULL
	`

	var availableSeats int
	err := r.db.QueryRow(ctx, query, scheduleID).Scan(&availableSeats)
	if err != nil {
		r.log.Error("Failed to check seat availability",
			zap.Error(err),
			zap.String("schedule_id", scheduleID.String()),
		)
		return 0, fmt.Errorf("failed to check seat availability: %w", err)
	}

	return availableSeats, nil
}
