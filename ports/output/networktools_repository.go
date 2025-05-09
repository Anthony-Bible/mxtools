// Package output contains the output ports (interfaces) for the application
package output

import (
	"context"
	"mxclone/domain/networktools"
	"time"
)

// NetworkToolsRepository defines the output interface for network diagnostic tools
type NetworkToolsRepository interface {
	// ExecutePingCommand executes the actual ping command/operation
	ExecutePingCommand(ctx context.Context, target string, count int, timeout time.Duration) (string, []time.Duration, int, int, error)

	// ResolveDomain resolves a domain name to an IP address
	ResolveDomain(ctx context.Context, domain string) (string, error)

	// ExecuteTracerouteCommand executes the actual traceroute command/operation
	ExecuteTracerouteCommand(ctx context.Context, target string, maxHops int, timeout time.Duration) (string, []networktools.TracerouteHop, bool, error)

	// ResolveIP resolves an IP address to a hostname
	ResolveIP(ctx context.Context, ip string) (string, error)

	// ExecuteWHOISCommand executes the actual WHOIS command/operation
	ExecuteWHOISCommand(ctx context.Context, target string, timeout time.Duration) (string, error)

	// ParseWHOISData parses raw WHOIS data to extract structured information
	ParseWHOISData(rawData string) (string, string, string, []string, error)

	// ExecuteTracerouteHop performs a single traceroute hop (with given TTL) to the target.
	ExecuteTracerouteHop(ctx context.Context, target string, ttl int, timeout time.Duration) (networktools.TracerouteHop, bool, error)
}
