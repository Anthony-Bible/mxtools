// Package networktools provides auxiliary network diagnostic tools.
package networktools

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestFormatTracerouteResult tests the FormatTracerouteResult function
func TestFormatTracerouteResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *TracerouteResult
		contains []string
		notContains []string
	}{
		{
			name: "Successful traceroute with privileges",
			result: &TracerouteResult{
				Target: "example.com",
				Hops: []TracerouteHop{
					{
						TTL:     1,
						Address: "192.168.1.1",
						RTT:     10 * time.Millisecond,
					},
					{
						TTL:      2,
						Address:  "203.0.113.1",
						Hostname: "router.example.net",
						RTT:      20 * time.Millisecond,
					},
					{
						TTL:     3,
						Address: "93.184.216.34", // example.com IP
						RTT:     30 * time.Millisecond,
					},
				},
				IsPrivileged: true,
			},
			contains: []string{
				"Traceroute to example.com",
				"1  192.168.1.1  10ms",
				"2  router.example.net (203.0.113.1)  20ms",
				"3  93.184.216.34  30ms",
			},
			notContains: []string{
				"unprivileged mode",
				"failed",
			},
		},
		{
			name: "Successful traceroute without privileges",
			result: &TracerouteResult{
				Target: "example.com",
				Hops: []TracerouteHop{
					{
						TTL:     1,
						Address: "192.168.1.1",
						RTT:     10 * time.Millisecond,
					},
					{
						TTL:     2,
						Address: "203.0.113.1",
						RTT:     20 * time.Millisecond,
					},
				},
				IsPrivileged: false,
			},
			contains: []string{
				"Traceroute to example.com",
				"1  192.168.1.1  10ms",
				"2  203.0.113.1  20ms",
				"unprivileged mode",
			},
			notContains: []string{
				"failed",
			},
		},
		{
			name: "Traceroute with errors",
			result: &TracerouteResult{
				Target: "example.com",
				Hops: []TracerouteHop{
					{
						TTL:     1,
						Address: "192.168.1.1",
						RTT:     10 * time.Millisecond,
					},
					{
						TTL:   2,
						Error: "Request timed out",
					},
					{
						TTL:     3,
						Address: "93.184.216.34", // example.com IP
						RTT:     30 * time.Millisecond,
					},
				},
				IsPrivileged: true,
			},
			contains: []string{
				"Traceroute to example.com",
				"1  192.168.1.1  10ms",
				"2  Request timed out",
				"3  93.184.216.34  30ms",
			},
			notContains: []string{
				"unprivileged mode",
				"failed",
			},
		},
		{
			name: "Failed traceroute",
			result: &TracerouteResult{
				Target: "example.com",
				Error:  "Failed to resolve target",
			},
			contains: []string{
				"Traceroute to example.com failed: Failed to resolve target",
			},
			notContains: []string{
				"unprivileged mode",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTracerouteResult(tt.result)
			
			// Check that the output contains all expected substrings
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("FormatTracerouteResult() = %v, should contain %v", result, s)
				}
			}
			
			// Check that the output does not contain any unwanted substrings
			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("FormatTracerouteResult() = %v, should not contain %v", result, s)
				}
			}
		})
	}
}

// TestTracerouteWithPrivilegeCheck tests the TracerouteWithPrivilegeCheck function
func TestTracerouteWithPrivilegeCheck(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping test that requires network access in short mode")
	}

	// This test verifies that TracerouteWithPrivilegeCheck adds the privilege message
	// when needed. We'll use a non-existent domain to ensure the traceroute fails.
	result, err := TracerouteWithPrivilegeCheck(context.Background(), "nonexistent.domain.that.should.fail", 3, 1*time.Second)
	
	// We expect an error
	if err == nil {
		t.Errorf("TracerouteWithPrivilegeCheck() should return an error for non-existent domain")
	}
	
	// Check that the result contains an error
	if result == nil || result.Error == "" {
		t.Errorf("TracerouteWithPrivilegeCheck() should return a result with an error message")
	} else {
		// If the traceroute was run without privileges, check that the error message mentions privileges
		if !result.IsPrivileged {
			if !strings.Contains(result.Error, "elevated privileges") {
				t.Errorf("TracerouteWithPrivilegeCheck() error = %v, should mention privileges when not privileged", result.Error)
			}
		}
	}
}