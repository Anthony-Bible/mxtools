// Package networktools provides auxiliary network diagnostic tools.
package networktools

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// WhoisServer represents a WHOIS server.
type WhoisServer struct {
	Host string
	Port int
}

// DefaultWhoisServer is the default WHOIS server to use.
var DefaultWhoisServer = WhoisServer{
	Host: "whois.iana.org",
	Port: 43,
}

// WhoisResult represents the result of a WHOIS query.
type WhoisResult struct {
	Query       string            `json:"query"`
	RawResponse string            `json:"rawResponse"`
	Fields      map[string]string `json:"fields,omitempty"`
	Referral    string            `json:"referral,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// Whois performs a WHOIS query for the specified domain or IP.
func Whois(ctx context.Context, query string, server WhoisServer, timeout time.Duration) (*WhoisResult, error) {
	result := &WhoisResult{
		Query:  query,
		Fields: make(map[string]string),
	}

	// Connect to the WHOIS server
	address := fmt.Sprintf("%s:%d", server.Host, server.Port)
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	conn, err := dialer.DialContext(ctx, "tcp",
		address)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to connect to WHOIS server: %s", err.Error())
		return result, err
	}
	defer conn.Close()

	// Set a deadline for the connection
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		result.Error = fmt.Sprintf("Failed to set deadline: %s", err.Error())
		return result, err
	}

	// Send the query
	fmt.Println("query", query)
	_, err = conn.Write([]byte(query + "\r\n"))
	if err != nil {
		result.Error = fmt.Sprintf("Failed to send WHOIS query: %s", err.Error())
		return result, err
	}

	// Read the response
	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		response.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		result.Error = fmt.Sprintf("Failed to read WHOIS response: %s", err.Error())
		return result, err
	}

	// Store the raw response
	result.RawResponse = response.String()

	// Parse the response
	parseWhoisResponse(result)

	return result, nil
}

// WhoisWithReferral performs a WHOIS query and follows referrals if necessary.
func WhoisWithReferral(ctx context.Context, query string, timeout time.Duration) (*WhoisResult, error) {
	// Check if the query is a domain name (contains a dot)
	if strings.Contains(query, ".") {
		// Extract the TLD from the domain
		parts := strings.Split(query, ".")
		tld := parts[len(parts)-1]

		// First, query the IANA WHOIS server for the TLD to get the appropriate WHOIS server
		server := DefaultWhoisServer
		tldResult, err := Whois(ctx, tld, server, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to query IANA WHOIS server for TLD: %s", err.Error())
		}

		if tldResult.Referral != "" {
			// We got a referral to the TLD's WHOIS server, now query it for the full domain
			referralServer := WhoisServer{
				Host: tldResult.Referral,
				Port: 43,
			}

			// Perform the query on the referral server with the full domain
			domainResult, err := Whois(ctx, query, referralServer, timeout)
			if err == nil {
				return domainResult, nil
			}
		}
		fmt.Println("referral is not available", tldResult.Referral)
	}

	// If the above approach didn't work or the query is not a domain, fall back to the original method
	server := DefaultWhoisServer

	// Perform the initial query
	fmt.Println("query", query)
	result, err := Whois(ctx, query, server, timeout)
	if err != nil {
		return result, err
	}

	// Check if we got a referral
	if result.Referral != "" {
		// Extract the server from the referral
		referralServer := WhoisServer{
			Host: result.Referral,
			Port: 43,
		}
		fmt.Println("referral", referralServer)

		// Perform the query on the referral server
		referralResult, err := Whois(ctx, query, referralServer, timeout)
		if err != nil {
			// If the referral query fails, return the original result
			return result, nil
		}

		// Return the referral result
		return referralResult, nil
	}

	return result, nil
}

// parseWhoisResponse parses a WHOIS response and extracts fields.
func parseWhoisResponse(result *WhoisResult) {
	// Split the response into lines
	lines := strings.Split(result.RawResponse, "\n")

	// Look for common field patterns
	fieldPattern := regexp.MustCompile(`^\s*([^:]+):\s*(.+)$`)
	referralPattern := regexp.MustCompile(`(?i)whois:\s*(\S+)`)

	for _, line := range lines {
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for field pattern
		if matches := fieldPattern.FindStringSubmatch(line); len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			result.Fields[key] = value

			// Check if this is a referral
			if referralMatches := referralPattern.FindStringSubmatch(line); len(referralMatches) == 2 {
				result.Referral = referralMatches[1]
			}
		}
	}
}

// ParseWhoisFields parses a WHOIS response and extracts common fields.
func ParseWhoisFields(response string) map[string]string {
	fields := make(map[string]string)

	// Split the response into lines
	lines := strings.Split(response, "\n")

	// Look for common field patterns
	fieldPattern := regexp.MustCompile(`^\s*([^:]+):\s*(.+)$`)

	for _, line := range lines {
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for field pattern
		if matches := fieldPattern.FindStringSubmatch(line); len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			fields[key] = value
		}
	}

	return fields
}

// ExtractCommonWhoisFields extracts common fields from a WHOIS result.
func ExtractCommonWhoisFields(result *WhoisResult) map[string]string {
	commonFields := make(map[string]string)

	// Define common field names and their possible variations
	fieldMappings := map[string][]string{
		"Domain Name":     {"domain name", "domain"},
		"Registrar":       {"registrar", "sponsor"},
		"Creation Date":   {"creation date", "created", "registered on", "registration date"},
		"Expiration Date": {"expiration date", "expires", "registry expiry date"},
		"Updated Date":    {"updated date", "last updated", "last-update"},
		"Name Servers":    {"name server", "nameserver", "nserver", "name servers"},
		"Status":          {"status", "domain status"},
		"DNSSEC":          {"dnssec"},
	}

	// Extract common fields
	for commonName, variations := range fieldMappings {
		for _, variation := range variations {
			for field, value := range result.Fields {
				if strings.ToLower(field) == variation {
					commonFields[commonName] = value
					break
				}
			}
		}
	}

	return commonFields
}

// FormatWhoisResult formats a WHOIS result as a string.
func FormatWhoisResult(result *WhoisResult) string {
	if result.Error != "" {
		return fmt.Sprintf("WHOIS query for %s failed: %s", result.Query, result.Error)
	}

	output := fmt.Sprintf("WHOIS information for %s:\n\n", result.Query)

	// Extract common fields
	commonFields := ExtractCommonWhoisFields(result)

	// Display common fields
	if len(commonFields) > 0 {
		output += "Common Fields:\n"
		for field, value := range commonFields {
			output += fmt.Sprintf("  %s: %s\n", field, value)
		}
		output += "\n"
	}

	// Display raw response
	output += "Raw Response:\n"
	output += result.RawResponse

	return output
}
