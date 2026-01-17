package entity

import (
	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type Payment struct {
	Base
	BookingID       uuid.UUID     `db:"booking_id"`
	PaymentMethodID uuid.UUID     `db:"payment_method_id"`
	Amount          float64       `db:"amount"`
	Status          PaymentStatus `db:"status"`
	TransactionID   *string       `db:"transaction_id"`
}
