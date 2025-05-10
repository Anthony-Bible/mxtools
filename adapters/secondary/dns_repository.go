// Package secondary contains the secondary adapters (implementing output ports)
package secondary

import (
	"context"
	"fmt"
	"mxclone/domain/dns"
	"net"

	mdns "github.com/miekg/dns"
)

// DNSRepository implements the DNS repository output port
type DNSRepository struct {
	// You could inject a resolver here if needed
	resolver *net.Resolver
}

// NewDNSRepository creates a new DNS repository
func NewDNSRepository() *DNSRepository {
	return &DNSRepository{
		resolver: net.DefaultResolver,
	}
}

// LookupRecords performs the actual DNS lookup operation
func (r *DNSRepository) LookupRecords(ctx context.Context, domain string, recordType dns.RecordType) ([]string, error) {
	switch recordType {
	case dns.TypeA:
		return r.lookupA(ctx, domain)
	case dns.TypeAAAA:
		return r.lookupAAAA(ctx, domain)
	case dns.TypeMX:
		return r.lookupMX(ctx, domain)
	case dns.TypeTXT:
		return r.lookupTXT(ctx, domain)
	case dns.TypeCNAME:
		return r.lookupCNAME(ctx, domain)
	case dns.TypeNS:
		return r.lookupNS(ctx, domain)
	case dns.TypeSOA:
		return r.lookupSOA(ctx, domain)
	case dns.TypePTR:
		return r.lookupPTR(ctx, domain)
	default:
		return nil, fmt.Errorf("unsupported record type: %s", recordType)
	}
}

// lookupA performs an A record lookup
func (r *DNSRepository) lookupA(ctx context.Context, domain string) ([]string, error) {
	ips, err := r.resolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		return nil, err
	}

	records := make([]string, 0, len(ips))
	for _, ip := range ips {
		records = append(records, ip.String())
	}

	return records, nil
}

// lookupAAAA performs an AAAA record lookup
func (r *DNSRepository) lookupAAAA(ctx context.Context, domain string) ([]string, error) {
	ips, err := r.resolver.LookupIP(ctx, "ip6", domain)
	if err != nil {
		return nil, err
	}

	records := make([]string, 0, len(ips))
	for _, ip := range ips {
		records = append(records, ip.String())
	}

	return records, nil
}

// lookupMX performs an MX record lookup
func (r *DNSRepository) lookupMX(ctx context.Context, domain string) ([]string, error) {
	mxs, err := r.resolver.LookupMX(ctx, domain)
	if err != nil {
		return nil, err
	}

	records := make([]string, 0, len(mxs))
	for _, mx := range mxs {
		records = append(records, fmt.Sprintf("%s (priority: %d)", mx.Host, mx.Pref))
	}

	return records, nil
}

// lookupTXT performs a TXT record lookup
func (r *DNSRepository) lookupTXT(ctx context.Context, domain string) ([]string, error) {
	return r.resolver.LookupTXT(ctx, domain)
}

// lookupCNAME performs a CNAME record lookup
func (r *DNSRepository) lookupCNAME(ctx context.Context, domain string) ([]string, error) {
	cname, err := r.resolver.LookupCNAME(ctx, domain)
	if err != nil {
		return nil, err
	}

	return []string{cname}, nil
}

// lookupNS performs an NS record lookup
func (r *DNSRepository) lookupNS(ctx context.Context, domain string) ([]string, error) {
	nss, err := r.resolver.LookupNS(ctx, domain)
	if err != nil {
		return nil, err
	}

	records := make([]string, 0, len(nss))
	for _, ns := range nss {
		records = append(records, ns.Host)
	}

	return records, nil
}

// lookupSOA performs an SOA record lookup
func (r *DNSRepository) lookupSOA(ctx context.Context, domain string) ([]string, error) {
	// Use miekg/dns for SOA lookup
	m := new(mdns.Msg)
	m.SetQuestion(mdns.Fqdn(domain), mdns.TypeSOA)

	client := new(mdns.Client)
	// Use system resolver if possible
	server := ""
	config, err := mdns.ClientConfigFromFile("/etc/resolv.conf")
	if err == nil && len(config.Servers) > 0 {
		server = config.Servers[0] + ":" + config.Port
	} else {
		server = "8.8.8.8:53" // fallback to Google DNS
	}

	resp, _, err := client.Exchange(m, server)
	if err != nil {
		return nil, fmt.Errorf("SOA lookup failed: %w", err)
	}
	if resp.Rcode != mdns.RcodeSuccess {
		return nil, fmt.Errorf("SOA lookup failed with response code: %s", mdns.RcodeToString[resp.Rcode])
	}

	records := make([]string, 0)
	for _, answer := range resp.Answer {
		if soa, ok := answer.(*mdns.SOA); ok {
			records = append(records, fmt.Sprintf(
				"primary: %s, admin: %s, serial: %d, refresh: %d, retry: %d, expire: %d, ttl: %d",
				soa.Ns, soa.Mbox, soa.Serial, soa.Refresh, soa.Retry, soa.Expire, soa.Minttl,
			))
		}
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no SOA record found for domain: %s", domain)
	}
	return records, nil
}

// lookupPTR performs a PTR record lookup
func (r *DNSRepository) lookupPTR(ctx context.Context, domain string) ([]string, error) {
	return r.resolver.LookupAddr(ctx, domain)
}
