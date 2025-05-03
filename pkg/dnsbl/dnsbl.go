// Package dnsbl provides functionality for checking IP addresses against DNS-based blacklists.
package dnsbl

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"mxclone/pkg/dns"
	"mxclone/pkg/types"
)

// ReverseIP reverses the octets of an IPv4 address for use in DNSBL queries.
// Example: 192.168.1.1 becomes 1.1.168.192
func ReverseIP(ip string) (string, error) {
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

// BuildDNSBLQuery builds a DNSBL query for the given IP and zone.
func BuildDNSBLQuery(ip string, zone string) (string, error) {
	reversedIP, err := ReverseIP(ip)
	if err != nil {
		return "", err
	}

	// Construct the DNSBL query
	query := fmt.Sprintf("%s.%s", reversedIP, zone)
	return query, nil
}

// CheckSingleBlacklist checks if an IP is listed on a specific DNSBL.
// It returns true if the IP is listed, along with any explanation text.
func CheckSingleBlacklist(ctx context.Context, ip string, zone string, timeout time.Duration) (bool, string, error) {
	// Build the DNSBL query
	query, err := BuildDNSBLQuery(ip, zone)
	if err != nil {
		return false, "", fmt.Errorf("failed to build DNSBL query: %w", err)
	}

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform A record lookup to check if IP is listed
	result, err := dns.LookupWithRetry(ctxWithTimeout, query, "A", 2, timeout)

	// If there's no A record, the IP is not listed
	if err != nil || len(result.Lookups["A"]) == 0 {
		return false, "", nil
	}

	// IP is listed, try to get explanation from TXT record
	explanation := ""
	txtResult, txtErr := dns.LookupWithRetry(ctxWithTimeout, query, "TXT", 1, timeout)
	if txtErr == nil && len(txtResult.Lookups["TXT"]) > 0 {
		explanation = txtResult.Lookups["TXT"][0]
	}

	return true, explanation, nil
}

// CheckBlacklist checks if an IP is listed on a specific DNSBL and returns a BlacklistResult.
func CheckBlacklist(ctx context.Context, ip string, zone string, timeout time.Duration) *types.BlacklistResult {
	result := &types.BlacklistResult{
		CheckedIP: ip,
		ListedOn:  make(map[string]string),
	}

	listed, explanation, err := CheckSingleBlacklist(ctx, ip, zone, timeout)
	if err != nil {
		result.CheckError = err.Error()
		return result
	}

	if listed {
		result.ListedOn[zone] = explanation
	}

	return result
}

// CheckMultipleBlacklists checks if an IP is listed on multiple DNSBLs concurrently.
func CheckMultipleBlacklists(ctx context.Context, ip string, zones []string, timeout time.Duration) *types.BlacklistResult {
	result := &types.BlacklistResult{
		CheckedIP: ip,
		ListedOn:  make(map[string]string),
	}

	// Use a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Use a mutex to protect concurrent writes to the result
	var mu sync.Mutex

	// Check each zone concurrently
	for _, zone := range zones {
		wg.Add(1)
		go func(z string) {
			defer wg.Done()

			// Create a context with timeout for this specific check
			ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			listed, explanation, err := CheckSingleBlacklist(ctxWithTimeout, ip, z, timeout)

			// Lock before updating the shared result
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if result.CheckError == "" {
					result.CheckError = fmt.Sprintf("Error checking %s: %s", z, err.Error())
				} else {
					result.CheckError += fmt.Sprintf("; Error checking %s: %s", z, err.Error())
				}
				return
			}

			if listed {
				result.ListedOn[z] = explanation
			}
		}(zone)
	}

	// Wait for all checks to complete
	wg.Wait()

	return result
}

// CheckDNSBLHealth checks if a DNSBL is operational by performing a test query.
// It returns true if the DNSBL is operational, false otherwise.
func CheckDNSBLHealth(ctx context.Context, zone string, timeout time.Duration) (bool, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform NS record lookup to check if the DNSBL is operational
	// We're not checking if the IP is listed, just if the DNSBL responds
	_, err := dns.LookupWithRetry(ctxWithTimeout, zone, "NS", 2, timeout)
	if err != nil {
		return false, nil // DNSBL is not operational
	}

	return true, nil // DNSBL is operational
}

// CheckMultipleDNSBLHealth checks the health of multiple DNSBLs concurrently.
// It returns a map of zone to health status.
func CheckMultipleDNSBLHealth(ctx context.Context, zones []string, timeout time.Duration) map[string]bool {
	healthStatus := make(map[string]bool)

	// Use a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Use a mutex to protect concurrent writes to the result
	var mu sync.Mutex

	// Check each zone concurrently
	for _, zone := range zones {
		wg.Add(1)
		go func(z string) {
			defer wg.Done()

			// Create a context with timeout for this specific check
			ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			healthy, _ := CheckDNSBLHealth(ctxWithTimeout, z, timeout)

			// Lock before updating the shared result
			mu.Lock()
			healthStatus[z] = healthy
			mu.Unlock()
		}(zone)
	}

	// Wait for all checks to complete
	wg.Wait()

	return healthStatus
}

// AggregateBlacklistResults aggregates the results from multiple DNSBL checks.
// It returns a summary of the blacklist check results.
func AggregateBlacklistResults(results []*types.BlacklistResult) *types.BlacklistResult {
	if len(results) == 0 {
		return nil
	}

	// Use the first result as the base
	aggregated := &types.BlacklistResult{
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

// GetBlacklistSummary returns a summary of the blacklist check results.
func GetBlacklistSummary(result *types.BlacklistResult) string {
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
