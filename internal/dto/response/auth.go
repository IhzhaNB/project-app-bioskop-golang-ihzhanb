package response

import (
	"cinema-booking/internal/data/entity"
	"time"
)

type AuthResponse struct {
	UserID     string          `json:"user_id"`
	Token      string          `json:"token"`
	ExpiresAt  time.Time       `json:"expires_at"`
	Email      string          `json:"email"`
	Username   string          `json:"username"`
	Role       entity.UserRole `json:"role"`
	IsVerified bool            `json:"is_verified"`
}

type UserResponse struct {
	ID         string          `json:"id"`
	Username   string          `json:"username"`
	Email      string          `json:"email"`
	Phone      *string         `json:"phone,omitempty"`
	Role       entity.UserRole `json:"role"`
	IsVerified bool            `json:"is_verified"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Helper converters
func UserToResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		Phone:      user.Phone,
		Role:       user.Role,
		IsVerified: user.EmailVerified,
		CreatedAt:  user.CreatedAt,
	}
}

func AuthToResponse(user *entity.User, session *entity.Session) AuthResponse {
	resp := AuthResponse{
		UserID:     user.ID.String(),
		Email:      user.Email,
		Username:   user.Username,
		Role:       user.Role,
		IsVerified: user.EmailVerified,
	}

	if session != nil {
		resp.Token = session.Token.String()
		resp.ExpiresAt = session.ExpiresAt
	}

	return resp
}
