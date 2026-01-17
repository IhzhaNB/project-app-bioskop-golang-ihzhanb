package entity

import (
	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusExpired   BookingStatus = "expired"
)

type Booking struct {
	Base
	OrderID    string        `db:"order_id"`
	UserID     uuid.UUID     `db:"user_id"`
	ScheduleID uuid.UUID     `db:"schedule_id"`
	TotalSeats int           `db:"total_seats"`
	TotalPrice float64       `db:"total_price"`
	Status     BookingStatus `db:"status"`
}
