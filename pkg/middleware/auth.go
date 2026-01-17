package middleware

import (
	"net/http"
	"strings"

	"cinema-booking/internal/data/repository"
	"cinema-booking/pkg/utils"

	"go.uber.org/zap"
)

// AuthSession middleware untuk validasi session token UUID
func AuthSession(sessionRepo repository.SessionRepository, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// // Skip auth untuk public routes
			// if isPublicRoute(r.URL.Path, r.Method) { // PASS METHOD!
			// 	next.ServeHTTP(w, r)
			// 	return
			// }

			// Extract token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ResponseUnauthorized(w, "Missing authorization token")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer " {
				utils.ResponseUnauthorized(w, "Invalid token format. Use: Bearer <token>")
				return
			}

			token := parts[1]

			// Find valid session
			session, err := sessionRepo.FindValidSession(r.Context(), token)
			if err != nil {
				logger.Error("Failed to validate session",
					zap.String("token", token),
					zap.Error(err))
				utils.ResponseInternalError(w, "Internal server error")
				return
			}

			if session == nil {
				logger.Warn("Invalid or expired session", zap.String("token", token))
				utils.ResponseUnauthorized(w, "Invalid or expired session")
				return
			}

			// Set context dengan user info DAN token
			ctx := utils.SetUserContext(r.Context(), session.UserID, "customer")
			ctx = utils.SetTokenContext(ctx, token) // SET TOKEN!

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Admin - middleware cek role admin
func Admin(userRepo repository.UserRepository, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Get user ID dari context (sudah diset AuthSession)
			userID, ok := utils.GetUserIDFromContext(r.Context())
			if !ok {
				utils.ResponseUnauthorized(w, "Authentication required")
				return
			}

			// 2. Get user dari repo
			user, err := userRepo.FindByID(r.Context(), userID)
			if err != nil {
				logger.Error("Admin check: failed to get user",
					zap.Error(err), zap.String("user_id", userID.String()))
				utils.ResponseInternalError(w, "Internal server error")
				return
			}

			// 3. Check if admin
			if user == nil || user.Role != "admin" {
				logger.Warn("Admin check: non-admin access attempt",
					zap.String("user_id", userID.String()),
					zap.String("path", r.URL.Path))
				utils.ResponseForbidden(w, "Admin access required")
				return
			}

			// 4. Lanjut ke handler
			next.ServeHTTP(w, r)
		})
	}
}

// // Helper untuk cek public routes (TERIMA METHOD PARAMETER!)
// func isPublicRoute(path, method string) bool {
// 	publicRoutes := map[string][]string{
// 		"/api/register":        {"POST"},
// 		"/api/login":           {"POST"},
// 		"/api/cinemas":         {"GET"},
// 		"/api/movies":          {"GET"},
// 		"/api/payment-methods": {"GET"},
// 		"/health":              {"GET"},
// 		"/api/send-otp":        {"POST"},
// 		"/api/verify-email":    {"POST"},
// 	}

// 	// Check exact path
// 	if methods, exists := publicRoutes[path]; exists {
// 		for _, m := range methods {
// 			if m == method {
// 				return true
// 			}
// 		}
// 	}

// 	// Pattern match: /api/cinemas/{id} (GET only)
// 	if strings.HasPrefix(path, "/api/cinemas/") && method == "GET" {
// 		parts := strings.Split(path, "/")
// 		if len(parts) == 4 && parts[3] != "seats" {
// 			// /api/cinemas/{id} → public
// 			return true
// 		}
// 	}

// 	// Pattern match: /api/cinemas/{id}/seats (GET with query params is public for checking availability)
// 	if strings.Contains(path, "/seats") && method == "GET" {
// 		// Checking seat availability bisa public
// 		return true
// 	}

// 	return false
// }
