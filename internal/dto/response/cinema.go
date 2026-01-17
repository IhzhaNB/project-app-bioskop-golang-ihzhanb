package response

import (
	"cinema-booking/internal/data/entity"
	"time"
)

type CinemaResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CinemaDetailResponse struct {
	CinemaResponse
	Halls []HallResponse `json:"halls,omitempty"`
}

type HallResponse struct {
	ID         string `json:"id"`
	HallNumber int    `json:"hall_number"`
	TotalSeats int    `json:"total_seats"`
}

type SeatResponse struct {
	ID          string `json:"id"`
	SeatNumber  string `json:"seat_number"`
	SeatRow     string `json:"seat_row"`
	SeatColumn  int    `json:"seat_column"`
	IsAvailable bool   `json:"is_available"`
}

type SeatAvailabilityResponse struct {
	HallID string         `json:"hall_id"`
	Date   string         `json:"date"`
	Time   string         `json:"time"`
	Seats  []SeatResponse `json:"seats"`
}

// Helper converters
func CinemaToResponse(cinema *entity.Cinema) CinemaResponse {
	return CinemaResponse{
		ID:        cinema.ID.String(),
		Name:      cinema.Name,
		Location:  cinema.Location,
		City:      cinema.City,
		CreatedAt: cinema.CreatedAt,
		UpdatedAt: cinema.UpdatedAt,
	}
}

func HallToResponse(hall *entity.Hall) HallResponse {
	return HallResponse{
		ID:         hall.ID.String(),
		HallNumber: hall.HallNumber,
		TotalSeats: hall.TotalSeats,
	}
}

func SeatToResponse(seat *entity.Seat) SeatResponse {
	return SeatResponse{
		ID:          seat.ID.String(),
		SeatNumber:  seat.SeatNumber,
		SeatRow:     seat.SeatRow,
		SeatColumn:  seat.SeatColumn,
		IsAvailable: seat.IsAvailable,
	}
}
