// Package input contains the input ports (interfaces) for the application
package input

import (
	"context"
	"mxclone/domain/emailauth"
	"time"
)

// EmailAuthPort defines the input interface for email authentication operations
type EmailAuthPort interface {
	// CheckSPF checks SPF (Sender Policy Framework) records for a domain
	CheckSPF(ctx context.Context, domain string, timeout time.Duration) (*emailauth.SPFResult, error)

	// CheckDKIM checks DKIM (DomainKeys Identified Mail) records for a domain
	CheckDKIM(ctx context.Context, domain string, selectors []string, timeout time.Duration) (*emailauth.DKIMResult, error)

	// CheckDMARC checks DMARC (Domain-based Message Authentication) records for a domain
	CheckDMARC(ctx context.Context, domain string, timeout time.Duration) (*emailauth.DMARCResult, error)

	// CheckAll performs SPF, DKIM, and DMARC checks for a domain
	CheckAll(ctx context.Context, domain string, dkimSelectors []string, timeout time.Duration) (*emailauth.AuthResult, error)

	// GetAuthSummary returns a human-readable summary of email authentication checks
	GetAuthSummary(result *emailauth.AuthResult) string
}
