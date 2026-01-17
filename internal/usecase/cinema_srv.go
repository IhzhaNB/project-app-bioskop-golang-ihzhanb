package usecase

import (
	"context"
	"fmt"
	"time"

	"cinema-booking/internal/data/entity"
	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/dto/response"
	"cinema-booking/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CinemaService interface {
	GetCinemas(ctx context.Context, req *request.PaginatedRequest, cityFilter *string) (*response.PaginatedResponse[response.CinemaResponse], error)
	GetCinemaByID(ctx context.Context, cinemaID string) (*response.CinemaDetailResponse, error)
	GetSeatAvailability(ctx context.Context, cinemaID, dateStr, timeStr string) ([]*response.SeatAvailabilityResponse, error)

	CreateCinema(ctx context.Context, req *request.CinemaRequest) (*response.CinemaResponse, error)
	UpdateCinema(ctx context.Context, cinemaID string, req *request.CinemaUpdateRequest) (*response.CinemaResponse, error)
	DeleteCinema(ctx context.Context, cinemaID string) error
}

type cinemaService struct {
	repo *repository.Repository // grouping semua cinema-related repos
	log  *zap.Logger
}

func NewCinemaService(repo *repository.Repository, log *zap.Logger) CinemaService {
	return &cinemaService{
		repo: repo,
		log:  log.With(zap.String("service", "cinema")),
	}
}

func (s *cinemaService) GetCinemas(ctx context.Context, req *request.PaginatedRequest, cityFilter *string) (*response.PaginatedResponse[response.CinemaResponse], error) {
	limit := req.Limit()
	offset := req.Offset()

	// Get cinemas from repository
	cinemas, err := s.repo.Cinema.FindAll(ctx, limit, offset, cityFilter)
	if err != nil {
		s.log.Error("Failed to get cinemas from repository",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
			zap.Stringp("city_filter", cityFilter),
		)
		return nil, fmt.Errorf("get cinemas: %w", err)
	}

	// Get total count
	total, err := s.repo.Cinema.CountAll(ctx, cityFilter)
	if err != nil {
		s.log.Error("Failed to count cinemas",
			zap.Error(err),
			zap.Stringp("city_filter", cityFilter),
		)
		return nil, fmt.Errorf("count cinemas: %w", err)
	}

	// Convert to response
	cinemaResponses := make([]response.CinemaResponse, len(cinemas))
	for i, cinema := range cinemas {
		cinemaResponses[i] = response.CinemaToResponse(cinema)
	}

	s.log.Info("Cinemas retrieved",
		zap.Int("count", len(cinemas)),
		zap.Int64("total", total),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
	)

	return response.NewPaginatedResponse(cinemaResponses, req.Page, req.PerPage, total), nil
}

func (s *cinemaService) GetCinemaByID(ctx context.Context, cinemaID string) (*response.CinemaDetailResponse, error) {
	// Parse cinema ID
	id, err := uuid.Parse(cinemaID)
	if err != nil {
		s.log.Warn("Invalid cinema ID format",
			zap.String("cinema_id", cinemaID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid cinema ID format %s: %w", cinemaID, err)
	}

	// Get cinema
	cinema, err := s.repo.Cinema.FindByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to get cinema by ID",
			zap.Error(err),
			zap.String("cinema_id", cinemaID),
		)
		return nil, fmt.Errorf("get cinema %s: %w", cinemaID, err)
	}

	if cinema == nil {
		return nil, fmt.Errorf("cinema %s not found", cinemaID)
	}

	// Get halls for this cinema
	halls, err := s.repo.Hall.FindByCinemaID(ctx, cinema.ID)
	if err != nil {
		s.log.Warn("Failed to get halls for cinema",
			zap.Error(err),
			zap.String("cinema_id", cinemaID),
		)
		// Continue with empty halls
	}

	// Convert halls to response
	hallResponses := make([]response.HallResponse, len(halls))
	for i, hall := range halls {
		hallResponses[i] = response.HallToResponse(hall)
	}

	s.log.Info("Cinema retrieved",
		zap.String("cinema_id", cinemaID),
		zap.String("name", cinema.Name),
		zap.Int("hall_count", len(halls)),
	)

	return &response.CinemaDetailResponse{
		CinemaResponse: response.CinemaToResponse(cinema),
		Halls:          hallResponses,
	}, nil
}

func (s *cinemaService) GetSeatAvailability(ctx context.Context, cinemaID, dateStr, timeStr string) ([]*response.SeatAvailabilityResponse, error) {
	// Parse cinema ID
	cinemaUUID, err := uuid.Parse(cinemaID)
	if err != nil {
		return nil, fmt.Errorf("invalid cinema ID format %s: %w", cinemaID, err)
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format %s: %w", dateStr, err)
	}

	// Parse time
	showTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid time format %s: %w", timeStr, err)
	}

	// Get cinema first (validate exists)
	cinema, err := s.repo.Cinema.FindByID(ctx, cinemaUUID)
	if err != nil || cinema == nil {
		return nil, fmt.Errorf("cinema %s not found", cinemaID)
	}

	// Get halls for this cinema
	halls, err := s.repo.Hall.FindByCinemaID(ctx, cinema.ID)
	if err != nil {
		s.log.Error("Failed to get halls for seat availability",
			zap.Error(err),
			zap.String("cinema_id", cinemaID),
		)
		return nil, fmt.Errorf("get halls for cinema %s: %w", cinemaID, err)
	}

	if len(halls) == 0 {
		return []*response.SeatAvailabilityResponse{}, nil
	}

	// For each hall, get seats and check schedule
	var results []*response.SeatAvailabilityResponse
	for _, hall := range halls {
		// Cari schedule untuk hall, date, dan time tertentu
		schedules, err := s.repo.Schedule.FindByDateAndHall(ctx, hall.ID, date)
		if err != nil {
			s.log.Warn("Failed to get schedules for hall",
				zap.Error(err),
				zap.String("hall_id", hall.ID.String()),
			)
			continue
		}

		// Cari schedule yang cocok dengan waktu yang diminta
		var targetSchedule *entity.Schedule
		for _, schedule := range schedules {
			scheduleTime := schedule.ShowTime.Format("15:04")
			if scheduleTime == showTime.Format("15:04") {
				targetSchedule = schedule
				break
			}
		}

		// Jika tidak ada schedule di waktu tersebut, return semua seat unavailable
		if targetSchedule == nil {
			s.log.Warn("No schedule found for hall at specified time",
				zap.String("hall_id", hall.ID.String()),
				zap.String("date", dateStr),
				zap.String("time", timeStr),
			)
			// Return empty atau semua seat unavailable
			continue
		}

		// Get all seats for this hall
		seats, err := s.repo.Seat.FindByHallID(ctx, hall.ID)
		if err != nil {
			s.log.Warn("Failed to get seats for hall",
				zap.Error(err),
				zap.String("hall_id", hall.ID.String()),
			)
			continue
		}

		// Get booked seats untuk schedule ini
		bookedSeats, err := s.repo.BookingSeat.FindBookedSeatsBySchedule(ctx, targetSchedule.ID)
		if err != nil {
			s.log.Warn("Failed to get booked seats for schedule",
				zap.Error(err),
				zap.String("schedule_id", targetSchedule.ID.String()),
			)
			// Continue dengan asumsi semua seat available
		}

		// Convert seats to response dengan status availability
		seatResponses := make([]response.SeatResponse, len(seats))
		for i, seat := range seats {
			seatResp := response.SeatToResponse(seat)

			// Check if seat is booked
			isBooked := false
			for _, bookedSeatID := range bookedSeats {
				if seat.ID == bookedSeatID {
					isBooked = true
					break
				}
			}

			// Update availability status
			seatResp.IsAvailable = !isBooked
			seatResponses[i] = seatResp
		}

		// Create response for this hall
		result := &response.SeatAvailabilityResponse{
			HallID: hall.ID.String(),
			Date:   date.Format("2006-01-02"),
			Time:   showTime.Format("15:04"),
			Seats:  seatResponses,
		}

		results = append(results, result)
	}

	s.log.Info("Seat availability checked",
		zap.String("cinema_id", cinemaID),
		zap.String("date", dateStr),
		zap.String("time", timeStr),
		zap.Int("hall_count", len(halls)),
		zap.Int("result_count", len(results)),
	)

	return results, nil
}

func (s *cinemaService) CreateCinema(ctx context.Context, req *request.CinemaRequest) (*response.CinemaResponse, error) {
	// Validate request
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Create cinema validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Create cinema entity
	now := time.Now()
	cinema := &entity.Cinema{
		Base: entity.Base{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:     req.Name,
		Location: req.Location,
		City:     req.City,
	}

	// Save cinema
	if err := s.repo.Cinema.Create(ctx, cinema); err != nil {
		s.log.Error("Failed to create cinema",
			zap.Error(err),
			zap.String("name", req.Name),
		)
		return nil, fmt.Errorf("create cinema: %w", err)
	}

	s.log.Info("Cinema created",
		zap.String("cinema_id", cinema.ID.String()),
		zap.String("name", cinema.Name),
		zap.String("city", cinema.City),
	)

	cinemaResp := response.CinemaToResponse(cinema)
	return &cinemaResp, nil
}

func (s *cinemaService) UpdateCinema(ctx context.Context, cinemaID string, req *request.CinemaUpdateRequest) (*response.CinemaResponse, error) {
	// Parse cinema ID
	id, err := uuid.Parse(cinemaID)
	if err != nil {
		return nil, fmt.Errorf("invalid cinema ID format %s: %w", cinemaID, err)
	}

	// Get existing cinema
	cinema, err := s.repo.Cinema.FindByID(ctx, id)
	if err != nil || cinema == nil {
		return nil, fmt.Errorf("cinema %s not found", cinemaID)
	}

	// Update fields if provided
	updated := false

	if req.Name != nil && *req.Name != cinema.Name {
		cinema.Name = *req.Name
		updated = true
	}

	if req.Location != nil && *req.Location != cinema.Location {
		cinema.Location = *req.Location
		updated = true
	}

	if req.City != nil && *req.City != cinema.City {
		cinema.City = *req.City
		updated = true
	}

	if updated {
		cinema.UpdatedAt = time.Now()
		if err := s.repo.Cinema.Update(ctx, cinema); err != nil {
			s.log.Error("Failed to update cinema",
				zap.Error(err),
				zap.String("cinema_id", cinemaID),
			)
			return nil, fmt.Errorf("update cinema %s: %w", cinemaID, err)
		}
	}

	s.log.Info("Cinema updated",
		zap.String("cinema_id", cinemaID),
		zap.String("name", cinema.Name),
		zap.Bool("was_updated", updated),
	)

	cinemaResp := response.CinemaToResponse(cinema)
	return &cinemaResp, nil
}

func (s *cinemaService) DeleteCinema(ctx context.Context, cinemaID string) error {
	// Parse cinema ID
	id, err := uuid.Parse(cinemaID)
	if err != nil {
		return fmt.Errorf("invalid cinema ID format %s: %w", cinemaID, err)
	}

	// Get cinema first for logging
	cinema, err := s.repo.Cinema.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find cinema: %w", err)
	}
	if cinema == nil {
		return fmt.Errorf("cinema %s not found", cinemaID)
	}

	// Soft delete cinema
	if err := s.repo.Cinema.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete cinema",
			zap.Error(err),
			zap.String("cinema_id", cinemaID),
		)
		return fmt.Errorf("delete cinema %s: %w", cinemaID, err)
	}

	s.log.Info("Cinema deleted",
		zap.String("cinema_id", cinemaID),
		zap.String("name", cinema.Name),
	)

	return nil
}
