// Package networktools contains the core domain logic for network diagnostic tools
package networktools

import (
	"fmt"
	"time"
)

// ToolType represents the type of network diagnostic tool
type ToolType string

const (
	// ToolTypePing represents ICMP echo request/reply tool
	ToolTypePing ToolType = "ping"
	// ToolTypeTraceroute represents network path tracing tool
	ToolTypeTraceroute ToolType = "traceroute"
	// ToolTypeWHOIS represents domain registration information lookup tool
	ToolTypeWHOIS ToolType = "whois"
)

// PingResult represents the result of a ping operation
type PingResult struct {
	// Target hostname or IP that was pinged
	Target string
	// IP address resolved for the target
	ResolvedIP string
	// Whether the ping was successful
	Success bool
	// Round-trip times for each ping attempt
	RTTs []time.Duration
	// Average round-trip time
	AvgRTT time.Duration
	// Minimum round-trip time
	MinRTT time.Duration
	// Maximum round-trip time
	MaxRTT time.Duration
	// Number of packets sent
	PacketsSent int
	// Number of packets received
	PacketsReceived int
	// Packet loss percentage
	PacketLoss float64
	// Error message if any
	Error string
}

// TracerouteHop represents a single hop in a traceroute path
type TracerouteHop struct {
	// Hop number
	Number int
	// IP address of the hop
	IP string
	// Hostname of the hop (if resolvable)
	Hostname string
	// RTT for this hop
	RTT time.Duration
	// Error message if any
	Error string
}

// TracerouteResult represents the result of a traceroute operation
type TracerouteResult struct {
	// Target hostname or IP
	Target string
	// IP address resolved for the target
	ResolvedIP string
	// Hops in the traceroute path
	Hops []TracerouteHop
	// Whether the target was reached
	TargetReached bool
	// Error message if any
	Error string
}

// WHOISResult represents the result of a WHOIS query
type WHOISResult struct {
	// Target domain or IP
	Target string
	// Raw WHOIS data
	RawData string
	// Registrar information
	Registrar string
	// Creation date
	CreatedDate string
	// Expiration date
	ExpirationDate string
	// Name servers
	NameServers []string
	// Error message if any
	Error string
}

// NetworkToolResult is a generic container for any network tool result
type NetworkToolResult struct {
	// Type of tool that was executed
	ToolType ToolType
	// The ping result (if ToolType is ToolTypePing)
	PingResult *PingResult
	// The traceroute result (if ToolType is ToolTypeTraceroute)
	TracerouteResult *TracerouteResult
	// The WHOIS result (if ToolType is ToolTypeWHOIS)
	WHOISResult *WHOISResult
	// Error message if any
	Error string
}

// Service defines the core NetworkTools business logic operations
type Service struct {
	// You can inject dependencies here if needed
}

// NewService creates a new NetworkTools service
func NewService() *Service {
	return &Service{}
}

// ProcessPingResult processes the result of a ping operation
func (s *Service) ProcessPingResult(target, resolvedIP string, rtts []time.Duration,
	packetsSent, packetsReceived int, err error) *PingResult {

	result := &PingResult{
		Target:          target,
		ResolvedIP:      resolvedIP,
		RTTs:            rtts,
		PacketsSent:     packetsSent,
		PacketsReceived: packetsReceived,
	}

	// Calculate success status
	result.Success = packetsReceived > 0 && err == nil

	// Calculate packet loss percentage
	if packetsSent > 0 {
		result.PacketLoss = 100 - (float64(packetsReceived) / float64(packetsSent) * 100)
	}

	// Calculate RTT statistics if we have any successful pings
	if len(rtts) > 0 {
		// Calculate average RTT
		var totalRTT time.Duration
		for _, rtt := range rtts {
			totalRTT += rtt
		}
		result.AvgRTT = totalRTT / time.Duration(len(rtts))

		// Calculate min and max RTT
		result.MinRTT = rtts[0]
		result.MaxRTT = rtts[0]
		for _, rtt := range rtts {
			if rtt < result.MinRTT {
				result.MinRTT = rtt
			}
			if rtt > result.MaxRTT {
				result.MaxRTT = rtt
			}
		}
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// ProcessTracerouteResult processes the result of a traceroute operation
func (s *Service) ProcessTracerouteResult(target, resolvedIP string, hops []TracerouteHop, targetReached bool, err error) *TracerouteResult {
	result := &TracerouteResult{
		Target:        target,
		ResolvedIP:    resolvedIP,
		Hops:          hops,
		TargetReached: targetReached,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// ProcessWHOISResult processes the result of a WHOIS query
func (s *Service) ProcessWHOISResult(target, rawData, registrar, createdDate, expirationDate string, nameServers []string, err error) *WHOISResult {
	result := &WHOISResult{
		Target:         target,
		RawData:        rawData,
		Registrar:      registrar,
		CreatedDate:    createdDate,
		ExpirationDate: expirationDate,
		NameServers:    nameServers,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// WrapResult wraps a specific tool result into a generic NetworkToolResult
func (s *Service) WrapResult(toolType ToolType, pingResult *PingResult, tracerouteResult *TracerouteResult, whoisResult *WHOISResult, err error) *NetworkToolResult {
	result := &NetworkToolResult{
		ToolType:         toolType,
		PingResult:       pingResult,
		TracerouteResult: tracerouteResult,
		WHOISResult:      whoisResult,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

// FormatPingSummary returns a human-readable summary of a ping result
func (s *Service) FormatPingSummary(result *PingResult) string {
	if result == nil {
		return "No ping results available"
	}

	summary := fmt.Sprintf("PING results for %s (%s):\n", result.Target, result.ResolvedIP)

	if !result.Success {
		summary += "Status: Failed\n"
		if result.Error != "" {
			summary += fmt.Sprintf("Error: %s\n", result.Error)
		}
		return summary
	}

	summary += "Status: Success\n"
	summary += fmt.Sprintf("Packets: %d sent, %d received, %.1f%% loss\n",
		result.PacketsSent, result.PacketsReceived, result.PacketLoss)

	if len(result.RTTs) > 0 {
		summary += fmt.Sprintf("Round-trip time: min=%v, avg=%v, max=%v\n",
			result.MinRTT, result.AvgRTT, result.MaxRTT)

		summary += "Individual RTTs: "
		for i, rtt := range result.RTTs {
			if i > 0 {
				summary += ", "
			}
			summary += fmt.Sprintf("%v", rtt)
		}
		summary += "\n"
	}

	return summary
}

// FormatTracerouteSummary returns a human-readable summary of a traceroute result
func (s *Service) FormatTracerouteSummary(result *TracerouteResult) string {
	if result == nil {
		return "No traceroute results available"
	}

	summary := fmt.Sprintf("TRACEROUTE results for %s (%s):\n", result.Target, result.ResolvedIP)

	if result.Error != "" {
		summary += fmt.Sprintf("Error: %s\n", result.Error)
		return summary
	}

	if !result.TargetReached {
		summary += "Status: Target not reached\n\n"
	} else {
		summary += "Status: Target reached\n\n"
	}

	summary += "Hop  IP Address       Hostname                 RTT\n"
	summary += "---- --------------- ------------------------- -------\n"

	for _, hop := range result.Hops {
		hostname := hop.Hostname
		if hostname == "" {
			hostname = "*"
		}

		if hop.Error != "" {
			summary += fmt.Sprintf("%3d  %-15s %-25s %s\n", hop.Number, hop.IP, hostname, hop.Error)
		} else {
			summary += fmt.Sprintf("%3d  %-15s %-25s %v\n", hop.Number, hop.IP, hostname, hop.RTT)
		}
	}

	return summary
}

// FormatWHOISSummary returns a human-readable summary of a WHOIS result
func (s *Service) FormatWHOISSummary(result *WHOISResult) string {
	if result == nil {
		return "No WHOIS results available"
	}

	summary := fmt.Sprintf("WHOIS results for %s:\n\n", result.Target)

	if result.Error != "" {
		summary += fmt.Sprintf("Error: %s\n", result.Error)
		return summary
	}

	if result.Registrar != "" {
		summary += fmt.Sprintf("Registrar: %s\n", result.Registrar)
	}

	if result.CreatedDate != "" {
		summary += fmt.Sprintf("Created: %s\n", result.CreatedDate)
	}

	if result.ExpirationDate != "" {
		summary += fmt.Sprintf("Expires: %s\n", result.ExpirationDate)
	}

	if len(result.NameServers) > 0 {
		summary += "Name Servers:\n"
		for _, ns := range result.NameServers {
			summary += fmt.Sprintf("  %s\n", ns)
		}
	}

	summary += "\nRaw WHOIS Data:\n"
	summary += result.RawData

	return summary
}

// FormatNetworkToolSummary returns a human-readable summary of a generic network tool result
func (s *Service) FormatNetworkToolSummary(result *NetworkToolResult) string {
	if result == nil {
		return "No network tool results available"
	}

	switch result.ToolType {
	case ToolTypePing:
		return s.FormatPingSummary(result.PingResult)
	case ToolTypeTraceroute:
		return s.FormatTracerouteSummary(result.TracerouteResult)
	case ToolTypeWHOIS:
		return s.FormatWHOISSummary(result.WHOISResult)
	default:
		return fmt.Sprintf("Unknown tool type: %s", result.ToolType)
	}
}
