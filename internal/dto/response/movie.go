package response

import (
	"cinema-booking/internal/data/entity"
	"fmt"
	"time"
)

type MovieResponse struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Description       *string   `json:"description,omitempty"`
	PosterURL         *string   `json:"poster_url,omitempty"`
	Rating            float64   `json:"rating"`
	ReviewCount       int       `json:"review_count"`
	ReleaseDate       string    `json:"release_date"`
	DurationInMinutes string    `json:"duration_in_minutes"`
	Genres            []string  `json:"genres"`
	ReleaseStatus     string    `json:"release_status"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
}

type MovieDetailResponse struct {
	MovieResponse
	Description *string    `json:"description,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// Helper converters
func MovieToResponse(movie *entity.Movie, genres []string, reviewCount int) MovieResponse {
	// Format duration sebagai string
	durationStr := fmt.Sprintf("%d", movie.DurationInMinutes)

	// Format release status sesuai requirement
	var statusStr string
	switch movie.ReleaseStatus {
	case entity.ReleaseStatusNowPlaying:
		statusStr = "now"
	case entity.ReleaseStatusComingSoon:
		statusStr = "coming_soon"
	default:
		statusStr = string(movie.ReleaseStatus)
	}

	return MovieResponse{
		ID:                movie.ID.String(),
		Title:             movie.Title,
		Description:       movie.Description,
		PosterURL:         movie.PosterURL,
		Rating:            movie.Rating,
		ReviewCount:       reviewCount,
		ReleaseDate:       movie.ReleaseDate.Format("2006-01-02"),
		DurationInMinutes: durationStr,
		Genres:            genres,
		ReleaseStatus:     statusStr,
		CreatedAt:         movie.CreatedAt,
	}
}

func MovieToDetailResponse(movie *entity.Movie, genres []string, reviewCount int) MovieDetailResponse {
	movieResp := MovieToResponse(movie, genres, reviewCount)
	return MovieDetailResponse{
		MovieResponse: movieResp,
		Description:   movie.Description,
		UpdatedAt:     &movie.UpdatedAt,
	}
}
