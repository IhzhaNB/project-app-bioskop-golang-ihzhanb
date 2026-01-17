package repository

import (
	"context"
	"fmt"
	"time"

	"cinema-booking/internal/data/entity"
	"cinema-booking/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *entity.Schedule) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error)
	FindByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.Schedule, error)
	FindByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Schedule, error)
	FindByDateAndHall(ctx context.Context, hallID uuid.UUID, date time.Time) ([]*entity.Schedule, error)
	Update(ctx context.Context, schedule *entity.Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error
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
			zap.Time("show_date", schedule.ShowDate),
		)
		return fmt.Errorf("create schedule for movie %s hall %s: %w",
			schedule.MovieID.String(), schedule.HallID.String(), err)
	}

	return nil
}

func (r *scheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE id = $1
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
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to find schedule by ID",
			zap.Error(err),
			zap.String("schedule_id", id.String()),
		)
		return nil, fmt.Errorf("find schedule by ID %s: %w", id.String(), err)
	}

	return &schedule, nil
}

func (r *scheduleRepository) FindByMovieID(ctx context.Context, movieID uuid.UUID) ([]*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE movie_id = $1
		ORDER BY show_date, show_time
	`

	rows, err := r.db.Query(ctx, query, movieID)
	if err != nil {
		r.log.Error("Failed to find schedules by movie ID",
			zap.Error(err),
			zap.String("movie_id", movieID.String()),
		)
		return nil, fmt.Errorf("find schedules by movie ID %s: %w", movieID.String(), err)
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
			return nil, fmt.Errorf("scan schedule row: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) FindByHallID(ctx context.Context, hallID uuid.UUID) ([]*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE hall_id = $1
		ORDER BY show_date, show_time
	`

	rows, err := r.db.Query(ctx, query, hallID)
	if err != nil {
		r.log.Error("Failed to find schedules by hall ID",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
		)
		return nil, fmt.Errorf("find schedules by hall ID %s: %w", hallID.String(), err)
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
			return nil, fmt.Errorf("scan schedule row: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) FindByDateAndHall(ctx context.Context, hallID uuid.UUID, date time.Time) ([]*entity.Schedule, error) {
	query := `
		SELECT id, movie_id, hall_id, show_date, show_time, price, created_at, updated_at
		FROM schedules
		WHERE hall_id = $1 AND show_date = $2
		ORDER BY show_time
	`

	rows, err := r.db.Query(ctx, query, hallID, date)
	if err != nil {
		r.log.Error("Failed to find schedules by hall and date",
			zap.Error(err),
			zap.String("hall_id", hallID.String()),
			zap.Time("date", date),
		)
		return nil, fmt.Errorf("find schedules by hall %s date %s: %w",
			hallID.String(), date.Format("2006-01-02"), err)
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
			return nil, fmt.Errorf("scan schedule row: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

func (r *scheduleRepository) Update(ctx context.Context, schedule *entity.Schedule) error {
	query := `
		UPDATE schedules
		SET movie_id = $2, hall_id = $3, show_date = $4, show_time = $5, price = $6, updated_at = $7
		WHERE id = $1
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
		return fmt.Errorf("update schedule %s: %w", schedule.ID.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("schedule %s not found", schedule.ID.String())
	}

	return nil
}

func (r *scheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM schedules WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error("Failed to delete schedule",
			zap.Error(err),
			zap.String("schedule_id", id.String()),
		)
		return fmt.Errorf("delete schedule %s: %w", id.String(), err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("schedule %s not found", id.String())
	}

	r.log.Info("Schedule deleted", zap.String("schedule_id", id.String()))
	return nil
}
