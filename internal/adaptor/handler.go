package adaptor

import (
	"cinema-booking/internal/usecase"

	"go.uber.org/zap"
)

type Handler struct {
	Auth  *AuthHandler
	User  *UserHandler
	Movie *MovieHandler
}

func NewHandler(service *usecase.Service, log *zap.Logger) *Handler {
	return &Handler{
		Auth:  NewAuthHandler(service.Auth, log),
		User:  NewUserHandler(service.User, log),
		Movie: NewMovieHandler(service.Movie, log),
	}
}
