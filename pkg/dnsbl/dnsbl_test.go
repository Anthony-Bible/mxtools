// Package dnsbl provides functionality for checking IP addresses against DNS-based blacklists.
package dnsbl

import (
	"strings"
	"testing"

	"mxclone/pkg/types"
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

func TestAggregateBlacklistResults(t *testing.T) {
	tests := []struct {
		name    string
		results []*types.BlacklistResult
		want    *types.BlacklistResult
	}{
		{
			name: "Empty results",
			results: []*types.BlacklistResult{},
			want: nil,
		},
		{
			name: "Single result",
			results: []*types.BlacklistResult{
				{
					CheckedIP: "192.168.1.1",
					ListedOn: map[string]string{
						"zen.spamhaus.org": "Listed for spam",
					},
					CheckError: "",
				},
			},
			want: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn: map[string]string{
					"zen.spamhaus.org": "Listed for spam",
				},
				CheckError: "",
			},
		},
		{
			name: "Multiple results for same IP",
			results: []*types.BlacklistResult{
				{
					CheckedIP: "192.168.1.1",
					ListedOn: map[string]string{
						"zen.spamhaus.org": "Listed for spam",
					},
					CheckError: "",
				},
				{
					CheckedIP: "192.168.1.1",
					ListedOn: map[string]string{
						"bl.spamcop.net": "Listed for abuse",
					},
					CheckError: "",
				},
			},
			want: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn: map[string]string{
					"zen.spamhaus.org": "Listed for spam",
					"bl.spamcop.net":   "Listed for abuse",
				},
				CheckError: "",
			},
		},
		{
			name: "Multiple results for different IPs",
			results: []*types.BlacklistResult{
				{
					CheckedIP: "192.168.1.1",
					ListedOn: map[string]string{
						"zen.spamhaus.org": "Listed for spam",
					},
					CheckError: "",
				},
				{
					CheckedIP: "192.168.1.2",
					ListedOn: map[string]string{
						"bl.spamcop.net": "Listed for abuse",
					},
					CheckError: "",
				},
			},
			want: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn: map[string]string{
					"zen.spamhaus.org": "Listed for spam",
				},
				CheckError: "",
			},
		},
		{
			name: "Results with errors",
			results: []*types.BlacklistResult{
				{
					CheckedIP:  "192.168.1.1",
					ListedOn:   map[string]string{},
					CheckError: "Error checking zen.spamhaus.org",
				},
				{
					CheckedIP:  "192.168.1.1",
					ListedOn:   map[string]string{},
					CheckError: "Error checking bl.spamcop.net",
				},
			},
			want: &types.BlacklistResult{
				CheckedIP:  "192.168.1.1",
				ListedOn:   map[string]string{},
				CheckError: "Error checking zen.spamhaus.org; Error checking bl.spamcop.net",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateBlacklistResults(tt.results)

			// Check if both are nil or both are not nil
			if (got == nil) != (tt.want == nil) {
				t.Errorf("AggregateBlacklistResults() = %v, want %v", got, tt.want)
				return
			}

			// If both are nil, we're done
			if got == nil && tt.want == nil {
				return
			}

			// Check CheckedIP
			if got.CheckedIP != tt.want.CheckedIP {
				t.Errorf("AggregateBlacklistResults().CheckedIP = %v, want %v", got.CheckedIP, tt.want.CheckedIP)
			}

			// Check ListedOn
			if len(got.ListedOn) != len(tt.want.ListedOn) {
				t.Errorf("AggregateBlacklistResults().ListedOn has %d entries, want %d", len(got.ListedOn), len(tt.want.ListedOn))
			}
			for zone, reason := range tt.want.ListedOn {
				if gotReason, ok := got.ListedOn[zone]; !ok {
					t.Errorf("AggregateBlacklistResults().ListedOn missing zone %s", zone)
				} else if gotReason != reason {
					t.Errorf("AggregateBlacklistResults().ListedOn[%s] = %s, want %s", zone, gotReason, reason)
				}
			}

			// Check CheckError
			if got.CheckError != tt.want.CheckError {
				t.Errorf("AggregateBlacklistResults().CheckError = %v, want %v", got.CheckError, tt.want.CheckError)
			}
		})
	}
}

func TestGetBlacklistSummary(t *testing.T) {
	tests := []struct {
		name          string
		result        *types.BlacklistResult
		wantContains  []string
		wantNotContains []string
	}{
		{
			name:          "Nil result",
			result:        nil,
			wantContains:  []string{"No blacklist check results available"},
			wantNotContains: []string{},
		},
		{
			name: "Not listed on any blacklists",
			result: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn:  map[string]string{},
			},
			wantContains:  []string{"IP 192.168.1.1 is not listed on any blacklists"},
			wantNotContains: []string{"Errors:"},
		},
		{
			name: "Listed on one blacklist with reason",
			result: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn: map[string]string{
					"zen.spamhaus.org": "Listed for spam",
				},
			},
			wantContains:  []string{
				"IP 192.168.1.1 is listed on 1 blacklists:",
				"zen.spamhaus.org: Listed for spam",
			},
			wantNotContains: []string{"Errors:"},
		},
		{
			name: "Listed on one blacklist without reason",
			result: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn: map[string]string{
					"zen.spamhaus.org": "",
				},
			},
			wantContains:  []string{
				"IP 192.168.1.1 is listed on 1 blacklists:",
				"zen.spamhaus.org",
			},
			wantNotContains: []string{"Errors:"},
		},
		{
			name: "Listed on multiple blacklists",
			result: &types.BlacklistResult{
				CheckedIP: "192.168.1.1",
				ListedOn: map[string]string{
					"zen.spamhaus.org": "Listed for spam",
					"bl.spamcop.net":   "Listed for abuse",
				},
			},
			wantContains:  []string{
				"IP 192.168.1.1 is listed on 2 blacklists:",
				"zen.spamhaus.org: Listed for spam",
				"bl.spamcop.net: Listed for abuse",
			},
			wantNotContains: []string{"Errors:"},
		},
		{
			name: "With error",
			result: &types.BlacklistResult{
				CheckedIP:  "192.168.1.1",
				ListedOn:   map[string]string{},
				CheckError: "Error checking blacklists",
			},
			wantContains:  []string{
				"IP 192.168.1.1 is not listed on any blacklists",
				"Errors: Error checking blacklists",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBlacklistSummary(tt.result)

			// Check that the output contains all the expected substrings
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GetBlacklistSummary() = %v, should contain %v", got, want)
				}
			}

			// Check that the output does not contain any of the unwanted substrings
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("GetBlacklistSummary() = %v, should not contain %v", got, notWant)
				}
			}
		})
	}
}
