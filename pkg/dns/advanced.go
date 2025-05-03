// Package dns provides DNS lookup functionality.
package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"mxclone/pkg/types"
)

// AdvancedLookup performs a DNS lookup using the miekg/dns library.
func AdvancedLookup(ctx context.Context, domain string, recordType string, server string) (*types.DNSResult, error) {
	result := &types.DNSResult{
		Lookups: make(map[string][]string),
	}

	// If no server is specified, use the system default
	if server == "" {
		config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			return nil, fmt.Errorf("failed to get DNS config: %w", err)
		}
		server = config.Servers[0] + ":" + config.Port
	} else if !strings.Contains(server, ":") {
		// If port is not specified, use the default DNS port
		server = server + ":53"
	}

	// Create a new DNS client
	client := new(dns.Client)
	client.Timeout = 5 * time.Second

	// Create a new DNS message
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dnsTypeFromString(recordType))
	m.RecursionDesired = true

	// Send the query
	r, _, err := client.Exchange(m, server)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Check for errors in the response
	if r.Rcode != dns.RcodeSuccess {
		err = fmt.Errorf("DNS query failed with response code: %s", dns.RcodeToString[r.Rcode])
		result.Error = err.Error()
		return result, err
	}

	// Parse the response
	records := parseResponse(r, recordType)
	if len(records) > 0 {
		result.Lookups[recordType] = records
	}

	return result, nil
}

// AdvancedLookupAll performs DNS lookups for all supported record types using the miekg/dns library.
func AdvancedLookupAll(ctx context.Context, domain string, server string) (*types.DNSResult, error) {
	result := &types.DNSResult{
		Lookups: make(map[string][]string),
	}

	recordTypes := []string{"A", "AAAA", "MX", "TXT", "CNAME", "NS", "SOA", "PTR"}
	
	for _, recordType := range recordTypes {
		// Skip PTR lookup for non-IP addresses
		if recordType == "PTR" && net.ParseIP(domain) == nil {
			continue
		}
		
		res, err := AdvancedLookup(ctx, domain, recordType, server)
		if err == nil && len(res.Lookups[recordType]) > 0 {
			result.Lookups[recordType] = res.Lookups[recordType]
		}
	}

	return result, nil
}

// dnsTypeFromString converts a string record type to a DNS type constant.
func dnsTypeFromString(recordType string) uint16 {
	switch recordType {
	case "A":
		return dns.TypeA
	case "AAAA":
		return dns.TypeAAAA
	case "MX":
		return dns.TypeMX
	case "TXT":
		return dns.TypeTXT
	case "CNAME":
		return dns.TypeCNAME
	case "NS":
		return dns.TypeNS
	case "SOA":
		return dns.TypeSOA
	case "PTR":
		return dns.TypePTR
	default:
		return dns.TypeA
	}
}

// parseResponse parses a DNS response message and extracts the records.
func parseResponse(r *dns.Msg, recordType string) []string {
	var records []string

	for _, answer := range r.Answer {
		switch recordType {
		case "A":
			if a, ok := answer.(*dns.A); ok {
				records = append(records, a.A.String())
			}
		case "AAAA":
			if aaaa, ok := answer.(*dns.AAAA); ok {
				records = append(records, aaaa.AAAA.String())
			}
		case "MX":
			if mx, ok := answer.(*dns.MX); ok {
				records = append(records, fmt.Sprintf("%s (priority: %d)", mx.Mx, mx.Preference))
			}
		case "TXT":
			if txt, ok := answer.(*dns.TXT); ok {
				records = append(records, strings.Join(txt.Txt, " "))
			}
		case "CNAME":
			if cname, ok := answer.(*dns.CNAME); ok {
				records = append(records, cname.Target)
			}
		case "NS":
			if ns, ok := answer.(*dns.NS); ok {
				records = append(records, ns.Ns)
			}
		case "SOA":
			if soa, ok := answer.(*dns.SOA); ok {
				records = append(records, fmt.Sprintf("primary: %s, admin: %s, serial: %d, refresh: %d, retry: %d, expire: %d, ttl: %d",
					soa.Ns, soa.Mbox, soa.Serial, soa.Refresh, soa.Retry, soa.Expire, soa.Minttl))
			}
		case "PTR":
			if ptr, ok := answer.(*dns.PTR); ok {
				records = append(records, ptr.Ptr)
			}
		}
	}

	return records
}