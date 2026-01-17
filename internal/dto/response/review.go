package response

import (
	"cinema-booking/internal/data/entity"
	"time"
)

type ReviewResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username,omitempty"`
	MovieID    string    `json:"movie_id"`
	MovieTitle string    `json:"movie_title,omitempty"`
	Rating     int       `json:"rating"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type MovieReviewStats struct {
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int64   `json:"review_count"`
}

// Helper converter
func ReviewToResponse(review *entity.Review, username, movieTitle string) ReviewResponse {
	return ReviewResponse{
		ID:         review.ID.String(),
		UserID:     review.UserID.String(),
		Username:   username,
		MovieID:    review.MovieID.String(),
		MovieTitle: movieTitle,
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
	}
}
