package request

type CreateBookingRequest struct {
	ScheduleID      string   `json:"schedule_id" validate:"required,uuid4"`
	SeatIDs         []string `json:"seat_ids" validate:"required,min=1,dive,uuid4"`
	PaymentMethodID string   `json:"payment_method_id" validate:"required,uuid4"`
}

type ProcessPaymentRequest struct {
	BookingID       string  `json:"booking_id" validate:"required,uuid4"`
	PaymentMethodID string  `json:"payment_method_id" validate:"required,uuid4"`
	Amount          float64 `json:"amount" validate:"required,min=1000"`
	TransactionID   *string `json:"transaction_id,omitempty"`
}
