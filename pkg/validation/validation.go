// Package validation provides functions for validating and sanitizing user input.
package validation

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Common validation errors
var (
	ErrEmptyInput      = fmt.Errorf("input cannot be empty")
	ErrInvalidDomain   = fmt.Errorf("invalid domain name")
	ErrInvalidIP       = fmt.Errorf("invalid IP address")
	ErrInvalidPort     = fmt.Errorf("invalid port number")
	ErrInvalidServer   = fmt.Errorf("invalid server address")
	ErrInvalidSelector = fmt.Errorf("invalid selector")
	ErrInvalidFile     = fmt.Errorf("invalid file path")
	ErrInvalidRecordType = fmt.Errorf("invalid DNS record type")
)

// ValidateDomain validates a domain name.
func ValidateDomain(domain string) error {
	if domain == "" {
		return ErrEmptyInput
	}

	// Simple domain validation using regexp
	// This is a basic check and doesn't validate all possible valid domains
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if !domainRegex.MatchString(domain) {
		return ErrInvalidDomain
	}

	return nil
}

// ValidateIP validates an IP address.
func ValidateIP(ip string) error {
	if ip == "" {
		return ErrEmptyInput
	}

	// Use net.ParseIP to validate the IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ErrInvalidIP
	}

	return nil
}

// ValidatePort validates a port number.
func ValidatePort(port int) error {
	if port < 0 || port > 65535 {
		return ErrInvalidPort
	}

	return nil
}

// ValidateServer validates a server address (hostname:port or IP:port).
func ValidateServer(server string) error {
	if server == "" {
		return ErrEmptyInput
	}

	// If server contains a port (hostname:port or IP:port)
	if strings.Contains(server, ":") {
		host, portStr, err := net.SplitHostPort(server)
		if err != nil {
			return ErrInvalidServer
		}

		// Validate host
		if err := ValidateHost(host); err != nil {
			return err
		}

		// Validate port
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return ErrInvalidPort
		}
		if err := ValidatePort(port); err != nil {
			return err
		}
	} else {
		// Server is just a hostname or IP
		if err := ValidateHost(server); err != nil {
			return err
		}
	}

	return nil
}

// ValidateHost validates a hostname or IP address.
func ValidateHost(host string) error {
	if host == "" {
		return ErrEmptyInput
	}

	// Try to parse as IP address
	if net.ParseIP(host) != nil {
		return nil
	}

	// Try to validate as domain name
	return ValidateDomain(host)
}

// ValidateSelector validates a DKIM selector.
func ValidateSelector(selector string) error {
	if selector == "" {
		return ErrEmptyInput
	}

	// Simple selector validation using regexp
	selectorRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\_]{0,61}[a-zA-Z0-9])?$`)
	if !selectorRegex.MatchString(selector) {
		return ErrInvalidSelector
	}

	return nil
}

// ValidateFile validates a file path.
func ValidateFile(filePath string) error {
	if filePath == "" {
		return ErrEmptyInput
	}

	// Check if file exists and is readable
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: file does not exist", ErrInvalidFile)
		}
		return fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}

	return nil
}

// ValidateURL validates a URL.
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return ErrEmptyInput
	}

	// Parse the URL
	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	return nil
}

// ValidateDNSRecordType validates a DNS record type.
func ValidateDNSRecordType(recordType string) error {
	if recordType == "" {
		return ErrEmptyInput
	}

	// List of valid DNS record types
	validTypes := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CNAME": true,
		"MX":    true,
		"NS":    true,
		"PTR":   true,
		"SOA":   true,
		"SRV":   true,
		"TXT":   true,
	}

	// Convert to uppercase for case-insensitive comparison
	recordType = strings.ToUpper(recordType)

	if !validTypes[recordType] {
		return ErrInvalidRecordType
	}

	return nil
}

// SanitizeDomain sanitizes a domain name.
func SanitizeDomain(domain string) string {
	// Remove any whitespace
	domain = strings.TrimSpace(domain)
	
	// Convert to lowercase
	domain = strings.ToLower(domain)
	
	// Remove any trailing dot
	domain = strings.TrimSuffix(domain, ".")
	
	return domain
}

// SanitizeIP sanitizes an IP address.
func SanitizeIP(ip string) string {
	// Remove any whitespace
	ip = strings.TrimSpace(ip)
	
	return ip
}

// SanitizeSelector sanitizes a DKIM selector.
func SanitizeSelector(selector string) string {
	// Remove any whitespace
	selector = strings.TrimSpace(selector)
	
	return selector
}

// SanitizeDNSRecordType sanitizes a DNS record type.
func SanitizeDNSRecordType(recordType string) string {
	// Remove any whitespace
	recordType = strings.TrimSpace(recordType)
	
	// Convert to uppercase
	recordType = strings.ToUpper(recordType)
	
	return recordType
}