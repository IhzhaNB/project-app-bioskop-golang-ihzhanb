package response

import "cinema-booking/internal/data/entity"

type GenreResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Helper converter
func GenreToResponse(genre *entity.Genre) GenreResponse {
	return GenreResponse{
		ID:   genre.ID.String(),
		Name: genre.Name,
	}
}
