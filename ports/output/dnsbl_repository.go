// Package output contains the output ports (interfaces) for the application
package output

import (
	"context"
	"time"
)

// DNSBLRepository defines the output interface for DNSBL operations
type DNSBLRepository interface {
	// LookupDNSBLRecord checks if a domain (reversed IP + zone) has A records
	LookupDNSBLRecord(ctx context.Context, query string, timeout time.Duration) (bool, error)

	// GetDNSBLExplanation tries to retrieve TXT record explanation for a DNSBL listing
	GetDNSBLExplanation(ctx context.Context, query string, timeout time.Duration) (string, error)

	// CheckDNSBLAvailability checks if a DNSBL service is available
	CheckDNSBLAvailability(ctx context.Context, zone string, timeout time.Duration) (bool, error)
}
