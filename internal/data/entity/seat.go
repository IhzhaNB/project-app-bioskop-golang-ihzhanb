package entity

import "github.com/google/uuid"

type Seat struct {
	Base
	HallID      uuid.UUID `db:"hall_id"`
	SeatNumber  string    `db:"seat_number"` // A1, A2, B1, etc.
	SeatRow     string    `db:"seat_row"`    // A, B, C, etc.
	SeatColumn  int       `db:"seat_column"` // 1, 2, 3, etc.
	IsAvailable bool      `db:"is_available"`
}
