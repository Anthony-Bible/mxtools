// Package networktools provides auxiliary network diagnostic tools.
package networktools

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

// mockICMPConn is a mock implementation of the icmp.PacketConn interface
type mockICMPConn struct {
	writeTo   func(b []byte, addr net.Addr) (int, error)
	readFrom  func(b []byte) (int, net.Addr, error)
	close     func() error
	deadline  func(t time.Time) error
	localAddr func() net.Addr
}

func (m *mockICMPConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if m.writeTo != nil {
		return m.writeTo(b, addr)
	}
	return 0, errors.New("WriteTo not implemented")
}

func (m *mockICMPConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if m.readFrom != nil {
		return m.readFrom(b)
	}
	return 0, nil, errors.New("ReadFrom not implemented")
}

func (m *mockICMPConn) Close() error {
	if m.close != nil {
		return m.close()
	}
	return nil
}

func (m *mockICMPConn) SetDeadline(t time.Time) error {
	if m.deadline != nil {
		return m.deadline(t)
	}
	return nil
}

func (m *mockICMPConn) LocalAddr() net.Addr {
	if m.localAddr != nil {
		return m.localAddr()
	}
	return nil
}

// TestFormatPingResult tests the FormatPingResult function
func TestFormatPingResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *PingResult
		expected string
		contains []string
	}{
		{
			name: "Successful ping",
			result: &PingResult{
				Target:       "example.com",
				Sent:         4,
				Received:     4,
				PacketLoss:   0,
				MinRTT:       10 * time.Millisecond,
				MaxRTT:       20 * time.Millisecond,
				AvgRTT:       15 * time.Millisecond,
				IsPrivileged: true,
			},
			contains: []string{
				"PING example.com",
				"Sent: 4, Received: 4, Packet Loss: 0.0%",
				"Round-trip min/avg/max: 10ms/15ms/20ms",
			},
		},
		{
			name: "Partial packet loss",
			result: &PingResult{
				Target:       "example.com",
				Sent:         4,
				Received:     2,
				PacketLoss:   50,
				MinRTT:       10 * time.Millisecond,
				MaxRTT:       20 * time.Millisecond,
				AvgRTT:       15 * time.Millisecond,
				IsPrivileged: true,
			},
			contains: []string{
				"PING example.com",
				"Sent: 4, Received: 2, Packet Loss: 50.0%",
				"Round-trip min/avg/max: 10ms/15ms/20ms",
			},
		},
		{
			name: "Complete packet loss",
			result: &PingResult{
				Target:       "example.com",
				Sent:         4,
				Received:     0,
				PacketLoss:   100,
				IsPrivileged: true,
			},
			contains: []string{
				"PING example.com",
				"Sent: 4, Received: 0, Packet Loss: 100.0%",
			},
		},
		{
			name: "Error result",
			result: &PingResult{
				Target: "example.com",
				Error:  "Failed to resolve target",
			},
			contains: []string{
				"Ping to example.com failed: Failed to resolve target",
			},
		},
		{
			name: "Unprivileged mode",
			result: &PingResult{
				Target:       "example.com",
				Sent:         4,
				Received:     4,
				PacketLoss:   0,
				MinRTT:       10 * time.Millisecond,
				MaxRTT:       20 * time.Millisecond,
				AvgRTT:       15 * time.Millisecond,
				IsPrivileged: false,
			},
			contains: []string{
				"PING example.com",
				"Sent: 4, Received: 4, Packet Loss: 0.0%",
				"Round-trip min/avg/max: 10ms/15ms/20ms",
				"Note: Running in unprivileged mode",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPingResult(tt.result)

			for _, s := range tt.contains {
				if !contains(result, s) {
					t.Errorf("FormatPingResult() = %v, should contain %v", result, s)
				}
			}
		})
	}
}

// TestPingWithPrivilegeCheck tests the PingWithPrivilegeCheck function
func TestPingWithPrivilegeCheck(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping test that requires network access in short mode")
	}

	// This test verifies that PingWithPrivilegeCheck adds the privilege message
	// when needed. We'll use a non-existent domain to ensure the ping fails.
	result, err := PingWithPrivilegeCheck(context.Background(), "nonexistent.domain.that.should.fail", 1, 1*time.Second)

	// We expect an error
	if err == nil {
		t.Errorf("PingWithPrivilegeCheck() should return an error for non-existent domain")
	}

	// Check that the result contains an error
	if result == nil || result.Error == "" {
		t.Errorf("PingWithPrivilegeCheck() should return a result with an error message")
	} else {
		// If the ping was run without privileges, check that the error message mentions privileges
		if !result.IsPrivileged {
			if !contains(result.Error, "elevated privileges") {
				t.Errorf("PingWithPrivilegeCheck() error = %v, should mention privileges when not privileged", result.Error)
			}
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s[0:len(substr)] == substr
}
