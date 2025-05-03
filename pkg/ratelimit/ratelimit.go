// Package ratelimit provides rate limiting functionality for external service queries.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter is an interface for rate limiting operations.
type RateLimiter interface {
	// Wait blocks until the rate limit allows an operation to proceed.
	Wait(ctx context.Context) error
	// Allow returns true if the operation is allowed to proceed.
	Allow() bool
}

// TokenBucketLimiter implements a token bucket rate limiter.
type TokenBucketLimiter struct {
	rate       float64     // tokens per second
	bucketSize float64     // maximum number of tokens
	tokens     float64     // current number of tokens
	lastTime   time.Time   // last time tokens were added
	mu         sync.Mutex  // mutex for thread safety
}

// NewTokenBucketLimiter creates a new token bucket rate limiter.
func NewTokenBucketLimiter(rate float64, bucketSize float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     bucketSize, // Start with a full bucket
		lastTime:   time.Now(),
	}
}

// Allow returns true if the operation is allowed to proceed.
func (l *TokenBucketLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	// Add tokens based on elapsed time
	l.tokens += elapsed * l.rate
	if l.tokens > l.bucketSize {
		l.tokens = l.bucketSize
	}

	// Check if we have enough tokens
	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true
	}

	return false
}

// Wait blocks until the rate limit allows an operation to proceed.
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to allow the operation
		if l.Allow() {
			return nil
		}

		// Calculate how long to wait
		l.mu.Lock()
		waitTime := time.Duration((1.0 - l.tokens) / l.rate * float64(time.Second))
		l.mu.Unlock()

		// Wait for a short time or until the context is done
		timer := time.NewTimer(waitTime)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue and try again
		}
	}
}

// ServiceLimiter manages rate limiters for different services.
type ServiceLimiter struct {
	limiters map[string]RateLimiter
	mu       sync.RWMutex
}

// NewServiceLimiter creates a new service limiter.
func NewServiceLimiter() *ServiceLimiter {
	return &ServiceLimiter{
		limiters: make(map[string]RateLimiter),
	}
}

// GetLimiter returns a rate limiter for the specified service.
func (s *ServiceLimiter) GetLimiter(service string, rate float64, bucketSize float64) RateLimiter {
	s.mu.RLock()
	limiter, ok := s.limiters[service]
	s.mu.RUnlock()

	if ok {
		return limiter
	}

	// Create a new limiter if one doesn't exist
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check again in case another goroutine created the limiter
	limiter, ok = s.limiters[service]
	if ok {
		return limiter
	}

	limiter = NewTokenBucketLimiter(rate, bucketSize)
	s.limiters[service] = limiter
	return limiter
}

// Wait blocks until the rate limit for the specified service allows an operation to proceed.
func (s *ServiceLimiter) Wait(ctx context.Context, service string, rate float64, bucketSize float64) error {
	limiter := s.GetLimiter(service, rate, bucketSize)
	return limiter.Wait(ctx)
}

// Allow returns true if the operation for the specified service is allowed to proceed.
func (s *ServiceLimiter) Allow(service string, rate float64, bucketSize float64) bool {
	limiter := s.GetLimiter(service, rate, bucketSize)
	return limiter.Allow()
}

// DefaultLimiter is the default service limiter.
var DefaultLimiter = NewServiceLimiter()

// Wait blocks until the rate limit for the specified service allows an operation to proceed.
func Wait(ctx context.Context, service string, rate float64, bucketSize float64) error {
	return DefaultLimiter.Wait(ctx, service, rate, bucketSize)
}

// Allow returns true if the operation for the specified service is allowed to proceed.
func Allow(service string, rate float64, bucketSize float64) bool {
	return DefaultLimiter.Allow(service, rate, bucketSize)
}

// RateLimitedFunc wraps a function with rate limiting.
func RateLimitedFunc(ctx context.Context, service string, rate float64, bucketSize float64, fn func() error) error {
	err := Wait(ctx, service, rate, bucketSize)
	if err != nil {
		return fmt.Errorf("rate limit wait error: %w", err)
	}
	return fn()
}