package usecase

import (
	"context"
	"fmt"

	"cinema-booking/internal/data/entity"
	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/dto/response"
	"cinema-booking/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService interface {
	GetProfile(ctx context.Context, userID string) (*response.UserResponse, error)
	GetAllUsers(ctx context.Context, req *request.PaginatedRequest) (*response.PaginatedResponse[response.UserResponse], error)
	DeleteUser(ctx context.Context, userID string) error
}

type userService struct {
	userRepo repository.UserRepository
	log      *zap.Logger
}

func NewUserService(userRepo repository.UserRepository, log *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		log:      log,
	}
}

func (us *userService) GetProfile(ctx context.Context, userID string) (*response.UserResponse, error) {
	// Parse userID
	id, err := uuid.Parse(userID)
	if err != nil {
		us.log.Warn("Invalid user ID", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("invalid user ID")
	}

	// Find user
	user, err := us.userRepo.FindByID(ctx, id)
	if err != nil {
		us.log.Error("Failed to find user", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get profile")
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Build response
	return &response.UserResponse{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		Phone:      user.Phone,
		Role:       entity.UserRole(user.Role),
		IsVerified: user.EmailVerified,
		CreatedAt:  user.CreatedAt,
	}, nil
}

func (us *userService) GetAllUsers(ctx context.Context, req *request.PaginatedRequest) (*response.PaginatedResponse[response.UserResponse], error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 10
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	// Get users with pagination
	users, err := us.userRepo.FindAll(ctx, req.Page, req.PerPage)
	if err != nil {
		us.log.Error("Failed to get all users",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
		)
		return nil, fmt.Errorf("failed to get users")
	}

	// Get total count
	total, err := us.userRepo.CountAll(ctx)
	if err != nil {
		us.log.Error("Failed to count users", zap.Error(err))
		return nil, fmt.Errorf("failed to count users")
	}

	// Convert to response
	userResponses := make([]response.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = response.UserResponse{
			ID:         user.ID.String(),
			Username:   user.Username,
			Email:      user.Email,
			Phone:      user.Phone,
			Role:       entity.UserRole(user.Role),
			IsVerified: user.EmailVerified,
			CreatedAt:  user.CreatedAt,
		}
	}

	// Calculate pagination
	totalPages := utils.CalculateTotalPages(total, req.PerPage)

	us.log.Info("Users retrieved",
		zap.Int("count", len(users)),
		zap.Int64("total", total),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
		zap.Int("total_pages", totalPages),
	)

	return response.NewPaginatedResponse(userResponses, req.Page, req.PerPage, total), nil
}

func (us *userService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	user, err := us.userRepo.FindByID(ctx, id)
	if err != nil {
		us.log.Error("Failed to get user for delete", zap.Error(err), zap.String("id", userID))
		return fmt.Errorf("user not found")
	}

	if err := us.userRepo.Delete(ctx, id); err != nil {
		us.log.Error("Failed to delete user", zap.Error(err), zap.String("id", userID))
		return fmt.Errorf("failed to delete user")
	}

	us.log.Info("User deleted", zap.String("user_id", id.String()), zap.String("email", user.Email))
	return nil
}
