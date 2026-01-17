package response

import (
	"cinema-booking/internal/data/entity"
	"time"
)

type PaymentMethodResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type BookingResponse struct {
	ID          string               `json:"id"`
	OrderID     string               `json:"order_id"`
	UserID      string               `json:"user_id"`
	ScheduleID  string               `json:"schedule_id"`
	MovieTitle  string               `json:"movie_title,omitempty"`
	CinemaName  string               `json:"cinema_name,omitempty"`
	HallNumber  int                  `json:"hall_number,omitempty"`
	ShowDate    string               `json:"show_date,omitempty"`
	ShowTime    string               `json:"show_time,omitempty"`
	TotalSeats  int                  `json:"total_seats"`
	TotalPrice  float64              `json:"total_price"`
	Status      entity.BookingStatus `json:"status"`
	SeatNumbers []string             `json:"seat_numbers,omitempty"`
	Payment     *PaymentResponse     `json:"payment,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

type PaymentResponse struct {
	ID            string                `json:"id"`
	BookingID     string                `json:"booking_id"`
	PaymentMethod PaymentMethodResponse `json:"payment_method"`
	Amount        float64               `json:"amount"`
	Status        entity.PaymentStatus  `json:"status"`
	TransactionID *string               `json:"transaction_id,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
}

type BookingDetailResponse struct {
	BookingResponse
	ScheduleDetails ScheduleDetails `json:"schedule_details"`
}

type ScheduleDetails struct {
	MovieTitle string  `json:"movie_title"`
	CinemaName string  `json:"cinema_name"`
	HallNumber int     `json:"hall_number"`
	ShowDate   string  `json:"show_date"`
	ShowTime   string  `json:"show_time"`
	Price      float64 `json:"price"`
}

// Helper converters
func PaymentMethodToResponse(pm *entity.PaymentMethod) PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:       pm.ID.String(),
		Name:     pm.Name,
		IsActive: pm.IsActive,
	}
}

func PaymentToResponse(payment *entity.Payment, paymentMethod *entity.PaymentMethod) PaymentResponse {
	return PaymentResponse{
		ID:            payment.ID.String(),
		BookingID:     payment.BookingID.String(),
		PaymentMethod: PaymentMethodToResponse(paymentMethod),
		Amount:        payment.Amount,
		Status:        payment.Status,
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
	}
}
