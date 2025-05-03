// Package dns provides DNS lookup functionality.
package dns

import (
	"context"
	"testing"
	"time"
)

// TestIntegrationLookup tests the Lookup function with real DNS queries.
func TestIntegrationLookup(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		domain     string
		recordType string
		wantErr    bool
	}{
		{
			name:       "Valid A record lookup",
			domain:     "example.com",
			recordType: "A",
			wantErr:    false,
		},
		{
			name:       "Valid AAAA record lookup",
			domain:     "example.com",
			recordType: "AAAA",
			wantErr:    false,
		},
		{
			name:       "Valid MX record lookup",
			domain:     "google.com",
			recordType: "MX",
			wantErr:    false,
		},
		{
			name:       "Valid TXT record lookup",
			domain:     "google.com",
			recordType: "TXT",
			wantErr:    false,
		},
		{
			name:       "Valid NS record lookup",
			domain:     "example.com",
			recordType: "NS",
			wantErr:    false,
		},
		{
			name:       "Invalid domain",
			domain:     "thisdoesnotexist.example",
			recordType: "A",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the function being tested
			result, err := Lookup(ctx, tt.domain, tt.recordType)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("Lookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check that we got some results
			if !tt.wantErr {
				if result == nil {
					t.Errorf("Lookup() returned nil result")
					return
				}

				if len(result.Lookups) == 0 {
					t.Errorf("Lookup() returned empty lookups")
					return
				}

				if records, ok := result.Lookups[tt.recordType]; !ok {
					t.Errorf("Lookup() did not return %s records", tt.recordType)
				} else if len(records) == 0 {
					t.Errorf("Lookup() returned empty %s records", tt.recordType)
				}
			}
		})
	}
}

// TestIntegrationLookupAll tests the LookupAll function with real DNS queries.
func TestIntegrationLookupAll(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{
			name:    "Valid domain",
			domain:  "example.com",
			wantErr: false,
		},
		{
			name:    "Invalid domain",
			domain:  "thisdoesnotexist.example",
			wantErr: false, // LookupAll doesn't return an error for invalid domains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Call the function being tested
			result, err := LookupAll(ctx, tt.domain)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check the results
			if !tt.wantErr {
				if result == nil {
					t.Errorf("LookupAll() returned nil result")
					return
				}

				// For valid domains, we expect to get some records
				if tt.domain == "example.com" {
					if len(result.Lookups) == 0 {
						t.Errorf("LookupAll() returned empty lookups for valid domain")
						return
					}

					// Check that we got at least A and NS records
					if _, ok := result.Lookups["A"]; !ok {
						t.Errorf("LookupAll() did not return A records for valid domain")
					}

					if _, ok := result.Lookups["NS"]; !ok {
						t.Errorf("LookupAll() did not return NS records for valid domain")
					}
				}
				// For invalid domains, we don't expect any records, but the result should not be nil
			}
		})
	}
}

// TestIntegrationAdvancedLookup tests the AdvancedLookup function with real DNS queries.
func TestIntegrationAdvancedLookup(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		domain     string
		recordType string
		server     string
		wantErr    bool
	}{
		{
			name:       "Valid A record lookup with Google DNS",
			domain:     "example.com",
			recordType: "A",
			server:     "8.8.8.8:53",
			wantErr:    false,
		},
		{
			name:       "Valid AAAA record lookup with Google DNS",
			domain:     "example.com",
			recordType: "AAAA",
			server:     "8.8.8.8:53",
			wantErr:    false,
		},
		{
			name:       "Valid MX record lookup with Google DNS",
			domain:     "google.com",
			recordType: "MX",
			server:     "8.8.8.8:53",
			wantErr:    false,
		},
		{
			name:       "Invalid domain with Google DNS",
			domain:     "thisdoesnotexist.example",
			recordType: "A",
			server:     "8.8.8.8:53",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the function being tested
			result, err := AdvancedLookup(ctx, tt.domain, tt.recordType, tt.server)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("AdvancedLookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check that we got some results
			if !tt.wantErr {
				if result == nil {
					t.Errorf("AdvancedLookup() returned nil result")
					return
				}

				if len(result.Lookups) == 0 {
					t.Errorf("AdvancedLookup() returned empty lookups")
					return
				}

				if records, ok := result.Lookups[tt.recordType]; !ok {
					t.Errorf("AdvancedLookup() did not return %s records", tt.recordType)
				} else if len(records) == 0 {
					t.Errorf("AdvancedLookup() returned empty %s records", tt.recordType)
				}
			}
		})
	}
}

// TestIntegrationAdvancedLookupAll tests the AdvancedLookupAll function with real DNS queries.
func TestIntegrationAdvancedLookupAll(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name    string
		domain  string
		server  string
		wantErr bool
	}{
		{
			name:    "Valid domain with Google DNS",
			domain:  "example.com",
			server:  "8.8.8.8:53",
			wantErr: false,
		},
		{
			name:    "Invalid domain with Google DNS",
			domain:  "thisdoesnotexist.example",
			server:  "8.8.8.8:53",
			wantErr: false, // AdvancedLookupAll doesn't return an error for invalid domains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Call the function being tested
			result, err := AdvancedLookupAll(ctx, tt.domain, tt.server)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("AdvancedLookupAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check the results
			if !tt.wantErr {
				if result == nil {
					t.Errorf("AdvancedLookupAll() returned nil result")
					return
				}

				// For valid domains, we expect to get some records
				if tt.domain == "example.com" {
					if len(result.Lookups) == 0 {
						t.Errorf("AdvancedLookupAll() returned empty lookups for valid domain")
						return
					}

					// Check that we got at least A and NS records
					if _, ok := result.Lookups["A"]; !ok {
						t.Errorf("AdvancedLookupAll() did not return A records for valid domain")
					}

					if _, ok := result.Lookups["NS"]; !ok {
						t.Errorf("AdvancedLookupAll() did not return NS records for valid domain")
					}
				}
				// For invalid domains, we don't expect any records, but the result should not be nil
			}
		})
	}
}

// TestIntegrationLookupWithRetry tests the LookupWithRetry function with real DNS queries.
func TestIntegrationLookupWithRetry(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		domain     string
		recordType string
		maxRetries int
		timeout    time.Duration
		wantErr    bool
	}{
		{
			name:       "Valid A record lookup with retry",
			domain:     "example.com",
			recordType: "A",
			maxRetries: 2,
			timeout:    5 * time.Second,
			wantErr:    false,
		},
		{
			name:       "Invalid domain with retry",
			domain:     "thisdoesnotexist.example",
			recordType: "A",
			maxRetries: 2,
			timeout:    5 * time.Second,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context
			ctx := context.Background()

			// Call the function being tested
			result, err := LookupWithRetry(ctx, tt.domain, tt.recordType, tt.maxRetries, tt.timeout)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupWithRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check that we got some results
			if !tt.wantErr {
				if result == nil {
					t.Errorf("LookupWithRetry() returned nil result")
					return
				}

				if len(result.Lookups) == 0 {
					t.Errorf("LookupWithRetry() returned empty lookups")
					return
				}

				if records, ok := result.Lookups[tt.recordType]; !ok {
					t.Errorf("LookupWithRetry() did not return %s records", tt.recordType)
				} else if len(records) == 0 {
					t.Errorf("LookupWithRetry() returned empty %s records", tt.recordType)
				}
			}
		})
	}
}

// TestIntegrationAdvancedLookupWithRetry tests the AdvancedLookupWithRetry function with real DNS queries.
func TestIntegrationAdvancedLookupWithRetry(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		domain     string
		recordType string
		server     string
		maxRetries int
		timeout    time.Duration
		wantErr    bool
	}{
		{
			name:       "Valid A record lookup with retry and Google DNS",
			domain:     "example.com",
			recordType: "A",
			server:     "8.8.8.8:53",
			maxRetries: 2,
			timeout:    5 * time.Second,
			wantErr:    false,
		},
		{
			name:       "Invalid domain with retry and Google DNS",
			domain:     "thisdoesnotexist.example",
			recordType: "A",
			server:     "8.8.8.8:53",
			maxRetries: 2,
			timeout:    5 * time.Second,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context
			ctx := context.Background()

			// Call the function being tested
			result, err := AdvancedLookupWithRetry(ctx, tt.domain, tt.recordType, tt.server, tt.maxRetries, tt.timeout)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("AdvancedLookupWithRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check that we got some results
			if !tt.wantErr {
				if result == nil {
					t.Errorf("AdvancedLookupWithRetry() returned nil result")
					return
				}

				if len(result.Lookups) == 0 {
					t.Errorf("AdvancedLookupWithRetry() returned empty lookups")
					return
				}

				if records, ok := result.Lookups[tt.recordType]; !ok {
					t.Errorf("AdvancedLookupWithRetry() did not return %s records", tt.recordType)
				} else if len(records) == 0 {
					t.Errorf("AdvancedLookupWithRetry() returned empty %s records", tt.recordType)
				}
			}
		})
	}
}
