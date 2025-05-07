// Package input contains the input ports (interfaces) for the application
package input

import (
	"context"
	"mxclone/domain/smtp"
	"time"
)

// SMTPPort defines the input interface for SMTP operations
type SMTPPort interface {
	// CheckSMTP performs a comprehensive SMTP check for a domain
	CheckSMTP(ctx context.Context, domain string, timeout time.Duration) (*smtp.SMTPResult, error)

	// TestSMTPConnection tests connection to a specific SMTP server
	TestSMTPConnection(ctx context.Context, server string, port int, timeout time.Duration) (*smtp.ConnectionResult, error)

	// GetSMTPSummary returns a human-readable summary of SMTP check results
	GetSMTPSummary(result *smtp.SMTPResult) string
}
