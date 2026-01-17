package repository

import (
	"cinema-booking/pkg/database"

	"go.uber.org/zap"
)

type Repository struct {
	User          UserRepository
	Session       SessionRepository
	OTP           OTPRepository
	Movie         MovieRepository
	Genre         GenreRepository
	MovieGenre    MovieGenreRepository
	Cinema        CinemaRepository
	Hall          HallRepository
	Seat          SeatRepository
	Schedule      ScheduleRepository
	PaymentMethod PaymentMethodRepository
	Booking       BookingRepository
	BookingSeat   BookingSeatRepository
	Payment       PaymentRepository
	Review        ReviewRepository
}

func NewRepository(db database.PgxIface, log *zap.Logger) *Repository {
	return &Repository{
		User:          NewUserRepository(db, log),
		Session:       NewSessionRepository(db, log),
		OTP:           NewOTPRepository(db, log),
		Movie:         NewMovieRepository(db, log),
		Genre:         NewGenreRepository(db, log),
		MovieGenre:    NewMovieGenreRepository(db, log),
		Cinema:        NewCinemaRepository(db, log),
		Hall:          NewHallRepository(db, log),
		Seat:          NewSeatRepository(db, log),
		Schedule:      NewScheduleRepository(db, log),
		PaymentMethod: NewPaymentMethodRepository(db, log),
		Booking:       NewBookingRepository(db, log),
		BookingSeat:   NewBookingSeatRepository(db, log),
		Payment:       NewPaymentRepository(db, log),
		Review:        NewReviewRepository(db, log),
	}
}
