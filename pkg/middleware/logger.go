package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Custom response writer to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// Override WriteHeader to capture status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Override Write to count response bytes
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Logger middleware factory
// Returns a middleware function that logs HTTP requests
func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timer
			start := time.Now()

			// Wrap response writer to capture metrics
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status
			}

			// Process request through next handler
			next.ServeHTTP(rw, r)

			// Calculate request duration
			duration := time.Since(start)

			// Log request details
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", rw.statusCode),
				zap.Int("bytes", rw.bytesWritten),
				zap.Duration("duration", duration),
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}
