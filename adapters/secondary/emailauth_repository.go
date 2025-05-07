// Package secondary contains the secondary adapters (implementing output ports)
package secondary

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"mxclone/domain/dns"
	"mxclone/ports/input"
)

// EmailAuthRepository implements the EmailAuth repository output port
type EmailAuthRepository struct {
	dnsService input.DNSPort
}

// NewEmailAuthRepository creates a new EmailAuth repository
func NewEmailAuthRepository(dnsService input.DNSPort) *EmailAuthRepository {
	return &EmailAuthRepository{
		dnsService: dnsService,
	}
}

// GetSPFRecord retrieves the SPF record for a domain
func (r *EmailAuthRepository) GetSPFRecord(ctx context.Context, domain string, timeout time.Duration) (string, bool, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Look up TXT records for the domain
	result, err := r.dnsService.Lookup(ctxWithTimeout, domain, dns.TypeTXT)
	if err != nil {
		return "", false, err
	}

	// Look for SPF record in TXT records
	for _, record := range result.Lookups["TXT"] {
		// Check if this is an SPF record (starts with "v=spf1")
		if strings.HasPrefix(strings.ToLower(record), "v=spf1") {
			return record, true, nil
		}
	}

	return "", false, nil
}

// ValidateSPFRecord validates the SPF record and extracts mechanisms
func (r *EmailAuthRepository) ValidateSPFRecord(record string) (bool, []string, error) {
	// Basic validation checks
	if !strings.HasPrefix(strings.ToLower(record), "v=spf1") {
		return false, nil, fmt.Errorf("invalid SPF record: does not start with v=spf1")
	}

	// Extract mechanisms (like ip4:, ip6:, include:, etc.)
	parts := strings.Fields(record)
	mechanisms := make([]string, 0)

	for _, part := range parts[1:] { // Skip "v=spf1"
		// Validate common mechanisms
		if strings.HasPrefix(part, "ip4:") ||
			strings.HasPrefix(part, "ip6:") ||
			strings.HasPrefix(part, "a:") ||
			strings.HasPrefix(part, "mx:") ||
			strings.HasPrefix(part, "include:") ||
			strings.HasPrefix(part, "exists:") ||
			part == "all" ||
			part == "+all" ||
			part == "-all" ||
			part == "~all" ||
			part == "?all" {
			mechanisms = append(mechanisms, part)
		}
	}

	// Check for valid termination with "all" mechanism
	hasAllMechanism := false
	for _, mechanism := range mechanisms {
		if strings.HasSuffix(mechanism, "all") {
			hasAllMechanism = true
			break
		}
	}

	if !hasAllMechanism {
		return false, mechanisms, fmt.Errorf("invalid SPF record: missing terminating 'all' mechanism")
	}

	return true, mechanisms, nil
}

// GetDKIMRecord retrieves the DKIM record for a domain and selector
func (r *EmailAuthRepository) GetDKIMRecord(ctx context.Context, domain string, selector string, timeout time.Duration) (string, bool, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// DKIM records are stored as TXT records at selector._domainkey.domain
	lookupDomain := fmt.Sprintf("%s._domainkey.%s", selector, domain)

	// Look up TXT records
	result, err := r.dnsService.Lookup(ctxWithTimeout, lookupDomain, dns.TypeTXT)
	if err != nil {
		return "", false, err
	}

	// Check if we found a DKIM record
	if len(result.Lookups["TXT"]) > 0 {
		for _, record := range result.Lookups["TXT"] {
			// Basic validation: check if it looks like a DKIM record (contains v=DKIM1)
			if strings.Contains(record, "v=DKIM1") {
				return record, true, nil
			}
		}
	}

	return "", false, nil
}

// ValidateDKIMRecord validates a DKIM record
func (r *EmailAuthRepository) ValidateDKIMRecord(record string) (bool, error) {
	// Basic DKIM record validation
	if !strings.Contains(record, "v=DKIM1") {
		return false, fmt.Errorf("invalid DKIM record: missing v=DKIM1")
	}

	// Check for required tags (k=, p=)
	hasKTag := strings.Contains(record, "k=")
	hasPTag := strings.Contains(record, "p=")

	if !hasKTag || !hasPTag {
		return false, fmt.Errorf("invalid DKIM record: missing required tags")
	}

	return true, nil
}

// GetDMARCRecord retrieves the DMARC record for a domain
func (r *EmailAuthRepository) GetDMARCRecord(ctx context.Context, domain string, timeout time.Duration) (string, bool, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// DMARC records are stored as TXT records at _dmarc.domain
	lookupDomain := fmt.Sprintf("_dmarc.%s", domain)

	// Look up TXT records
	result, err := r.dnsService.Lookup(ctxWithTimeout, lookupDomain, dns.TypeTXT)
	if err != nil {
		return "", false, err
	}

	// Check if we found a DMARC record
	if len(result.Lookups["TXT"]) > 0 {
		for _, record := range result.Lookups["TXT"] {
			// Basic validation: check if it looks like a DMARC record (starts with v=DMARC1)
			if strings.HasPrefix(strings.ToLower(record), "v=dmarc1") {
				return record, true, nil
			}
		}
	}

	return "", false, nil
}

// ParseDMARCRecord parses a DMARC record to extract policy information
func (r *EmailAuthRepository) ParseDMARCRecord(record string) (bool, string, string, int, error) {
	// Basic validation
	if !strings.HasPrefix(strings.ToLower(record), "v=dmarc1") {
		return false, "", "", 0, fmt.Errorf("invalid DMARC record: does not start with v=DMARC1")
	}

	// Extract policy
	policyMatch := regexp.MustCompile(`p=([^;]+)`).FindStringSubmatch(record)
	if len(policyMatch) < 2 {
		return false, "", "", 0, fmt.Errorf("invalid DMARC record: missing required p= tag")
	}
	policy := strings.ToLower(policyMatch[1])

	// Valid policies are "none", "quarantine", or "reject"
	if policy != "none" && policy != "quarantine" && policy != "reject" {
		return false, policy, "", 0, fmt.Errorf("invalid DMARC policy value: %s", policy)
	}

	// Extract subdomain policy (optional)
	subdomainPolicy := ""
	spMatch := regexp.MustCompile(`sp=([^;]+)`).FindStringSubmatch(record)
	if len(spMatch) >= 2 {
		subdomainPolicy = strings.ToLower(spMatch[1])

		// Valid subdomain policies are also "none", "quarantine", or "reject"
		if subdomainPolicy != "none" && subdomainPolicy != "quarantine" && subdomainPolicy != "reject" {
			return false, policy, subdomainPolicy, 0, fmt.Errorf("invalid DMARC subdomain policy value: %s", subdomainPolicy)
		}
	}

	// Extract percentage (optional, defaults to 100)
	percentage := 100
	pctMatch := regexp.MustCompile(`pct=(\d+)`).FindStringSubmatch(record)
	if len(pctMatch) >= 2 {
		var err error
		percentage, err = strconv.Atoi(pctMatch[1])
		if err != nil {
			return false, policy, subdomainPolicy, 0, fmt.Errorf("invalid DMARC percentage value: %s", pctMatch[1])
		}

		// Percentage should be between 0 and 100
		if percentage < 0 || percentage > 100 {
			return false, policy, subdomainPolicy, percentage, fmt.Errorf("invalid DMARC percentage value: %d (should be 0-100)", percentage)
		}
	}

	return true, policy, subdomainPolicy, percentage, nil
}
