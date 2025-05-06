// Package dns contains the core domain logic for DNS operations
package dns

import (
	"context"
)

// DNSResult represents the result of a DNS lookup operation
type DNSResult struct {
	// Map of record type to records
	Lookups map[string][]string
	// Error message if any
	Error string
}

// RecordType represents a DNS record type
type RecordType string

const (
	TypeA     RecordType = "A"
	TypeAAAA  RecordType = "AAAA"
	TypeMX    RecordType = "MX"
	TypeTXT   RecordType = "TXT"
	TypeCNAME RecordType = "CNAME"
	TypeNS    RecordType = "NS"
	TypeSOA   RecordType = "SOA"
	TypePTR   RecordType = "PTR"
)

// AllRecordTypes returns all supported record types
func AllRecordTypes() []RecordType {
	return []RecordType{TypeA, TypeAAAA, TypeMX, TypeTXT, TypeCNAME, TypeNS, TypeSOA}
}

// Service defines the core DNS business logic operations
type Service struct {
	// You can inject dependencies here if needed
}

// NewService creates a new DNS service
func NewService() *Service {
	return &Service{}
}

// ProcessDNSLookup contains the core business logic for DNS lookups
// This should be independent of how records are actually looked up
func (s *Service) ProcessDNSLookup(ctx context.Context, domain string, recordType RecordType, records []string) *DNSResult {
	result := &DNSResult{
		Lookups: make(map[string][]string),
	}

	result.Lookups[string(recordType)] = records

	return result
}

// AggregateDNSResults combines multiple DNS results into a single result
func (s *Service) AggregateDNSResults(results []*DNSResult) *DNSResult {
	aggregated := &DNSResult{
		Lookups: make(map[string][]string),
	}

	for _, result := range results {
		for recordType, records := range result.Lookups {
			aggregated.Lookups[recordType] = records
		}

		if result.Error != "" {
			if aggregated.Error == "" {
				aggregated.Error = result.Error
			} else {
				aggregated.Error += "; " + result.Error
			}
		}
	}

	return aggregated
}
