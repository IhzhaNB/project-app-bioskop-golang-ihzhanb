package usecase

import (
	"context"
	"fmt"

	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/dto/response"

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

// GetProfile retrieves user information for the authenticated user
func (us *userService) GetProfile(ctx context.Context, userID string) (*response.UserResponse, error) {
	// Parse string userID to UUID format
	id, err := uuid.Parse(userID)
	if err != nil {
		us.log.Warn("Invalid user ID format", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	// Find user
	user, err := us.userRepo.FindByID(ctx, id)
	if err != nil {
		us.log.Error("Failed to find user", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("find user profile %s: %w", userID, err)
	}
	if user == nil {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	// Build response
	userResp := response.UserToResponse(user)
	return &userResp, nil
}

func (us *userService) GetAllUsers(ctx context.Context, req *request.PaginatedRequest) (*response.PaginatedResponse[response.UserResponse], error) {
	// Calculate pagination parameters using helper methods
	limit := req.Limit()   // Default: 10, Max: 100
	offset := req.Offset() // (page-1) * per_page

	// Get users with pagination
	users, err := us.userRepo.FindAll(ctx, limit, offset)
	if err != nil {
		us.log.Error("Failed to get all users",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
		)
		return nil, fmt.Errorf("get all users page %d per_page %d: %w", req.Page, req.PerPage, err)
	}

	// Get total count of users for pagination metadata
	total, err := us.userRepo.CountAll(ctx)
	if err != nil {
		us.log.Error("Failed to count users", zap.Error(err))
		return nil, fmt.Errorf("count all users: %w", err)
	}

	// Convert to response
	userResponses := make([]response.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = response.UserToResponse(user)
	}

	// Create paginated response seperti movie_service
	paginatedResp := response.NewPaginatedResponse(userResponses, req.Page, req.PerPage, total)

	us.log.Info("Users retrieved",
		zap.Int("count", len(users)),
		zap.Int64("total", total),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Int("total_pages", paginatedResp.Pagination.TotalPages),
	)

	return paginatedResp, nil
}

func (us *userService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format %s: %w", userID, err)
	}

	user, err := us.userRepo.FindByID(ctx, id)
	if err != nil {
		us.log.Error("Failed to get user for delete", zap.Error(err), zap.String("id", userID))
		return fmt.Errorf("find user for delete %s: %w", userID, err)
	}

	if user == nil {
		return fmt.Errorf("user %s not found", userID)
	}

	if err := us.userRepo.Delete(ctx, id); err != nil {
		us.log.Error("Failed to delete user", zap.Error(err), zap.String("id", userID))
		return fmt.Errorf("delete user %s: %w", userID, err)
	}

	us.log.Info("User deleted",
		zap.String("user_id", id.String()),
		zap.String("email", user.Email),
		zap.String("username", user.Username),
	)
	return nil
}
