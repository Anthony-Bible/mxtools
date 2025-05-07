package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mxclone/internal/api/errors"
	"mxclone/internal/config"
	"mxclone/pkg/logging"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter represents a middleware adapter for rate limiting
type RateLimiter struct {
	requestsPerMinute int
	burstSize         int
	cleanupInterval   time.Duration
	store             map[string][]time.Time
	mu                sync.Mutex
	logger            *logging.Logger
	errorHandler      *errors.ErrorHandler
}

// NewRateLimiter creates a new rate limiting middleware with default settings
func NewRateLimiter(logger *logging.Logger) *RateLimiter {
	apiConfig := config.NewAPIConfig()
	return &RateLimiter{
		requestsPerMinute: apiConfig.RateLimitRequestsPerMinute,
		burstSize:         apiConfig.RateLimitBurstSize,
		cleanupInterval:   apiConfig.RateLimitCleanupInterval,
		store:             make(map[string][]time.Time),
		logger:            logger,
	}
}

// WithErrorHandler sets the error handler for the rate limiter
func (rl *RateLimiter) WithErrorHandler(errorHandler *errors.ErrorHandler) *RateLimiter {
	rl.errorHandler = errorHandler
	return rl
}

// WithConfig allows custom configuration of the rate limiter
func (rl *RateLimiter) WithConfig(requestsPerMinute, burstSize int, cleanupInterval time.Duration) *RateLimiter {
	rl.requestsPerMinute = requestsPerMinute
	rl.burstSize = burstSize
	rl.cleanupInterval = cleanupInterval
	return rl
}

// Start begins the background cleanup routine
func (rl *RateLimiter) Start() {
	go rl.startCleanupRoutine()
}

// startCleanupRoutine periodically cleans up old entries
func (rl *RateLimiter) startCleanupRoutine() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes old entries from the store
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, times := range rl.store {
		var recent []time.Time
		for _, t := range times {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}

		if len(recent) == 0 {
			// No recent requests, remove the IP from the store
			delete(rl.store, ip)
		} else {
			// Update with only recent requests
			rl.store[ip] = recent
		}
	}

	rl.logger.Debug("Rate limit store cleaned up, %d IPs in store", len(rl.store))
}

// getClientIP extracts the client IP from the request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for clients behind proxy)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// The first IP in the list is the client's IP
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}

	// Then check X-Real-IP
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, use the whole RemoteAddr
		return r.RemoteAddr
	}
	return ip
}

// Limit is a middleware that limits requests per IP
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getClientIP(r)
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

		// Apply burst allowance - allow a burst of requests up to burstSize
		burstAllowed := rl.burstSize > 0 && len(recent) < rl.burstSize

		// Check if rate limit exceeded
		if len(recent) >= rl.requestsPerMinute && !burstAllowed {
			rl.mu.Unlock()
			rl.logger.Info("Rate limit exceeded for IP: %s (%d requests in last minute)", ip, len(recent))

			// Use error handler if available
			if rl.errorHandler != nil {
				rl.errorHandler.HandleRateLimitError(w)
				return
			}

			// Fall back to simple error response
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "rate limit exceeded",
				"message": "too many requests, please try again later",
			})
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

// Validator represents a middleware adapter for request validation
type Validator struct {
	logger *logging.Logger
}

// NewValidator creates a new validation middleware
func NewValidator(logger *logging.Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidateJSON is a middleware that validates JSON request bodies
func (v *Validator) ValidateJSON(next http.HandlerFunc, validate func([]byte) (bool, map[string]interface{})) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only apply to POST, PUT, PATCH methods
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			v.logger.Error("Error reading request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Restore body for future handlers
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Validate the body
		valid, errors := validate(body)
		if !valid {
			v.logger.Info("Request validation failed: %v", errors)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":       "Validation failed",
				"validations": errors,
			})
			return
		}

		// If validation passes, store the validated body in the context for later use
		ctx := context.WithValue(r.Context(), "validatedBody", body)
		r = r.WithContext(ctx)

		// Call next handler
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

// ValidateParams is a middleware that validates URL parameters
func (v *Validator) ValidateParams(next http.HandlerFunc, validate func(map[string]string) (bool, map[string]interface{})) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract URL parameters (this assumes you're using a router that places params in context)
		params := extractParams(r)

		// Validate the parameters
		valid, errors := validate(params)
		if !valid {
			v.logger.Info("Parameter validation failed: %v", errors)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":       "Validation failed",
				"validations": errors,
			})
			return
		}

		// Call next handler
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

// Helper function to extract parameters - implementation depends on router
func extractParams(r *http.Request) map[string]string {
	// This is a placeholder implementation
	// The actual implementation would depend on your router
	// For example, if using gorilla/mux:
	// return mux.Vars(r)

	// For now, just returning query parameters
	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return params
}
