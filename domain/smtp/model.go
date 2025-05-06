// Package smtp contains the core domain logic for SMTP operations
package smtp

import (
	"fmt"
	"time"
)

// SMTPResult represents the result of an SMTP check operation
type SMTPResult struct {
	// The domain that was checked
	Domain string
	// The MX records for the domain
	MXRecords []string
	// Connection results for each MX server
	ConnectionResults map[string]*ConnectionResult
	// Response code and banner from SMTP server
	Banner string
	// Error message if any
	Error string
}

// ConnectionResult represents the result of a connection attempt to an SMTP server
type ConnectionResult struct {
	// Server hostname or IP
	Server string
	// Whether the connection was successful
	Connected bool
	// Connection latency
	Latency time.Duration
	// TLS/SSL support
	SupportsStartTLS bool
	// Authentication methods supported
	AuthMethods []string
	// Banner message
	Banner string
	// Error message if any
	Error string
}

// Service defines the core SMTP business logic operations
type Service struct {
	// You can inject dependencies here if needed
}

// NewService creates a new SMTP service
func NewService() *Service {
	return &Service{}
}

// ProcessSMTPResult processes the result of an SMTP check
func (s *Service) ProcessSMTPResult(domain string, mxRecords []string, connectionResults map[string]*ConnectionResult, banner string, err error) *SMTPResult {
	result := &SMTPResult{
		Domain:            domain,
		MXRecords:         mxRecords,
		ConnectionResults: connectionResults,
		Banner:            banner,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// CreateConnectionResult creates a connection result for a specific server
func (s *Service) CreateConnectionResult(server string, connected bool, latency time.Duration, supportsStartTLS bool, authMethods []string, banner string, err error) *ConnectionResult {
	result := &ConnectionResult{
		Server:           server,
		Connected:        connected,
		Latency:          latency,
		SupportsStartTLS: supportsStartTLS,
		AuthMethods:      authMethods,
		Banner:           banner,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// FormatSMTPSummary returns a human-readable summary of SMTP check results
func (s *Service) FormatSMTPSummary(result *SMTPResult) string {
	if result == nil {
		return "No SMTP check results available"
	}

	summary := fmt.Sprintf("SMTP check results for domain %s:\n", result.Domain)

	if len(result.MXRecords) == 0 {
		summary += "No MX records found\n"
	} else {
		summary += fmt.Sprintf("\nFound %d MX records:\n", len(result.MXRecords))
		for _, mx := range result.MXRecords {
			summary += fmt.Sprintf("- %s\n", mx)
		}
	}

	if len(result.ConnectionResults) > 0 {
		summary += "\nConnection results:\n"
		for server, connResult := range result.ConnectionResults {
			summary += fmt.Sprintf("- %s: ", server)
			if connResult.Connected {
				summary += fmt.Sprintf("Connected (latency: %v)\n", connResult.Latency)
				if connResult.SupportsStartTLS {
					summary += "  Supports STARTTLS: Yes\n"
				} else {
					summary += "  Supports STARTTLS: No\n"
				}

				if len(connResult.AuthMethods) > 0 {
					summary += "  Auth methods: " + fmt.Sprintf("%v", connResult.AuthMethods) + "\n"
				} else {
					summary += "  Auth methods: None\n"
				}

				if connResult.Banner != "" {
					summary += fmt.Sprintf("  Banner: %s\n", connResult.Banner)
				}
			} else {
				summary += fmt.Sprintf("Failed to connect: %s\n", connResult.Error)
			}
		}
	}

	if result.Error != "" {
		summary += fmt.Sprintf("\nErrors: %s\n", result.Error)
	}

	return summary
}
