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

type BookingService interface {
	// Public endpoints (butuh auth)
	CreateBooking(ctx context.Context, userID string, req *request.CreateBookingRequest) (*response.BookingResponse, error)
	GetUserBookings(ctx context.Context, userID string, req *request.PaginatedRequest) (*response.PaginatedResponse[response.BookingResponse], error)

	// Payment
	ProcessPayment(ctx context.Context, userID string, req *request.ProcessPaymentRequest) (*response.PaymentResponse, error)
	GetPaymentMethods(ctx context.Context) ([]*response.PaymentMethodResponse, error)

	// Admin endpoints (optional)
	GetBookingByID(ctx context.Context, bookingID string) (*response.BookingDetailResponse, error)
	CancelBooking(ctx context.Context, bookingID string) error
}

type bookingService struct {
	repo *repository.Repository // grouping semua booking-related repos
	log  *zap.Logger
}

func NewBookingService(repo *repository.Repository, log *zap.Logger) BookingService {
	return &bookingService{
		repo: repo,
		log:  log.With(zap.String("service", "booking")),
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID string, req *request.CreateBookingRequest) (*response.BookingResponse, error) {
	// Validate request
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Create booking validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Parse IDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	scheduleID, err := uuid.Parse(req.ScheduleID)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID format %s: %w", req.ScheduleID, err)
	}

	// Validate schedule exists
	schedule, err := s.repo.Schedule.FindByID(ctx, scheduleID)
	if err != nil || schedule == nil {
		return nil, fmt.Errorf("schedule %s not found", req.ScheduleID)
	}

	// Check if schedule is in the future
	if schedule.ShowDate.Before(time.Now().Add(-24 * time.Hour)) {
		return nil, fmt.Errorf("cannot book for past schedule")
	}

	// Parse seat IDs
	seatUUIDs := make([]uuid.UUID, len(req.SeatIDs))
	for i, seatIDStr := range req.SeatIDs {
		seatID, err := uuid.Parse(seatIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid seat ID format %s: %w", seatIDStr, err)
		}
		seatUUIDs[i] = seatID
	}

	// Check seat availability
	bookedSeats, err := s.repo.BookingSeat.FindBookedSeatsBySchedule(ctx, scheduleID)
	if err != nil {
		s.log.Error("Failed to check booked seats", zap.Error(err))
		return nil, fmt.Errorf("check seat availability: %w", err)
	}

	// Check each seat
	for _, seatID := range seatUUIDs {
		// Check if seat exists and in correct hall
		seat, err := s.repo.Seat.FindByID(ctx, seatID)
		if err != nil || seat == nil {
			return nil, fmt.Errorf("seat %s not found", seatID.String())
		}

		// Check if seat is in the correct hall for this schedule
		if seat.HallID != schedule.HallID {
			return nil, fmt.Errorf("seat %s not in schedule hall", seatID.String())
		}

		// Check if seat is already booked
		for _, bookedSeatID := range bookedSeats {
			if seatID == bookedSeatID {
				return nil, fmt.Errorf("seat %s is already booked", seatID.String())
			}
		}
	}

	// Get hall for price calculation
	hall, err := s.repo.Hall.FindByID(ctx, schedule.HallID)
	if err != nil || hall == nil {
		return nil, fmt.Errorf("hall not found for schedule")
	}

	// Calculate total price
	totalPrice := schedule.Price * float64(len(seatUUIDs))

	// Create booking entity
	now := time.Now()
	booking := &entity.Booking{
		Base: entity.Base{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		OrderID:    utils.GenerateOrderID(),
		UserID:     userUUID,
		ScheduleID: scheduleID,
		TotalSeats: len(seatUUIDs),
		TotalPrice: totalPrice,
		Status:     entity.BookingStatusPending,
	}

	// Start transaction (simplified - kita pakai sequential untuk sekarang)
	// Save booking
	if err := s.repo.Booking.Create(ctx, booking); err != nil {
		s.log.Error("Failed to create booking",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("schedule_id", req.ScheduleID),
		)
		return nil, fmt.Errorf("create booking: %w", err)
	}

	// Create booking seats
	bookingSeats := make([]*entity.BookingSeat, len(seatUUIDs))
	for i, seatID := range seatUUIDs {
		bookingSeats[i] = &entity.BookingSeat{
			BaseSimple: entity.BaseSimple{
				ID:        uuid.New(),
				CreatedAt: now,
			},
			BookingID: booking.ID,
			SeatID:    seatID,
		}
	}

	if err := s.repo.BookingSeat.CreateBatch(ctx, bookingSeats); err != nil {
		// Rollback: delete booking
		s.repo.Booking.Delete(ctx, booking.ID)
		return nil, fmt.Errorf("create booking seats: %w", err)
	}

	s.log.Info("Booking created",
		zap.String("booking_id", booking.ID.String()),
		zap.String("order_id", booking.OrderID),
		zap.String("user_id", userID),
		zap.Int("seat_count", len(seatUUIDs)),
		zap.Float64("total_price", totalPrice),
	)

	// Get seat numbers for response
	seatNumbers := make([]string, len(seatUUIDs))
	for i, seatID := range seatUUIDs {
		seat, _ := s.repo.Seat.FindByID(ctx, seatID)
		if seat != nil {
			seatNumbers[i] = seat.SeatNumber
		}
	}

	// Build response
	return s.buildBookingResponse(ctx, booking, seatNumbers), nil
}

func (s *bookingService) GetUserBookings(ctx context.Context, userID string, req *request.PaginatedRequest) (*response.PaginatedResponse[response.BookingResponse], error) {
	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	limit := req.Limit()
	offset := req.Offset()

	// Get bookings
	bookings, err := s.repo.Booking.FindByUserID(ctx, userUUID, limit, offset)
	if err != nil {
		s.log.Error("Failed to get user bookings",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
		)
		return nil, fmt.Errorf("get user bookings: %w", err)
	}

	// Get total count
	total, err := s.repo.Booking.CountByUserID(ctx, userUUID)
	if err != nil {
		s.log.Error("Failed to count user bookings", zap.Error(err))
		return nil, fmt.Errorf("count user bookings: %w", err)
	}

	// Convert to response
	bookingResponses := make([]response.BookingResponse, len(bookings))
	for i, booking := range bookings {
		// Get seat numbers
		bookingSeats, _ := s.repo.BookingSeat.FindByBookingID(ctx, booking.ID)
		seatNumbers := make([]string, len(bookingSeats))
		for j, bs := range bookingSeats {
			seat, _ := s.repo.Seat.FindByID(ctx, bs.SeatID)
			if seat != nil {
				seatNumbers[j] = seat.SeatNumber
			}
		}

		// Get schedule details
		var movieTitle, cinemaName string
		var hallNumber int
		var showDate, showTime string

		schedule, _ := s.repo.Schedule.FindByID(ctx, booking.ScheduleID)
		if schedule != nil {
			movie, _ := s.repo.Movie.FindByID(ctx, schedule.MovieID)
			if movie != nil {
				movieTitle = movie.Title
			}

			hall, _ := s.repo.Hall.FindByID(ctx, schedule.HallID)
			if hall != nil {
				hallNumber = hall.HallNumber

				cinema, _ := s.repo.Cinema.FindByID(ctx, hall.CinemaID)
				if cinema != nil {
					cinemaName = cinema.Name
				}
			}

			showDate = schedule.ShowDate.Format("2006-01-02")
			showTime = schedule.ShowTime.Format("15:04")
		}

		// Get payment
		var paymentResp *response.PaymentResponse
		payment, _ := s.repo.Payment.FindByBookingID(ctx, booking.ID)
		if payment != nil {
			paymentMethod, _ := s.repo.PaymentMethod.FindByID(ctx, payment.PaymentMethodID)
			if paymentMethod != nil {
				paymentRespValue := response.PaymentToResponse(payment, paymentMethod)
				paymentResp = &paymentRespValue
			}
		}

		bookingResponses[i] = response.BookingResponse{
			ID:          booking.ID.String(),
			OrderID:     booking.OrderID,
			UserID:      booking.UserID.String(),
			ScheduleID:  booking.ScheduleID.String(),
			MovieTitle:  movieTitle,
			CinemaName:  cinemaName,
			HallNumber:  hallNumber,
			ShowDate:    showDate,
			ShowTime:    showTime,
			TotalSeats:  booking.TotalSeats,
			TotalPrice:  booking.TotalPrice,
			Status:      booking.Status,
			SeatNumbers: seatNumbers,
			Payment:     paymentResp,
			CreatedAt:   booking.CreatedAt,
		}
	}

	s.log.Info("User bookings retrieved",
		zap.String("user_id", userID),
		zap.Int("count", len(bookings)),
		zap.Int64("total", total),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
	)

	return response.NewPaginatedResponse(bookingResponses, req.Page, req.PerPage, total), nil
}

func (s *bookingService) ProcessPayment(ctx context.Context, userID string, req *request.ProcessPaymentRequest) (*response.PaymentResponse, error) {
	// Validate request
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Process payment validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Parse IDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID format %s: %w", req.BookingID, err)
	}

	paymentMethodID, err := uuid.Parse(req.PaymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method ID format %s: %w", req.PaymentMethodID, err)
	}

	// Get booking
	booking, err := s.repo.Booking.FindByID(ctx, bookingID)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("booking %s not found", req.BookingID)
	}

	// Check if booking belongs to user
	if booking.UserID != userUUID {
		return nil, fmt.Errorf("unauthorized to process payment for this booking")
	}

	// Check booking status
	if booking.Status != entity.BookingStatusPending {
		return nil, fmt.Errorf("booking status is %s, cannot process payment", booking.Status)
	}

	// Check if amount matches
	if req.Amount != booking.TotalPrice {
		return nil, fmt.Errorf("payment amount %.2f does not match booking total %.2f", req.Amount, booking.TotalPrice)
	}

	// Check payment method
	paymentMethod, err := s.repo.PaymentMethod.FindByID(ctx, paymentMethodID)
	if err != nil || paymentMethod == nil {
		return nil, fmt.Errorf("payment method %s not found", req.PaymentMethodID)
	}

	if !paymentMethod.IsActive {
		return nil, fmt.Errorf("payment method %s is not active", paymentMethod.Name)
	}

	// Create payment
	now := time.Now()
	payment := &entity.Payment{
		Base: entity.Base{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		BookingID:       bookingID,
		PaymentMethodID: paymentMethodID,
		Amount:          req.Amount,
		Status:          entity.PaymentStatusPending,
		TransactionID:   req.TransactionID,
	}

	// Simulate payment processing (dummy implementation)
	// In real app, integrate with payment gateway
	payment.Status = entity.PaymentStatusCompleted

	// Update booking status
	booking.Status = entity.BookingStatusConfirmed
	booking.UpdatedAt = now

	// Save payment and update booking (simplified - no transaction)
	if err := s.repo.Payment.Create(ctx, payment); err != nil {
		s.log.Error("Failed to create payment",
			zap.Error(err),
			zap.String("booking_id", req.BookingID),
		)
		return nil, fmt.Errorf("create payment: %w", err)
	}

	if err := s.repo.Booking.Update(ctx, booking); err != nil {
		s.log.Error("Failed to update booking status",
			zap.Error(err),
			zap.String("booking_id", req.BookingID),
		)
		// Continue anyway
	}

	s.log.Info("Payment processed",
		zap.String("payment_id", payment.ID.String()),
		zap.String("booking_id", req.BookingID),
		zap.String("payment_method", paymentMethod.Name),
		zap.Float64("amount", req.Amount),
		zap.String("status", string(payment.Status)),
	)

	// Build response
	paymentResp := response.PaymentToResponse(payment, paymentMethod)
	return &paymentResp, nil
}

func (s *bookingService) GetPaymentMethods(ctx context.Context) ([]*response.PaymentMethodResponse, error) {
	paymentMethods, err := s.repo.PaymentMethod.FindAllActive(ctx)
	if err != nil {
		s.log.Error("Failed to get payment methods", zap.Error(err))
		return nil, fmt.Errorf("get payment methods: %w", err)
	}

	paymentMethodResponses := make([]*response.PaymentMethodResponse, len(paymentMethods))
	for i, pm := range paymentMethods {
		pmResp := response.PaymentMethodToResponse(pm)
		paymentMethodResponses[i] = &pmResp
	}

	s.log.Info("Payment methods retrieved", zap.Int("count", len(paymentMethods)))
	return paymentMethodResponses, nil
}

// ==================== ADMIN METHODS ====================

func (s *bookingService) GetBookingByID(ctx context.Context, bookingID string) (*response.BookingDetailResponse, error) {
	// Parse booking ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID format %s: %w", bookingID, err)
	}

	booking, err := s.repo.Booking.FindByID(ctx, id)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("booking %s not found", bookingID)
	}

	// Get seat numbers
	bookingSeats, _ := s.repo.BookingSeat.FindByBookingID(ctx, booking.ID)
	seatNumbers := make([]string, len(bookingSeats))
	for i, bs := range bookingSeats {
		seat, _ := s.repo.Seat.FindByID(ctx, bs.SeatID)
		if seat != nil {
			seatNumbers[i] = seat.SeatNumber
		}
	}

	// Get schedule details
	var scheduleDetails response.ScheduleDetails
	schedule, _ := s.repo.Schedule.FindByID(ctx, booking.ScheduleID)
	if schedule != nil {
		movie, _ := s.repo.Movie.FindByID(ctx, schedule.MovieID)
		if movie != nil {
			scheduleDetails.MovieTitle = movie.Title
		}

		hall, _ := s.repo.Hall.FindByID(ctx, schedule.HallID)
		if hall != nil {
			scheduleDetails.HallNumber = hall.HallNumber

			cinema, _ := s.repo.Cinema.FindByID(ctx, hall.CinemaID)
			if cinema != nil {
				scheduleDetails.CinemaName = cinema.Name
			}
		}

		scheduleDetails.ShowDate = schedule.ShowDate.Format("2006-01-02")
		scheduleDetails.ShowTime = schedule.ShowTime.Format("15:04")
		scheduleDetails.Price = schedule.Price
	}

	// Get payment
	var paymentResp *response.PaymentResponse
	payment, _ := s.repo.Payment.FindByBookingID(ctx, booking.ID)
	if payment != nil {
		paymentMethod, _ := s.repo.PaymentMethod.FindByID(ctx, payment.PaymentMethodID)
		if paymentMethod != nil {
			paymentRespValue := response.PaymentToResponse(payment, paymentMethod)
			paymentResp = &paymentRespValue
		}
	}

	bookingResp := response.BookingResponse{
		ID:          booking.ID.String(),
		OrderID:     booking.OrderID,
		UserID:      booking.UserID.String(),
		ScheduleID:  booking.ScheduleID.String(),
		MovieTitle:  scheduleDetails.MovieTitle,
		CinemaName:  scheduleDetails.CinemaName,
		HallNumber:  scheduleDetails.HallNumber,
		ShowDate:    scheduleDetails.ShowDate,
		ShowTime:    scheduleDetails.ShowTime,
		TotalSeats:  booking.TotalSeats,
		TotalPrice:  booking.TotalPrice,
		Status:      booking.Status,
		SeatNumbers: seatNumbers,
		Payment:     paymentResp,
		CreatedAt:   booking.CreatedAt,
	}

	return &response.BookingDetailResponse{
		BookingResponse: bookingResp,
		ScheduleDetails: scheduleDetails,
	}, nil
}

func (s *bookingService) CancelBooking(ctx context.Context, bookingID string) error {
	// Parse booking ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
		return fmt.Errorf("invalid booking ID format %s: %w", bookingID, err)
	}

	booking, err := s.repo.Booking.FindByID(ctx, id)
	if err != nil || booking == nil {
		return fmt.Errorf("booking %s not found", bookingID)
	}

	// Check if booking can be cancelled
	if booking.Status != entity.BookingStatusPending && booking.Status != entity.BookingStatusConfirmed {
		return fmt.Errorf("booking status is %s, cannot cancel", booking.Status)
	}

	// Update booking status
	if err := s.repo.Booking.UpdateStatus(ctx, booking.ID, entity.BookingStatusCancelled); err != nil {
		s.log.Error("Failed to cancel booking",
			zap.Error(err),
			zap.String("booking_id", bookingID),
		)
		return fmt.Errorf("cancel booking %s: %w", bookingID, err)
	}

	s.log.Info("Booking cancelled",
		zap.String("booking_id", bookingID),
		zap.String("order_id", booking.OrderID),
	)

	return nil
}

// ==================== HELPER METHODS ====================

func (s *bookingService) buildBookingResponse(ctx context.Context, booking *entity.Booking, seatNumbers []string) *response.BookingResponse {
	// Get schedule details
	var movieTitle, cinemaName string
	var hallNumber int
	var showDate, showTime string

	schedule, _ := s.repo.Schedule.FindByID(ctx, booking.ScheduleID)
	if schedule != nil {
		movie, _ := s.repo.Movie.FindByID(ctx, schedule.MovieID)
		if movie != nil {
			movieTitle = movie.Title
		}

		hall, _ := s.repo.Hall.FindByID(ctx, schedule.HallID)
		if hall != nil {
			hallNumber = hall.HallNumber

			cinema, _ := s.repo.Cinema.FindByID(ctx, hall.CinemaID)
			if cinema != nil {
				cinemaName = cinema.Name
			}
		}

		showDate = schedule.ShowDate.Format("2006-01-02")
		showTime = schedule.ShowTime.Format("15:04")
	}

	return &response.BookingResponse{
		ID:          booking.ID.String(),
		OrderID:     booking.OrderID,
		UserID:      booking.UserID.String(),
		ScheduleID:  booking.ScheduleID.String(),
		MovieTitle:  movieTitle,
		CinemaName:  cinemaName,
		HallNumber:  hallNumber,
		ShowDate:    showDate,
		ShowTime:    showTime,
		TotalSeats:  booking.TotalSeats,
		TotalPrice:  booking.TotalPrice,
		Status:      booking.Status,
		SeatNumbers: seatNumbers,
		CreatedAt:   booking.CreatedAt,
	}
}
