// Package input contains the input ports (interfaces) for the application
package input

import (
	"context"
	"mxclone/domain/networktools"
	"time"
)

// NetworkToolsPort defines the input interface for network diagnostic tools
type NetworkToolsPort interface {
	// ExecutePing performs a ping operation to a target
	ExecutePing(ctx context.Context, target string, count int, timeout time.Duration) (*networktools.PingResult, error)

	// ExecuteTraceroute performs a traceroute operation to a target
	ExecuteTraceroute(ctx context.Context, target string, maxHops int, timeout time.Duration) (*networktools.TracerouteResult, error)

	// ExecuteWHOIS performs a WHOIS lookup for a domain or IP
	ExecuteWHOIS(ctx context.Context, target string, timeout time.Duration) (*networktools.WHOISResult, error)

	// ExecuteNetworkTool is a generic method that executes a network diagnostic tool
	ExecuteNetworkTool(ctx context.Context, toolType networktools.ToolType, target string, options map[string]interface{}) (*networktools.NetworkToolResult, error)

	// WrapResult wraps individual tool results into a generic NetworkToolResult
	WrapResult(toolType networktools.ToolType, pingResult *networktools.PingResult, tracerouteResult *networktools.TracerouteResult, whoisResult *networktools.WHOISResult, err error) *networktools.NetworkToolResult

	// FormatToolResult returns a human-readable summary of a network tool result
	FormatToolResult(result *networktools.NetworkToolResult) string

	// ResolveDomain resolves a domain name to an IP address
	ResolveDomain(ctx context.Context, domain string) (string, error)

	// TracerouteHop performs a single hop of a traceroute to the target with the given TTL.
	TracerouteHop(ctx context.Context, target string, ttl int, timeout time.Duration) (networktools.TracerouteHop, bool, error)
}
