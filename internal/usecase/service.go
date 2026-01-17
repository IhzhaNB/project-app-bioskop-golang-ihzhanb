package usecase

import (
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/utils"

	"go.uber.org/zap"
)

type Service struct {
	Auth    AuthService
	User    UserService
	Movie   MovieService
	Cinema  CinemaService
	Booking BookingService
	Review  ReviewService
}

func NewService(repo *repository.Repository, config *utils.Config, log *zap.Logger) *Service {
	return &Service{
		Auth:    NewAuthService(repo, config, log),
		User:    NewUserService(repo.User, log),
		Movie:   NewMovieService(repo, log),
		Cinema:  NewCinemaService(repo, log),
		Booking: NewBookingService(repo, log),
		Review:  NewReviewService(repo, log),
	}
}
