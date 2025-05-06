// Package primary contains the primary adapters (implementing input ports)
package primary

import (
	"context"
	"fmt"
	"time"

	"mxclone/domain/networktools"
	"mxclone/ports/output"
)

// NetworkToolsAdapter implements the NetworkTools input port
type NetworkToolsAdapter struct {
	service    *networktools.Service
	repository output.NetworkToolsRepository
}

// NewNetworkToolsAdapter creates a new NetworkTools adapter
func NewNetworkToolsAdapter(repository output.NetworkToolsRepository) *NetworkToolsAdapter {
	return &NetworkToolsAdapter{
		service:    networktools.NewService(),
		repository: repository,
	}
}

// ExecutePing performs a ping operation to a target
func (a *NetworkToolsAdapter) ExecutePing(ctx context.Context, target string, count int, timeout time.Duration) (*networktools.PingResult, error) {
	// Resolve domain name to IP if needed
	resolvedIP, err := a.repository.ResolveDomain(ctx, target)
	if err != nil {
		resolvedIP = target // Use the target as is if resolution fails
	}

	// Execute ping command
	_, rtts, packetsSent, packetsReceived, err := a.repository.ExecutePingCommand(ctx, target, count, timeout)

	// Process and return the result
	result := a.service.ProcessPingResult(target, resolvedIP, rtts, packetsSent, packetsReceived, err)
	return result, nil
}

// ExecuteTraceroute performs a traceroute operation to a target
func (a *NetworkToolsAdapter) ExecuteTraceroute(ctx context.Context, target string, maxHops int, timeout time.Duration) (*networktools.TracerouteResult, error) {
	// Resolve domain name to IP if needed
	resolvedIP, err := a.repository.ResolveDomain(ctx, target)
	if err != nil {
		resolvedIP = target // Use the target as is if resolution fails
	}

	// Execute traceroute command
	_, hops, targetReached, err := a.repository.ExecuteTracerouteCommand(ctx, target, maxHops, timeout)

	// Process and return the result
	result := a.service.ProcessTracerouteResult(target, resolvedIP, hops, targetReached, err)
	return result, nil
}

// ExecuteWHOIS performs a WHOIS lookup for a domain or IP
func (a *NetworkToolsAdapter) ExecuteWHOIS(ctx context.Context, target string, timeout time.Duration) (*networktools.WHOISResult, error) {
	// Execute WHOIS command
	rawData, err := a.repository.ExecuteWHOISCommand(ctx, target, timeout)
	if err != nil {
		return a.service.ProcessWHOISResult(target, "", "", "", "", nil, err), err
	}

	// Parse WHOIS data
	registrar, createdDate, expirationDate, nameServers, parseErr := a.repository.ParseWHOISData(rawData)

	// Process and return the result
	result := a.service.ProcessWHOISResult(target, rawData, registrar, createdDate, expirationDate, nameServers, parseErr)
	return result, nil
}

// ExecuteNetworkTool is a generic method that executes a network diagnostic tool
func (a *NetworkToolsAdapter) ExecuteNetworkTool(ctx context.Context, toolType networktools.ToolType, target string, options map[string]interface{}) (*networktools.NetworkToolResult, error) {
	var pingResult *networktools.PingResult
	var tracerouteResult *networktools.TracerouteResult
	var whoisResult *networktools.WHOISResult
	var err error

	// Set default values for options
	count := 4                  // Default ping count
	maxHops := 30               // Default max hops for traceroute
	timeout := 30 * time.Second // Default timeout

	// Override defaults with provided options
	if val, ok := options["count"].(int); ok {
		count = val
	}

	if val, ok := options["maxHops"].(int); ok {
		maxHops = val
	}

	if val, ok := options["timeout"].(time.Duration); ok {
		timeout = val
	}

	// Execute the appropriate tool based on type
	switch toolType {
	case networktools.ToolTypePing:
		pingResult, err = a.ExecutePing(ctx, target, count, timeout)

	case networktools.ToolTypeTraceroute:
		tracerouteResult, err = a.ExecuteTraceroute(ctx, target, maxHops, timeout)

	case networktools.ToolTypeWHOIS:
		whoisResult, err = a.ExecuteWHOIS(ctx, target, timeout)

	default:
		err = fmt.Errorf("unsupported tool type: %s", toolType)
	}

	// Wrap the result in a generic NetworkToolResult
	result := a.service.WrapResult(toolType, pingResult, tracerouteResult, whoisResult, err)
	return result, err
}

// FormatToolResult returns a human-readable summary of a network tool result
func (a *NetworkToolsAdapter) FormatToolResult(result *networktools.NetworkToolResult) string {
	return a.service.FormatNetworkToolSummary(result)
}
