// Package primary contains the primary adapters (implementing input ports)
package primary

import (
	"context"
	"fmt"
	"mxclone/domain/dns"
	"mxclone/ports/output"
)

// DNSAdapter implements the DNS input port
type DNSAdapter struct {
	dnsService *dns.Service
	repository output.DNSRepository
}

// NewDNSAdapter creates a new DNS adapter
func NewDNSAdapter(repository output.DNSRepository) *DNSAdapter {
	return &DNSAdapter{
		dnsService: dns.NewService(),
		repository: repository,
	}
}

// Lookup performs a DNS lookup for a specific record type
func (a *DNSAdapter) Lookup(ctx context.Context, domain string, recordType dns.RecordType) (*dns.DNSResult, error) {
	records, err := a.repository.LookupRecords(ctx, domain, recordType)

	result := a.dnsService.ProcessDNSLookup(ctx, domain, recordType, records)

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	return result, nil
}

// LookupAll performs DNS lookups for all supported record types
func (a *DNSAdapter) LookupAll(ctx context.Context, domain string) (*dns.DNSResult, error) {
	recordTypes := dns.AllRecordTypes()
	results := make([]*dns.DNSResult, 0, len(recordTypes))
	var firstError error

	for _, recordType := range recordTypes {
		result, err := a.Lookup(ctx, domain, recordType)
		results = append(results, result)

		// Keep track of first error, but continue with other lookups
		if err != nil && firstError == nil {
			firstError = fmt.Errorf("%s lookup error: %s", recordType, err.Error())
		}
	}

	aggregatedResult := a.dnsService.AggregateDNSResults(results)

	return aggregatedResult, firstError
}
