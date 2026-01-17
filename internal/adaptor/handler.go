package adaptor

import (
	"cinema-booking/internal/usecase"

	"go.uber.org/zap"
)

type Handler struct {
	Auth    *AuthHandler
	User    *UserHandler
	Movie   *MovieHandler
	Cinema  *CinemaHandler
	Booking *BookingHandler
	Review  *ReviewHandler
}

func NewHandler(service *usecase.Service, log *zap.Logger) *Handler {
	return &Handler{
		Auth:    NewAuthHandler(service.Auth, log),
		User:    NewUserHandler(service.User, log),
		Movie:   NewMovieHandler(service.Movie, log),
		Cinema:  NewCinemaHandler(service.Cinema, log),
		Booking: NewBookingHandler(service.Booking, log),
		Review:  NewReviewHandler(service.Review, log),
	}
}
