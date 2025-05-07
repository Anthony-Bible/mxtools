// Package input contains the input ports (interfaces) for the application
package input

import (
	"context"
	"mxclone/domain/dnsbl"
	"time"
)

// DNSBLPort defines the input interface for DNSBL operations
type DNSBLPort interface {
	// CheckSingleBlacklist checks if an IP is listed on a specific DNSBL
	CheckSingleBlacklist(ctx context.Context, ip string, zone string, timeout time.Duration) (*dnsbl.BlacklistResult, error)

	// CheckMultipleBlacklists checks if an IP is listed on multiple DNSBLs concurrently
	CheckMultipleBlacklists(ctx context.Context, ip string, zones []string, timeout time.Duration) (*dnsbl.BlacklistResult, error)

	// GetBlacklistSummary returns a human-readable summary of blacklist check results
	GetBlacklistSummary(result *dnsbl.BlacklistResult) string

	// CheckDNSBLHealth checks if a DNSBL is operational
	CheckDNSBLHealth(ctx context.Context, zone string, timeout time.Duration) (bool, error)

	// CheckMultipleDNSBLHealth checks the health of multiple DNSBLs concurrently
	CheckMultipleDNSBLHealth(ctx context.Context, zones []string, timeout time.Duration) map[string]bool
}
