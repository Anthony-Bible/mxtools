// Package types provides shared data structures for the MXToolbox clone.
package types

import (
	"time"
)

// Result represents a generic result from any check.
type Result struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// DNSResult represents the result of a DNS lookup.
type DNSResult struct {
	Lookups map[string][]string `json:"lookups"`           // Map of record type to records
	Error   string              `json:"error,omitempty"`
}

// BlacklistResult represents the result of a blacklist check.
type BlacklistResult struct {
	CheckedIP  string            `json:"checkedIp"`
	ListedOn   map[string]string `json:"listedOn,omitempty"` // Map DNSBL zone to reason/status
	CheckError string            `json:"checkError,omitempty"`
}

// SMTPResult represents the result of an SMTP check.
type SMTPResult struct {
	ConnectSuccess  bool          `json:"connectSuccess"`
	ConnectError    string        `json:"connectError,omitempty"`
	SupportsSTARTTLS *bool         `json:"supportsStarttls,omitempty"` // Use pointer for tri-state (yes/no/error)
	STARTTLSError   string        `json:"starttlsError,omitempty"`
	IsOpenRelay     *bool         `json:"isOpenRelay,omitempty"`
	RelayCheckError string        `json:"relayCheckError,omitempty"`
	ResponseTime    time.Duration `json:"responseTime,omitempty"`
}

// AuthResult represents the result of an email authentication check.
type AuthResult struct {
	SPFRecord   string `json:"spfRecord,omitempty"`
	SPFResult   string `json:"spfResult,omitempty"` // Pass, Fail, SoftFail, Neutral, etc.
	SPFError    string `json:"spfError,omitempty"`
	DMARCRecord string `json:"dmarcRecord,omitempty"`
	DMARCPolicy string `json:"dmarcPolicy,omitempty"` // p= tag
	DMARCError  string `json:"dmarcError,omitempty"`
}

// DomainHealthReport represents a comprehensive report for a domain.
type DomainHealthReport struct {
	Domain        string           `json:"domain"`
	Timestamp     time.Time        `json:"timestamp"`
	DNS           *DNSResult       `json:"dns,omitempty"`
	Blacklist     *BlacklistResult `json:"blacklist,omitempty"` // Could be multiple for different IPs
	SMTP          *SMTPResult      `json:"smtp,omitempty"`      // Check each MX server
	Auth          *AuthResult      `json:"auth,omitempty"`
	OverallStatus string           `json:"overallStatus,omitempty"` // e.g., "Healthy", "Issues Found"
}

// CheckRequest represents a request to perform a check.
type CheckRequest struct {
	Target     string   `json:"target"`     // Domain or IP address to check
	CheckTypes []string `json:"checkTypes"` // Types of checks to perform (dns, blacklist, smtp, auth, health)
	Options    map[string]interface{} `json:"options,omitempty"` // Additional options for the check
}

// Job represents a unit of work to be performed.
type Job struct {
	ID      string      `json:"id"`
	Request CheckRequest `json:"request"`
	Result  Result      `json:"result"`
	Done    bool        `json:"done"`
	Error   error       `json:"error,omitempty"`
}
