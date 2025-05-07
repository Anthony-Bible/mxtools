// Package input contains the input ports (interfaces) for the application
package input

import (
	"context"
	"mxclone/domain/dns"
)

// DNSPort defines the input interface for DNS operations
type DNSPort interface {
	// Lookup performs a DNS lookup for a specific record type
	Lookup(ctx context.Context, domain string, recordType dns.RecordType) (*dns.DNSResult, error)

	// LookupAll performs DNS lookups for all supported record types
	LookupAll(ctx context.Context, domain string) (*dns.DNSResult, error)
}
