package models

import (
	"mxclone/domain/dns"
	"mxclone/domain/dnsbl"
	"mxclone/domain/smtp"
)

// APIError represents an API error response
type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// CheckRequest represents a generic diagnostic request
type CheckRequest struct {
	Target string `json:"target"`
	Option string `json:"option,omitempty"` // Optional parameter for specific checks
}

// DNSResponse wraps the domain DNS result for API responses
type DNSResponse struct {
	Records map[string][]string `json:"records"`
	Timing  string              `json:"timing,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// FromDNSResult converts a domain DNS result to an API response
func FromDNSResult(result *dns.DNSResult) *DNSResponse {
	if result == nil {
		return &DNSResponse{
			Error: "no result available",
		}
	}

	return &DNSResponse{
		Records: result.Lookups,
		Error:   result.Error,
	}
}

// BlacklistResponse wraps the domain blacklist result for API responses
type BlacklistResponse struct {
	IP       string            `json:"ip"`
	ListedOn map[string]string `json:"listedOn"`
	Error    string            `json:"error,omitempty"`
}

// FromBlacklistResult converts a domain blacklist result to an API response
func FromBlacklistResult(result *dnsbl.BlacklistResult) *BlacklistResponse {
	if result == nil {
		return &BlacklistResponse{
			Error: "no result available",
		}
	}

	return &BlacklistResponse{
		IP:       result.CheckedIP,
		ListedOn: result.ListedOn,
		Error:    result.CheckError,
	}
}

// SMTPResponse wraps the domain SMTP result for API responses
type SMTPResponse struct {
	Connected        bool   `json:"connected"`
	SupportsStartTLS bool   `json:"supportsStartTLS"`
	Error            string `json:"error,omitempty"`
}

// FromSMTPResult converts a domain SMTP result to an API response
func FromSMTPResult(result *smtp.SMTPResult) *SMTPResponse {
	if result == nil {
		return &SMTPResponse{
			Error: "no result available",
		}
	}

	// For now, we're creating a simple response with the information we have
	// This can be expanded later to match the full domain model
	return &SMTPResponse{
		Error: result.Error,
	}
}

// EmailAuthResponse wraps the domain email authentication result for API responses
type EmailAuthResponse struct {
	Domain    string `json:"domain"`
	SPF       bool   `json:"spf"`
	DKIM      bool   `json:"dkim"`
	DMARC     bool   `json:"dmarc"`
	AllPassed bool   `json:"allPassed"`
	Error     string `json:"error,omitempty"`
}

// NetworkToolResponse wraps the domain network tool result for API responses
type NetworkToolResponse struct {
	Target    string `json:"target"`
	Tool      string `json:"tool"`
	RawOutput string `json:"rawOutput,omitempty"`
	Error     string `json:"error,omitempty"`
}
