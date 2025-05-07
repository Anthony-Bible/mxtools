// Package dnsbl contains the core domain logic for DNS-based blacklist operations
package dnsbl

import (
	"fmt"
	"net"
)

// BlacklistResult represents the result of a blacklist check operation
type BlacklistResult struct {
	// The IP address that was checked
	CheckedIP string
	// Map of blacklist zone to explanation text
	ListedOn map[string]string
	// Error message if any
	CheckError string
}

// Service defines the core DNSBL business logic operations
type Service struct {
	// You can inject dependencies here if needed
}

// NewService creates a new DNSBL service
func NewService() *Service {
	return &Service{}
}

// ReverseIP reverses the octets of an IPv4 address for use in DNSBL queries
// Example: 192.168.1.1 becomes 1.1.168.192
func (s *Service) ReverseIP(ip string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	// Extract IPv4 address
	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return "", fmt.Errorf("not an IPv4 address: %s", ip)
	}

	// Reverse the octets
	reversed := fmt.Sprintf("%d.%d.%d.%d", ipv4[3], ipv4[2], ipv4[1], ipv4[0])
	return reversed, nil
}

// BuildDNSBLQuery builds a DNSBL query for the given IP and zone
func (s *Service) BuildDNSBLQuery(ip string, zone string) (string, error) {
	reversedIP, err := s.ReverseIP(ip)
	if err != nil {
		return "", err
	}

	// Construct the DNSBL query
	query := fmt.Sprintf("%s.%s", reversedIP, zone)
	return query, nil
}

// ProcessBlacklistResult creates a BlacklistResult for a given IP and zone with lookup results
func (s *Service) ProcessBlacklistResult(ip string, zone string, isListed bool, explanation string, err error) *BlacklistResult {
	result := &BlacklistResult{
		CheckedIP: ip,
		ListedOn:  make(map[string]string),
	}

	if err != nil {
		result.CheckError = err.Error()
		return result
	}

	if isListed {
		result.ListedOn[zone] = explanation
	}

	return result
}

// AggregateBlacklistResults aggregates the results from multiple DNSBL checks
func (s *Service) AggregateBlacklistResults(results []*BlacklistResult) *BlacklistResult {
	if len(results) == 0 {
		return nil
	}

	// Use the first result as the base
	aggregated := &BlacklistResult{
		CheckedIP: results[0].CheckedIP,
		ListedOn:  make(map[string]string),
	}

	// Aggregate the results
	for _, result := range results {
		// Ensure we're aggregating results for the same IP
		if result.CheckedIP != aggregated.CheckedIP {
			continue
		}

		// Merge the ListedOn maps
		for zone, explanation := range result.ListedOn {
			aggregated.ListedOn[zone] = explanation
		}

		// Merge the error messages
		if result.CheckError != "" {
			if aggregated.CheckError == "" {
				aggregated.CheckError = result.CheckError
			} else {
				aggregated.CheckError += "; " + result.CheckError
			}
		}
	}

	return aggregated
}

// GetBlacklistSummary returns a summary of the blacklist check results
func (s *Service) GetBlacklistSummary(result *BlacklistResult) string {
	if result == nil {
		return "No blacklist check results available"
	}

	if len(result.ListedOn) == 0 {
		return fmt.Sprintf("IP %s is not listed on any blacklists", result.CheckedIP)
	}

	summary := fmt.Sprintf("IP %s is listed on %d blacklists:\n", result.CheckedIP, len(result.ListedOn))
	for zone, explanation := range result.ListedOn {
		if explanation != "" {
			summary += fmt.Sprintf("- %s: %s\n", zone, explanation)
		} else {
			summary += fmt.Sprintf("- %s\n", zone)
		}
	}

	if result.CheckError != "" {
		summary += fmt.Sprintf("\nErrors: %s", result.CheckError)
	}

	return summary
}
