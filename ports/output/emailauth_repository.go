// Package output contains the output ports (interfaces) for the application
package output

import (
	"context"
	"time"
)

// EmailAuthRepository defines the output interface for email authentication operations
type EmailAuthRepository interface {
	// GetSPFRecord retrieves the SPF record for a domain
	GetSPFRecord(ctx context.Context, domain string, timeout time.Duration) (string, bool, error)

	// ValidateSPFRecord validates the SPF record and extracts mechanisms
	ValidateSPFRecord(record string) (bool, []string, error)

	// GetDKIMRecord retrieves the DKIM record for a domain and selector
	GetDKIMRecord(ctx context.Context, domain string, selector string, timeout time.Duration) (string, bool, error)

	// ValidateDKIMRecord validates a DKIM record
	ValidateDKIMRecord(record string) (bool, error)

	// GetDMARCRecord retrieves the DMARC record for a domain
	GetDMARCRecord(ctx context.Context, domain string, timeout time.Duration) (string, bool, error)

	// ParseDMARCRecord parses a DMARC record to extract policy information
	ParseDMARCRecord(record string) (bool, string, string, int, error)
}
