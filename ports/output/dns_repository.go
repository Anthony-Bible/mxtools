// Package output contains the output ports (interfaces) for the application
package output

import (
	"context"
	"mxclone/domain/dns"
)

// DNSRepository defines the output interface for DNS operations
type DNSRepository interface {
	// LookupRecords performs the actual DNS lookup operation
	LookupRecords(ctx context.Context, domain string, recordType dns.RecordType) ([]string, error)
}
