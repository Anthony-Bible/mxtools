package middleware

import (
	"encoding/json"
	"mxclone/pkg/logging"
	"net/http"
	"sync"
	"time"
)

// RateLimiter represents a middleware adapter for rate limiting
type RateLimiter struct {
	requestsPerMinute int
	store             map[string][]time.Time
	mu                sync.Mutex
	logger            *logging.Logger
}

// NewRateLimiter creates a new rate limiting middleware
func NewRateLimiter(requestsPerMinute int, logger *logging.Logger) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		store:             make(map[string][]time.Time),
		logger:            logger,
	}
}

// Limit is a middleware that limits requests per IP
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		rl.mu.Lock()

		// Get recent requests from this IP
		times := rl.store[ip]
		now := time.Now()
		var recent []time.Time

		// Filter to keep only requests within the last minute
		for _, t := range times {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}

		// Check if rate limit exceeded
		if len(recent) >= rl.requestsPerMinute {
			rl.mu.Unlock()
			rl.logger.Info("rate limit exceeded for IP: %s", ip)
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
			return
		}

		// Add this request to the store
		recent = append(recent, now)
		rl.store[ip] = recent
		rl.mu.Unlock()

		// Call the next handler
		next(w, r)
	}
}

// Logger represents a middleware adapter for logging
type Logger struct {
	logger *logging.Logger
}

// NewLogger creates a new logging middleware
func NewLogger(logger *logging.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

// Log is a middleware that logs HTTP requests
func (l *Logger) Log(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next(rw, r)

		// Log request details
		duration := time.Since(start)
		l.logger.Info("%s %s %s %d %s", r.RemoteAddr, r.Method, r.URL.Path, rw.statusCode, duration)
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and calls the underlying ResponseWriter's WriteHeader
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
