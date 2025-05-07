// Package secondary contains the secondary adapters (implementing output ports)
package secondary

import (
	"context"
	"mxclone/domain/dns"
	"mxclone/ports/input"
	"time"
)

// DNSBLRepository implements the DNSBL repository output port
type DNSBLRepository struct {
	dnsService input.DNSPort
}

// NewDNSBLRepository creates a new DNSBL repository
func NewDNSBLRepository(dnsService input.DNSPort) *DNSBLRepository {
	return &DNSBLRepository{
		dnsService: dnsService,
	}
}

// LookupDNSBLRecord checks if a domain (reversed IP + zone) has A records
func (r *DNSBLRepository) LookupDNSBLRecord(ctx context.Context, query string, timeout time.Duration) (bool, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform A record lookup to check if IP is listed
	result, err := r.dnsService.Lookup(ctxWithTimeout, query, dns.TypeA)

	// If there's an error or no A record, the IP is not listed
	if err != nil || len(result.Lookups["A"]) == 0 {
		return false, err
	}

	return true, nil
}

// GetDNSBLExplanation tries to retrieve TXT record explanation for a DNSBL listing
func (r *DNSBLRepository) GetDNSBLExplanation(ctx context.Context, query string, timeout time.Duration) (string, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform TXT record lookup to get explanation
	result, err := r.dnsService.Lookup(ctxWithTimeout, query, dns.TypeTXT)

	// If there's an error or no TXT record, there's no explanation
	if err != nil || len(result.Lookups["TXT"]) == 0 {
		return "", err
	}

	return result.Lookups["TXT"][0], nil
}

// CheckDNSBLAvailability checks if a DNSBL service is available
func (r *DNSBLRepository) CheckDNSBLAvailability(ctx context.Context, zone string, timeout time.Duration) (bool, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform NS record lookup to check if the DNSBL is operational
	// We're not checking if the IP is listed, just if the DNSBL responds
	_, err := r.dnsService.Lookup(ctxWithTimeout, zone, dns.TypeNS)
	if err != nil {
		return false, nil // DNSBL is not operational
	}

	return true, nil // DNSBL is operational
}
