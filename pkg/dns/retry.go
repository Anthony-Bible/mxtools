// Package dns provides DNS lookup functionality.
package dns

import (
	"context"
	"fmt"
	"time"

	"mxclone/pkg/types"
)

// LookupWithRetry performs a DNS lookup with retry logic.
func LookupWithRetry(ctx context.Context, domain string, recordType string, maxRetries int, timeout time.Duration) (*types.DNSResult, error) {
	var result *types.DNSResult
	var err error
	var retryCount int

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Retry loop
	for retryCount = 0; retryCount <= maxRetries; retryCount++ {
		// Check if context is done
		select {
		case <-ctxWithTimeout.Done():
			if ctxWithTimeout.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("DNS lookup timed out after %s", timeout)
			}
			return nil, ctxWithTimeout.Err()
		default:
			// Continue with the lookup
		}

		// Perform the lookup
		result, err = Lookup(ctxWithTimeout, domain, recordType)
		if err == nil {
			// Successful lookup
			return result, nil
		}

		// If this was the last retry, break
		if retryCount == maxRetries {
			break
		}

		// Wait before retrying (exponential backoff)
		backoffDuration := time.Duration(1<<uint(retryCount)) * 100 * time.Millisecond
		select {
		case <-ctxWithTimeout.Done():
			return nil, ctxWithTimeout.Err()
		case <-time.After(backoffDuration):
			// Continue with retry
		}
	}

	// All retries failed
	return result, fmt.Errorf("DNS lookup failed after %d retries: %w", retryCount, err)
}

// AdvancedLookupWithRetry performs an advanced DNS lookup with retry logic.
func AdvancedLookupWithRetry(ctx context.Context, domain string, recordType string, server string, maxRetries int, timeout time.Duration) (*types.DNSResult, error) {
	var result *types.DNSResult
	var err error
	var retryCount int

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Retry loop
	for retryCount = 0; retryCount <= maxRetries; retryCount++ {
		// Check if context is done
		select {
		case <-ctxWithTimeout.Done():
			if ctxWithTimeout.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("DNS lookup timed out after %s", timeout)
			}
			return nil, ctxWithTimeout.Err()
		default:
			// Continue with the lookup
		}

		// Perform the lookup
		result, err = AdvancedLookup(ctxWithTimeout, domain, recordType, server)
		if err == nil {
			// Successful lookup
			return result, nil
		}

		// If this was the last retry, break
		if retryCount == maxRetries {
			break
		}

		// Wait before retrying (exponential backoff)
		backoffDuration := time.Duration(1<<uint(retryCount)) * 100 * time.Millisecond
		select {
		case <-ctxWithTimeout.Done():
			return nil, ctxWithTimeout.Err()
		case <-time.After(backoffDuration):
			// Continue with retry
		}
	}

	// All retries failed
	return result, fmt.Errorf("DNS lookup failed after %d retries: %w", retryCount, err)
}