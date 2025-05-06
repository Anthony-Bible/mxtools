// Package primary contains the primary adapters (implementing input ports)
package primary

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mxclone/domain/smtp"
	"mxclone/ports/output"
)

// SMTPAdapter implements the SMTP input port
type SMTPAdapter struct {
	smtpService *smtp.Service
	repository  output.SMTPRepository
}

// NewSMTPAdapter creates a new SMTP adapter
func NewSMTPAdapter(repository output.SMTPRepository) *SMTPAdapter {
	return &SMTPAdapter{
		smtpService: smtp.NewService(),
		repository:  repository,
	}
}

// CheckSMTP performs a comprehensive SMTP check for a domain
func (a *SMTPAdapter) CheckSMTP(ctx context.Context, domain string, timeout time.Duration) (*smtp.SMTPResult, error) {
	// Get MX records for the domain
	mxRecords, err := a.repository.GetMXRecords(ctx, domain)
	if err != nil {
		return a.smtpService.ProcessSMTPResult(domain, nil, nil, "", fmt.Errorf("failed to retrieve MX records: %w", err)), err
	}

	// If no MX records found, return early
	if len(mxRecords) == 0 {
		return a.smtpService.ProcessSMTPResult(domain, mxRecords, nil, "", fmt.Errorf("no MX records found for domain")), nil
	}

	// Test connection to each MX server
	connectionResults := make(map[string]*smtp.ConnectionResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, server := range mxRecords {
		wg.Add(1)
		go func(srv string) {
			defer wg.Done()

			// Test connection to this server
			connResult, _ := a.TestSMTPConnection(ctx, srv, 25, timeout)

			// Store connection result
			mu.Lock()
			connectionResults[srv] = connResult
			mu.Unlock()
		}(server)
	}

	// Wait for all connection tests to complete
	wg.Wait()

	// Extract banner from the first successful connection
	var banner string
	for _, result := range connectionResults {
		if result.Connected && result.Banner != "" {
			banner = result.Banner
			break
		}
	}

	// Process and return the result
	return a.smtpService.ProcessSMTPResult(domain, mxRecords, connectionResults, banner, nil), nil
}

// TestSMTPConnection tests connection to a specific SMTP server
func (a *SMTPAdapter) TestSMTPConnection(ctx context.Context, server string, port int, timeout time.Duration) (*smtp.ConnectionResult, error) {
	// Connect to SMTP server and test capabilities
	connected, latency, supportsStartTLS, authMethods, banner, err := a.repository.ConnectToSMTPServer(ctx, server, port, timeout)

	// Process and return the connection result
	connResult := a.smtpService.CreateConnectionResult(server, connected, latency, supportsStartTLS, authMethods, banner, err)

	return connResult, nil
}

// GetSMTPSummary returns a human-readable summary of SMTP check results
func (a *SMTPAdapter) GetSMTPSummary(result *smtp.SMTPResult) string {
	return a.smtpService.FormatSMTPSummary(result)
}
