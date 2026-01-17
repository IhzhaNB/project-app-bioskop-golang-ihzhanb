package wire

import (
	"cinema-booking/internal/adaptor"
	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/middleware"
	"cinema-booking/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func wireBooking(
	r chi.Router,
	bookingHandler *adaptor.BookingHandler,
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) {
	// ==================== PROTECTED ROUTES (require auth) ====================
	// Group routes that require authentication
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthSession(repo.Session, log))

		// POST /api/booking - Create new booking (authenticated users only)
		r.Post("/api/booking", bookingHandler.CreateBooking)

		// GET /api/user/bookings - View booking history (user's own bookings)
		r.Get("/api/user/bookings", bookingHandler.GetUserBookings)

		// POST /api/pay - Process payment for booking
		r.Post("/api/pay", bookingHandler.ProcessPayment)
	})

	// ==================== PUBLIC ROUTES ====================
	// GET /api/payment-methods - List available payment methods (public)
	r.Get("/api/payment-methods", bookingHandler.GetPaymentMethods)

	// ==================== ADMIN ROUTES ====================
	// Admin booking management routes
	r.Route("/api/admin/bookings", func(r chi.Router) {
		// Require both authentication AND admin role
		r.Use(middleware.AuthSession(repo.Session, log))
		r.Use(middleware.Admin(repo.User, log))

		// GET /api/admin/bookings/{id} - View any booking details (admin)
		r.Get("/{id}", bookingHandler.GetBookingByID)

		// PUT /api/admin/bookings/{id}/cancel - Cancel any booking (admin)
		r.Put("/{id}/cancel", bookingHandler.CancelBooking)
	})
}
