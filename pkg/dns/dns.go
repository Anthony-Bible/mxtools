// Package dns provides DNS lookup functionality.
package dns

import (
	"context"
	"fmt"
	"net"

	"mxclone/pkg/types"
)

// Lookup performs a DNS lookup for the specified domain and record type.
func Lookup(ctx context.Context, domain string, recordType string) (*types.DNSResult, error) {
	result := &types.DNSResult{
		Lookups: make(map[string][]string),
	}

	var err error
	switch recordType {
	case "A":
		err = lookupA(ctx, domain, result)
	case "AAAA":
		err = lookupAAAA(ctx, domain, result)
	case "MX":
		err = lookupMX(ctx, domain, result)
	case "TXT":
		err = lookupTXT(ctx, domain, result)
	case "CNAME":
		err = lookupCNAME(ctx, domain, result)
	case "NS":
		err = lookupNS(ctx, domain, result)
	case "SOA":
		err = lookupSOA(ctx, domain, result)
	case "PTR":
		err = lookupPTR(ctx, domain, result)
	default:
		return nil, fmt.Errorf("unsupported record type: %s", recordType)
	}

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	return result, nil
}

// LookupAll performs DNS lookups for all supported record types.
func LookupAll(ctx context.Context, domain string) (*types.DNSResult, error) {
	result := &types.DNSResult{
		Lookups: make(map[string][]string),
	}

	recordTypes := []string{"A", "AAAA", "MX", "TXT", "CNAME", "NS", "SOA"}

	for _, recordType := range recordTypes {
		var err error
		switch recordType {
		case "A":
			err = lookupA(ctx, domain, result)
		case "AAAA":
			err = lookupAAAA(ctx, domain, result)
		case "MX":
			err = lookupMX(ctx, domain, result)
		case "TXT":
			err = lookupTXT(ctx, domain, result)
		case "CNAME":
			err = lookupCNAME(ctx, domain, result)
		case "NS":
			err = lookupNS(ctx, domain, result)
		case "SOA":
			err = lookupSOA(ctx, domain, result)
		}

		// Continue with other record types even if one fails
		if err != nil {
			// Just log the error for this record type
			if result.Error == "" {
				result.Error = fmt.Sprintf("%s lookup error: %s", recordType, err.Error())
			}
		}
	}

	return result, nil
}

// lookupA performs an A record lookup.
func lookupA(ctx context.Context, domain string, result *types.DNSResult) error {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		return err
	}

	records := make([]string, 0, len(ips))
	for _, ip := range ips {
		records = append(records, ip.String())
	}

	result.Lookups["A"] = records
	return nil
}

// lookupAAAA performs an AAAA record lookup.
func lookupAAAA(ctx context.Context, domain string, result *types.DNSResult) error {
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip6", domain)
	if err != nil {
		return err
	}

	records := make([]string, 0, len(ips))
	for _, ip := range ips {
		records = append(records, ip.String())
	}

	result.Lookups["AAAA"] = records
	return nil
}

// lookupMX performs an MX record lookup.
func lookupMX(ctx context.Context, domain string, result *types.DNSResult) error {
	mxs, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err != nil {
		return err
	}

	records := make([]string, 0, len(mxs))
	for _, mx := range mxs {
		records = append(records, fmt.Sprintf("%s (priority: %d)", mx.Host, mx.Pref))
	}

	result.Lookups["MX"] = records
	return nil
}

// lookupTXT performs a TXT record lookup.
func lookupTXT(ctx context.Context, domain string, result *types.DNSResult) error {
	txts, err := net.DefaultResolver.LookupTXT(ctx, domain)
	if err != nil {
		return err
	}

	result.Lookups["TXT"] = txts
	return nil
}

// lookupCNAME performs a CNAME record lookup.
func lookupCNAME(ctx context.Context, domain string, result *types.DNSResult) error {
	cname, err := net.DefaultResolver.LookupCNAME(ctx, domain)
	if err != nil {
		return err
	}

	result.Lookups["CNAME"] = []string{cname}
	return nil
}

// lookupNS performs an NS record lookup.
func lookupNS(ctx context.Context, domain string, result *types.DNSResult) error {
	nss, err := net.DefaultResolver.LookupNS(ctx, domain)
	if err != nil {
		return err
	}

	records := make([]string, 0, len(nss))
	for _, ns := range nss {
		records = append(records, ns.Host)
	}

	result.Lookups["NS"] = records
	return nil
}

// lookupSOA performs an SOA record lookup.
func lookupSOA(ctx context.Context, domain string, result *types.DNSResult) error {
	// Standard library doesn't have direct SOA lookup
	// This is a placeholder - we'll implement this with miekg/dns
	result.Lookups["SOA"] = []string{"SOA lookup not implemented with standard library"}
	return nil
}

// lookupPTR performs a PTR record lookup.
func lookupPTR(ctx context.Context, domain string, result *types.DNSResult) error {
	names, err := net.DefaultResolver.LookupAddr(ctx, domain)
	if err != nil {
		return err
	}

	result.Lookups["PTR"] = names
	return nil
}
