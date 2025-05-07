// Package output contains the output ports (interfaces) for the application
package output

import (
	"context"
	"time"
)

// SMTPRepository defines the output interface for SMTP operations
type SMTPRepository interface {
	// GetMXRecords retrieves the MX records for a domain
	GetMXRecords(ctx context.Context, domain string) ([]string, error)

	// ConnectToSMTPServer connects to an SMTP server and tests its capabilities
	ConnectToSMTPServer(ctx context.Context, server string, port int, timeout time.Duration) (bool, time.Duration, bool, []string, string, error)
}
