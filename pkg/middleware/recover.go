package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

// Recover middleware
func Recover(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("PANIC recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.Stack("stack"),
					)

					// Return internal server error
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"status":false,"message":"Internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
