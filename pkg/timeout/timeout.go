// Package timeout provides timeout handling functionality for the MXToolbox clone.
package timeout

import (
	"context"
	"time"

	"mxclone/pkg/errors"
)

// DefaultTimeout is the default timeout for network operations.
const DefaultTimeout = 10 * time.Second

// WithTimeout executes a function with a timeout.
func WithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create a channel to receive the result
	done := make(chan error, 1)

	// Execute the function in a goroutine
	go func() {
		done <- fn(ctx)
	}()

	// Wait for the function to complete or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return errors.New(errors.ErrorTypeTimeout, "operation timed out")
		}
		return ctx.Err()
	}
}

// WithTimeoutResult executes a function with a timeout and returns a result.
func WithTimeoutResult[T any](ctx context.Context, timeout time.Duration, fn func(context.Context) (T, error)) (T, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create a channel to receive the result
	type result struct {
		value T
		err   error
	}
	done := make(chan result, 1)

	// Execute the function in a goroutine
	go func() {
		value, err := fn(ctx)
		done <- result{value, err}
	}()

	// Wait for the function to complete or timeout
	select {
	case res := <-done:
		return res.value, res.err
	case <-ctx.Done():
		var zero T
		if ctx.Err() == context.DeadlineExceeded {
			return zero, errors.New(errors.ErrorTypeTimeout, "operation timed out")
		}
		return zero, ctx.Err()
	}
}

// WithRetry executes a function with retries and timeout.
func WithRetry(ctx context.Context, timeout time.Duration, retries int, fn func(context.Context) error) error {
	var err error
	for i := 0; i <= retries; i++ {
		err = WithTimeout(ctx, timeout, fn)
		if err == nil {
			return nil
		}

		// If this was the last retry, break
		if i == retries {
			break
		}

		// Wait before retrying (exponential backoff)
		backoffDuration := time.Duration(1<<uint(i)) * 100 * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffDuration):
			// Continue with retry
		}
	}

	return errors.Wrap(err, errors.ErrorTypeTimeout, "operation failed after retries")
}

// WithRetryResult executes a function with retries and timeout and returns a result.
func WithRetryResult[T any](ctx context.Context, timeout time.Duration, retries int, fn func(context.Context) (T, error)) (T, error) {
	var result T
	var err error
	for i := 0; i <= retries; i++ {
		result, err = WithTimeoutResult(ctx, timeout, fn)
		if err == nil {
			return result, nil
		}

		// If this was the last retry, break
		if i == retries {
			break
		}

		// Wait before retrying (exponential backoff)
		backoffDuration := time.Duration(1<<uint(i)) * 100 * time.Millisecond
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		case <-time.After(backoffDuration):
			// Continue with retry
		}
	}

	return result, errors.Wrap(err, errors.ErrorTypeTimeout, "operation failed after retries")
}