// Package emailauth contains the core domain logic for email authentication operations
package emailauth

import (
	"fmt"
)

// AuthResult represents the result of email authentication checks (SPF, DKIM, DMARC)
type AuthResult struct {
	// Domain that was checked
	Domain string
	// SPF (Sender Policy Framework) check results
	SPF *SPFResult
	// DKIM (DomainKeys Identified Mail) check results
	DKIM *DKIMResult
	// DMARC (Domain-based Message Authentication, Reporting & Conformance) check results
	DMARC *DMARCResult
	// Error message if any
	Error string
}

// SPFResult represents the result of an SPF check
type SPFResult struct {
	// Whether the domain has an SPF record
	HasRecord bool
	// The SPF record value
	Record string
	// Whether the SPF record is valid
	IsValid bool
	// SPF mechanisms found in the record
	Mechanisms []string
	// Error message if any
	Error string
}

// DKIMResult represents the result of a DKIM check
type DKIMResult struct {
	// Whether the domain has DKIM records
	HasRecords bool
	// Map of selector to DKIM record
	Records map[string]string
	// Whether the DKIM records are valid
	IsValid bool
	// Error message if any
	Error string
}

// DMARCResult represents the result of a DMARC check
type DMARCResult struct {
	// Whether the domain has a DMARC record
	HasRecord bool
	// The DMARC record value
	Record string
	// Whether the DMARC record is valid
	IsValid bool
	// The DMARC policy (none, quarantine, reject)
	Policy string
	// The DMARC subdomain policy
	SubdomainPolicy string
	// Percentage of messages to which the policy applies
	Percentage int
	// Error message if any
	Error string
}

// Service defines the core EmailAuth business logic operations
type Service struct {
	// You can inject dependencies here if needed
}

// NewService creates a new EmailAuth service
func NewService() *Service {
	return &Service{}
}

// ProcessSPFResult processes SPF check results
func (s *Service) ProcessSPFResult(hasRecord bool, record string, isValid bool, mechanisms []string, err error) *SPFResult {
	result := &SPFResult{
		HasRecord:  hasRecord,
		Record:     record,
		IsValid:    isValid,
		Mechanisms: mechanisms,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// ProcessDKIMResult processes DKIM check results
func (s *Service) ProcessDKIMResult(hasRecords bool, records map[string]string, isValid bool, err error) *DKIMResult {
	result := &DKIMResult{
		HasRecords: hasRecords,
		Records:    records,
		IsValid:    isValid,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// ProcessDMARCResult processes DMARC check results
func (s *Service) ProcessDMARCResult(hasRecord bool, record string, isValid bool, policy string, subdomainPolicy string, percentage int, err error) *DMARCResult {
	result := &DMARCResult{
		HasRecord:       hasRecord,
		Record:          record,
		IsValid:         isValid,
		Policy:          policy,
		SubdomainPolicy: subdomainPolicy,
		Percentage:      percentage,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// ProcessAuthResult processes overall email authentication check results
func (s *Service) ProcessAuthResult(domain string, spf *SPFResult, dkim *DKIMResult, dmarc *DMARCResult, err error) *AuthResult {
	result := &AuthResult{
		Domain: domain,
		SPF:    spf,
		DKIM:   dkim,
		DMARC:  dmarc,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// FormatAuthSummary returns a human-readable summary of email authentication checks
func (s *Service) FormatAuthSummary(result *AuthResult) string {
	if result == nil {
		return "No email authentication results available"
	}

	summary := fmt.Sprintf("Email Authentication results for %s:\n", result.Domain)

	// SPF summary
	if result.SPF != nil {
		summary += "\nSPF:\n"
		if result.SPF.HasRecord {
			summary += fmt.Sprintf("  Record: %s\n", result.SPF.Record)
			if result.SPF.IsValid {
				summary += "  Status: Valid\n"
			} else {
				summary += "  Status: Invalid\n"
			}

			if len(result.SPF.Mechanisms) > 0 {
				summary += "  Mechanisms: "
				for i, mechanism := range result.SPF.Mechanisms {
					if i > 0 {
						summary += ", "
					}
					summary += mechanism
				}
				summary += "\n"
			}
		} else {
			summary += "  No SPF record found\n"
		}

		if result.SPF.Error != "" {
			summary += fmt.Sprintf("  Error: %s\n", result.SPF.Error)
		}
	}

	// DKIM summary
	if result.DKIM != nil {
		summary += "\nDKIM:\n"
		if result.DKIM.HasRecords {
			for selector, record := range result.DKIM.Records {
				summary += fmt.Sprintf("  Selector: %s\n", selector)
				summary += fmt.Sprintf("  Record: %s\n", record)
			}

			if result.DKIM.IsValid {
				summary += "  Status: Valid\n"
			} else {
				summary += "  Status: Invalid\n"
			}
		} else {
			summary += "  No DKIM records found\n"
		}

		if result.DKIM.Error != "" {
			summary += fmt.Sprintf("  Error: %s\n", result.DKIM.Error)
		}
	}

	// DMARC summary
	if result.DMARC != nil {
		summary += "\nDMARC:\n"
		if result.DMARC.HasRecord {
			summary += fmt.Sprintf("  Record: %s\n", result.DMARC.Record)
			if result.DMARC.IsValid {
				summary += "  Status: Valid\n"
			} else {
				summary += "  Status: Invalid\n"
			}

			summary += fmt.Sprintf("  Policy: %s\n", result.DMARC.Policy)
			if result.DMARC.SubdomainPolicy != "" {
				summary += fmt.Sprintf("  Subdomain Policy: %s\n", result.DMARC.SubdomainPolicy)
			}
			summary += fmt.Sprintf("  Percentage: %d%%\n", result.DMARC.Percentage)
		} else {
			summary += "  No DMARC record found\n"
		}

		if result.DMARC.Error != "" {
			summary += fmt.Sprintf("  Error: %s\n", result.DMARC.Error)
		}
	}

	if result.Error != "" {
		summary += fmt.Sprintf("\nOverall Errors: %s\n", result.Error)
	}

	return summary
}
