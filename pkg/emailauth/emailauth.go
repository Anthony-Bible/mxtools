// Package emailauth provides functionality for checking email authentication mechanisms.
package emailauth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mxclone/pkg/dns"
	"mxclone/pkg/types"
)

// SPFRecord represents a parsed SPF record.
type SPFRecord struct {
	Raw       string
	Version   string
	Mechanisms []string
	Modifiers  map[string]string
	Redirect  string
	All       string // +all, -all, ~all, ?all
}

// DMARCRecord represents a parsed DMARC record.
type DMARCRecord struct {
	Raw      string
	Version  string
	Policy   string // p tag
	SubPolicy string // sp tag
	Pct      int    // pct tag
	RUA      string // rua tag
	RUF      string // ruf tag
	ADKIM    string // adkim tag
	ASPF     string // aspf tag
}

// DKIMRecord represents a parsed DKIM record.
type DKIMRecord struct {
	Raw       string
	Version   string
	PublicKey string
	KeyType   string
	Notes     string
	Service   string
	Flags     string
}

// GetSPFRecord retrieves the SPF record for a domain.
func GetSPFRecord(ctx context.Context, domain string, timeout time.Duration) (string, error) {
	// SPF records are stored as TXT records
	result, err := dns.LookupWithRetry(ctx, domain, "TXT", 2, timeout)
	if err != nil {
		return "", fmt.Errorf("failed to lookup TXT records: %w", err)
	}

	// Look for SPF record in TXT records
	for _, txt := range result.Lookups["TXT"] {
		if strings.HasPrefix(strings.TrimSpace(txt), "v=spf1") {
			return txt, nil
		}
	}

	return "", fmt.Errorf("no SPF record found for domain: %s", domain)
}

// ParseSPFRecord parses an SPF record string into a structured format.
func ParseSPFRecord(record string) (*SPFRecord, error) {
	if record == "" {
		return nil, fmt.Errorf("empty SPF record")
	}

	spf := &SPFRecord{
		Raw:       record,
		Mechanisms: []string{},
		Modifiers:  make(map[string]string),
	}

	// Split the record into parts
	parts := strings.Fields(record)

	// First part should be the version
	if !strings.HasPrefix(parts[0], "v=spf1") {
		return nil, fmt.Errorf("invalid SPF record: does not start with v=spf1")
	}
	spf.Version = parts[0]

	// Process the rest of the parts
	for _, part := range parts[1:] {
		// Check if it's a modifier
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				spf.Modifiers[key] = value

				// Special handling for redirect
				if key == "redirect" {
					spf.Redirect = value
				}
			}
		} else {
			// It's a mechanism
			spf.Mechanisms = append(spf.Mechanisms, part)

			// Special handling for all mechanism
			if strings.HasSuffix(part, "all") {
				spf.All = part
			}
		}
	}

	return spf, nil
}

// ValidateSPF validates an SPF record.
func ValidateSPF(spf *SPFRecord) (string, error) {
	if spf == nil {
		return "", fmt.Errorf("nil SPF record")
	}

	// Check if the record has a valid version
	if !strings.HasPrefix(spf.Version, "v=spf1") {
		return "", fmt.Errorf("invalid SPF version: %s", spf.Version)
	}

	// Check if the record has an all mechanism
	if spf.All == "" && spf.Redirect == "" {
		return "neutral", fmt.Errorf("SPF record has no all mechanism or redirect")
	}

	// Determine the result based on the all mechanism
	switch spf.All {
	case "+all":
		return "pass", nil
	case "-all":
		return "fail", nil
	case "~all":
		return "softfail", nil
	case "?all":
		return "neutral", nil
	default:
		if spf.Redirect != "" {
			return "redirect", nil
		}
		return "neutral", nil
	}
}

// GetDMARCRecord retrieves the DMARC record for a domain.
func GetDMARCRecord(ctx context.Context, domain string, timeout time.Duration) (string, error) {
	// DMARC records are stored as TXT records at _dmarc.domain
	dmarcDomain := fmt.Sprintf("_dmarc.%s", domain)
	result, err := dns.LookupWithRetry(ctx, dmarcDomain, "TXT", 2, timeout)
	if err != nil {
		return "", fmt.Errorf("failed to lookup DMARC record: %w", err)
	}

	// Look for DMARC record in TXT records
	for _, txt := range result.Lookups["TXT"] {
		if strings.HasPrefix(strings.TrimSpace(txt), "v=DMARC1") {
			return txt, nil
		}
	}

	return "", fmt.Errorf("no DMARC record found for domain: %s", domain)
}

// ParseDMARCRecord parses a DMARC record string into a structured format.
func ParseDMARCRecord(record string) (*DMARCRecord, error) {
	if record == "" {
		return nil, fmt.Errorf("empty DMARC record")
	}

	dmarc := &DMARCRecord{
		Raw:     record,
		Pct:     100, // Default value
	}

	// Split the record into parts
	parts := strings.Split(record, ";")

	// First part should be the version
	if !strings.HasPrefix(strings.TrimSpace(parts[0]), "v=DMARC1") {
		return nil, fmt.Errorf("invalid DMARC record: does not start with v=DMARC1")
	}
	dmarc.Version = strings.TrimSpace(parts[0])

	// Process the rest of the parts
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "p":
			dmarc.Policy = value
		case "sp":
			dmarc.SubPolicy = value
		case "pct":
			// Ignore conversion errors, keep default
			fmt.Sscanf(value, "%d", &dmarc.Pct)
		case "rua":
			dmarc.RUA = value
		case "ruf":
			dmarc.RUF = value
		case "adkim":
			dmarc.ADKIM = value
		case "aspf":
			dmarc.ASPF = value
		}
	}

	return dmarc, nil
}

// GetDKIMRecord retrieves the DKIM record for a domain and selector.
func GetDKIMRecord(ctx context.Context, domain, selector string, timeout time.Duration) (string, error) {
	// DKIM records are stored as TXT records at selector._domainkey.domain
	dkimDomain := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	result, err := dns.LookupWithRetry(ctx, dkimDomain, "TXT", 2, timeout)
	if err != nil {
		return "", fmt.Errorf("failed to lookup DKIM record: %w", err)
	}

	// Look for DKIM record in TXT records
	for _, txt := range result.Lookups["TXT"] {
		if strings.Contains(txt, "v=DKIM1") {
			return txt, nil
		}
	}

	return "", fmt.Errorf("no DKIM record found for selector %s at domain: %s", selector, domain)
}

// ParseDKIMRecord parses a DKIM record string into a structured format.
func ParseDKIMRecord(record string) (*DKIMRecord, error) {
	if record == "" {
		return nil, fmt.Errorf("empty DKIM record")
	}

	dkim := &DKIMRecord{
		Raw: record,
	}

	// Split the record into parts
	parts := strings.Split(record, ";")

	// Process each part
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "v":
			dkim.Version = value
		case "k":
			dkim.KeyType = value
		case "p":
			dkim.PublicKey = value
		case "n":
			dkim.Notes = value
		case "s":
			dkim.Service = value
		case "t":
			dkim.Flags = value
		}
	}

	return dkim, nil
}

// AnalyzeEmailHeader analyzes an email header for authentication results.
func AnalyzeEmailHeader(header string) (map[string]string, error) {
	results := make(map[string]string)

	// Look for Authentication-Results header
	lines := strings.Split(header, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Authentication-Results:") {
			// Extract the authentication results
			authResults := strings.TrimPrefix(line, "Authentication-Results:")
			authResults = strings.TrimSpace(authResults)

			// Parse the authentication results
			parts := strings.Split(authResults, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				// Check for SPF result
				if strings.HasPrefix(part, "spf=") {
					results["SPF"] = strings.TrimPrefix(part, "spf=")
				}

				// Check for DKIM result
				if strings.HasPrefix(part, "dkim=") {
					results["DKIM"] = strings.TrimPrefix(part, "dkim=")
				}

				// Check for DMARC result
				if strings.HasPrefix(part, "dmarc=") {
					results["DMARC"] = strings.TrimPrefix(part, "dmarc=")
				}
			}
		}
	}

	return results, nil
}

// CheckEmailAuth performs a comprehensive email authentication check for a domain.
func CheckEmailAuth(ctx context.Context, domain string, timeout time.Duration) (*types.AuthResult, error) {
	result := &types.AuthResult{}

	// Get SPF record
	spfRecord, err := GetSPFRecord(ctx, domain, timeout)
	if err != nil {
		result.SPFError = err.Error()
	} else {
		result.SPFRecord = spfRecord

		// Parse and validate SPF record
		parsedSPF, err := ParseSPFRecord(spfRecord)
		if err != nil {
			result.SPFError = fmt.Sprintf("Failed to parse SPF record: %s", err.Error())
		} else {
			spfResult, err := ValidateSPF(parsedSPF)
			if err != nil {
				result.SPFError = fmt.Sprintf("SPF validation warning: %s", err.Error())
			}
			result.SPFResult = spfResult
		}
	}

	// Get DMARC record
	dmarcRecord, err := GetDMARCRecord(ctx, domain, timeout)
	if err != nil {
		result.DMARCError = err.Error()
	} else {
		result.DMARCRecord = dmarcRecord

		// Parse DMARC record
		parsedDMARC, err := ParseDMARCRecord(dmarcRecord)
		if err != nil {
			result.DMARCError = fmt.Sprintf("Failed to parse DMARC record: %s", err.Error())
		} else {
			result.DMARCPolicy = parsedDMARC.Policy
		}
	}

	return result, nil
}

// CheckEmailAuthWithDKIM performs a comprehensive email authentication check for a domain, including DKIM.
func CheckEmailAuthWithDKIM(ctx context.Context, domain string, selector string, timeout time.Duration) (*types.AuthResult, error) {
	// First get the basic auth results
	result, err := CheckEmailAuth(ctx, domain, timeout)
	if err != nil {
		return result, err
	}

	// Get DKIM record if a selector is provided
	if selector != "" {
		dkimRecord, err := GetDKIMRecord(ctx, domain, selector, timeout)
		if err != nil {
			result.DKIMError = err.Error()
		} else {
			result.DKIMRecord = dkimRecord

			// Parse DKIM record
			parsedDKIM, err := ParseDKIMRecord(dkimRecord)
			if err != nil {
				result.DKIMError = fmt.Sprintf("Failed to parse DKIM record: %s", err.Error())
			} else {
				// Simple validation - check if public key exists
				if parsedDKIM.PublicKey == "" {
					result.DKIMResult = "invalid"
					result.DKIMError = "No public key found in DKIM record"
				} else {
					result.DKIMResult = "valid"
				}
			}
		}
	}

	return result, nil
}

// CheckEmailAuthWithHeader performs a comprehensive email authentication check for a domain and analyzes an email header.
func CheckEmailAuthWithHeader(ctx context.Context, domain string, header string, timeout time.Duration) (*types.AuthResult, error) {
	// First get the basic auth results
	result, err := CheckEmailAuth(ctx, domain, timeout)
	if err != nil {
		return result, err
	}

	// Analyze the email header if provided
	if header != "" {
		headerResults, err := AnalyzeEmailHeader(header)
		if err != nil {
			return result, fmt.Errorf("failed to analyze email header: %w", err)
		}
		result.HeaderAuth = headerResults
	}

	return result, nil
}
