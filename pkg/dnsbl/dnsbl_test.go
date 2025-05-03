// Package dnsbl provides functionality for checking IP addresses against DNS-based blacklists.
package dnsbl

import (
	"testing"
)

func TestReverseIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid IPv4",
			ip:       "192.168.1.1",
			expected: "1.1.168.192",
			wantErr:  false,
		},
		{
			name:     "Valid IPv4 with leading zeros",
			ip:       "8.8.8.8",
			expected: "8.8.8.8",
			wantErr:  false,
		},
		{
			name:     "Invalid IP",
			ip:       "not an ip",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "IPv6 address",
			ip:       "2001:db8::1",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReverseIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReverseIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ReverseIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildDNSBLQuery(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		zone     string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid IP and zone",
			ip:       "192.168.1.1",
			zone:     "zen.spamhaus.org",
			expected: "1.1.168.192.zen.spamhaus.org",
			wantErr:  false,
		},
		{
			name:     "Invalid IP",
			ip:       "not an ip",
			zone:     "zen.spamhaus.org",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildDNSBLQuery(tt.ip, tt.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDNSBLQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("BuildDNSBLQuery() = %v, want %v", got, tt.expected)
			}
		})
	}
}