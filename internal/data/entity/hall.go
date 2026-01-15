package entity

import "github.com/google/uuid"

type Hall struct {
	Base
	CinemaID   uuid.UUID `db:"cinema_id"`
	HallNumber int       `db:"hall_number"`
	TotalSeats int       `db:"total_seats"`
}
