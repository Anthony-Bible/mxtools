// Package primary contains the primary adapters (implementing input ports)
package primary

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mxclone/domain/dnsbl"
	"mxclone/ports/output"
)

// DNSBLAdapter implements the DNSBL input port
type DNSBLAdapter struct {
	dnsblService *dnsbl.Service
	repository   output.DNSBLRepository
}

// NewDNSBLAdapter creates a new DNSBL adapter
func NewDNSBLAdapter(repository output.DNSBLRepository) *DNSBLAdapter {
	return &DNSBLAdapter{
		dnsblService: dnsbl.NewService(),
		repository:   repository,
	}
}

// CheckSingleBlacklist checks if an IP is listed on a specific DNSBL
func (a *DNSBLAdapter) CheckSingleBlacklist(ctx context.Context, ip string, zone string, timeout time.Duration) (*dnsbl.BlacklistResult, error) {
	// Build the DNSBL query
	query, err := a.dnsblService.BuildDNSBLQuery(ip, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to build DNSBL query: %w", err)
	}

	// Check if IP is listed
	listed, err := a.repository.LookupDNSBLRecord(ctx, query, timeout)

	// If IP is not listed or there's an error, process and return the result
	if !listed || err != nil {
		return a.dnsblService.ProcessBlacklistResult(ip, zone, false, "", err), nil
	}

	// IP is listed, try to get explanation
	explanation, _ := a.repository.GetDNSBLExplanation(ctx, query, timeout)

	// Process and return the result
	return a.dnsblService.ProcessBlacklistResult(ip, zone, true, explanation, nil), nil
}

// CheckMultipleBlacklists checks if an IP is listed on multiple DNSBLs concurrently
func (a *DNSBLAdapter) CheckMultipleBlacklists(ctx context.Context, ip string, zones []string, timeout time.Duration) (*dnsbl.BlacklistResult, error) {
	// Results channel to collect results from goroutines
	resultsChan := make(chan *dnsbl.BlacklistResult, len(zones))

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Launch a goroutine for each zone
	for _, zone := range zones {
		wg.Add(1)
		go func(z string) {
			defer wg.Done()

			// Create a context with timeout for this specific check
			ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Check if IP is listed on this zone
			result, _ := a.CheckSingleBlacklist(ctxWithTimeout, ip, z, timeout)
			resultsChan <- result
		}(zone)
	}

	// Wait for all goroutines to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := []*dnsbl.BlacklistResult{}
	for result := range resultsChan {
		results = append(results, result)
	}

	// Aggregate results
	aggregatedResult := a.dnsblService.AggregateBlacklistResults(results)

	return aggregatedResult, nil
}

// GetBlacklistSummary returns a human-readable summary of blacklist check results
func (a *DNSBLAdapter) GetBlacklistSummary(result *dnsbl.BlacklistResult) string {
	return a.dnsblService.GetBlacklistSummary(result)
}

// CheckDNSBLHealth checks if a DNSBL is operational
func (a *DNSBLAdapter) CheckDNSBLHealth(ctx context.Context, zone string, timeout time.Duration) (bool, error) {
	return a.repository.CheckDNSBLAvailability(ctx, zone, timeout)
}

// CheckMultipleDNSBLHealth checks the health of multiple DNSBLs concurrently
func (a *DNSBLAdapter) CheckMultipleDNSBLHealth(ctx context.Context, zones []string, timeout time.Duration) map[string]bool {
	healthStatus := make(map[string]bool)

	// Use a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Use a mutex to protect concurrent writes to the health status map
	var mu sync.Mutex

	// Check each zone concurrently
	for _, zone := range zones {
		wg.Add(1)
		go func(z string) {
			defer wg.Done()

			// Create a context with timeout for this specific check
			ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Check if the DNSBL is operational
			healthy, _ := a.CheckDNSBLHealth(ctxWithTimeout, z, timeout)

			// Lock before updating the shared health status map
			mu.Lock()
			healthStatus[z] = healthy
			mu.Unlock()
		}(zone)
	}

	// Wait for all checks to complete
	wg.Wait()

	return healthStatus
}
