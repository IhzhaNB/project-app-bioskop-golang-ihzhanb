package entity

import "github.com/google/uuid"

type BookingSeat struct {
	BaseSimple
	BookingID uuid.UUID `db:"booking_id"`
	SeatID    uuid.UUID `db:"seat_id"`
}
