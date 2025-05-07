package models

import (
	"mxclone/domain/dns"
	"mxclone/domain/dnsbl"
	"mxclone/domain/emailauth"
	"mxclone/domain/networktools"
	"mxclone/domain/smtp"
	"strings"
)

// APIError represents an API error response
type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// CheckRequest represents a generic diagnostic request
type CheckRequest struct {
	Target string `json:"target"`
	Option string `json:"option,omitempty"` // Optional parameter for specific checks
}

// DNSResponse wraps the domain DNS result for API responses
type DNSResponse struct {
	Records map[string][]string `json:"records"`
	Timing  string              `json:"timing,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// FromDNSResult converts a domain DNS result to an API response
func FromDNSResult(result *dns.DNSResult) *DNSResponse {
	if result == nil {
		return &DNSResponse{
			Error: "no result available",
		}
	}

	return &DNSResponse{
		Records: result.Lookups,
		Error:   result.Error,
	}
}

// BlacklistResponse wraps the domain blacklist result for API responses
type BlacklistResponse struct {
	IP       string            `json:"ip"`
	ListedOn map[string]string `json:"listedOn"`
	Error    string            `json:"error,omitempty"`
}

// FromBlacklistResult converts a domain blacklist result to an API response
func FromBlacklistResult(result *dnsbl.BlacklistResult) *BlacklistResponse {
	if result == nil {
		return &BlacklistResponse{
			Error: "no result available",
		}
	}

	return &BlacklistResponse{
		IP:       result.CheckedIP,
		ListedOn: result.ListedOn,
		Error:    result.CheckError,
	}
}

// SMTPResponse wraps the domain SMTP result for API responses
type SMTPResponse struct {
	Connected        bool   `json:"connected"`
	SupportsStartTLS bool   `json:"supportsStartTLS"`
	Error            string `json:"error,omitempty"`
}

// FromSMTPResult converts a domain SMTP result to an API response
func FromSMTPResult(result *smtp.SMTPResult) *SMTPResponse {
	if result == nil {
		return &SMTPResponse{
			Error: "no result available",
		}
	}

	// For now, we're creating a simple response with the information we have
	// This can be expanded later to match the full domain model
	return &SMTPResponse{
		Error: result.Error,
	}
}

// SMTPConnectionResponse represents the result of an SMTP connection check
type SMTPConnectionResponse struct {
	Host             string   `json:"host"`
	Port             int      `json:"port"`
	Connected        bool     `json:"connected"`
	Latency          string   `json:"latency,omitempty"`
	SupportsStartTLS bool     `json:"supportsStartTLS"`
	AuthMethods      []string `json:"authMethods,omitempty"`
	Banner           string   `json:"banner,omitempty"`
	Error            string   `json:"error,omitempty"`
}

// FromSMTPConnectionResult converts a domain SMTP connection result to an API response
func FromSMTPConnectionResult(result *smtp.ConnectionResult) *SMTPConnectionResponse {
	if result == nil {
		return &SMTPConnectionResponse{
			Error: "no result available",
		}
	}

	response := &SMTPConnectionResponse{
		Host:             result.Server,
		Connected:        result.Connected,
		SupportsStartTLS: result.SupportsStartTLS,
		AuthMethods:      result.AuthMethods,
		Banner:           result.Banner,
	}

	if result.Latency > 0 {
		response.Latency = result.Latency.String()
	}

	if result.Error != "" {
		response.Error = result.Error
	}

	return response
}

// SMTPStartTLSResponse represents the result of an SMTP STARTTLS check
type SMTPStartTLSResponse struct {
	Host             string                             `json:"host"`
	MXRecords        []string                           `json:"mxRecords,omitempty"`
	ConnectionStatus map[string]*SMTPConnectionResponse `json:"connectionStatus,omitempty"`
	Error            string                             `json:"error,omitempty"`
}

// FromSMTPStartTLSResult extracts STARTTLS information from a comprehensive SMTP check
func FromSMTPStartTLSResult(result *smtp.SMTPResult) *SMTPStartTLSResponse {
	if result == nil {
		return &SMTPStartTLSResponse{
			Error: "no result available",
		}
	}

	response := &SMTPStartTLSResponse{
		Host:      result.Domain,
		MXRecords: result.MXRecords,
	}

	if len(result.ConnectionResults) > 0 {
		connectionStatus := make(map[string]*SMTPConnectionResponse)
		for server, connResult := range result.ConnectionResults {
			connectionStatus[server] = FromSMTPConnectionResult(connResult)
		}
		response.ConnectionStatus = connectionStatus
	}

	if result.Error != "" {
		response.Error = result.Error
	}

	return response
}

// SMTPRelayTestRequest represents a request to test an SMTP server for open relay
type SMTPRelayTestRequest struct {
	Host           string `json:"host"`
	Port           int    `json:"port,omitempty"` // Default to 25 if not specified
	FromAddress    string `json:"fromAddress"`
	ToAddress      string `json:"toAddress"`
	Timeout        int    `json:"timeout,omitempty"` // In seconds, default to 10
	Authentication bool   `json:"authentication,omitempty"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
}

// SMTPRelayTestResponse represents the result of an SMTP open relay test
type SMTPRelayTestResponse struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	IsOpenRelay  bool   `json:"isOpenRelay"`
	AuthRequired bool   `json:"authRequired"`
	ResponseCode int    `json:"responseCode,omitempty"`
	ResponseText string `json:"responseText,omitempty"`
	TestDetails  string `json:"testDetails,omitempty"`
	Error        string `json:"error,omitempty"`
}

// FromSMTPRelayTestResult extracts relay test information from the SMTP check
func FromSMTPRelayTestResult(result *smtp.SMTPResult) *SMTPRelayTestResponse {
	if result == nil {
		return &SMTPRelayTestResponse{
			Error: "no result available",
		}
	}

	// For now, we assume the relay test result based on available data
	// In a real implementation, SMTP service would have a dedicated relay test method
	response := &SMTPRelayTestResponse{
		Host: result.Domain,
		Port: 25, // Default value, would get from actual test
	}

	// Simple determination - not a real relay test
	// This should be replaced with actual relay test logic
	isOpenRelay := false
	authRequired := true
	testDetails := "Relay test simulated - not actual test results"

	if len(result.ConnectionResults) > 0 {
		for _, connResult := range result.ConnectionResults {
			// Just use the first result for this demo
			response.ResponseText = connResult.Banner

			// Simple check for now - in reality, need to try to relay mail
			if len(connResult.AuthMethods) == 0 && connResult.Connected {
				authRequired = false
				// Only mark as potential open relay if no auth methods AND connected
				if connResult.Connected {
					isOpenRelay = true
					testDetails += " - Potential security issue: Server connected with no authentication methods required"
				}
			}

			break
		}
	}

	response.IsOpenRelay = isOpenRelay
	response.AuthRequired = authRequired
	response.TestDetails = testDetails

	if result.Error != "" {
		response.Error = result.Error
	}

	return response
}

// EmailAuthResponse wraps the domain email authentication result for API responses
type EmailAuthResponse struct {
	Domain    string `json:"domain"`
	SPF       bool   `json:"spf"`
	DKIM      bool   `json:"dkim"`
	DMARC     bool   `json:"dmarc"`
	AllPassed bool   `json:"allPassed"`
	Error     string `json:"error,omitempty"`
}

// SPFResponse represents the result of an SPF record check
type SPFResponse struct {
	Domain     string   `json:"domain"`
	HasRecord  bool     `json:"hasRecord"`
	Record     string   `json:"record,omitempty"`
	IsValid    bool     `json:"isValid"`
	Mechanisms []string `json:"mechanisms,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// FromSPFResult converts a domain SPF result to an API response
func FromSPFResult(result *emailauth.SPFResult) *SPFResponse {
	if result == nil {
		return &SPFResponse{
			Error: "no result available",
		}
	}

	return &SPFResponse{
		HasRecord:  result.HasRecord,
		Record:     result.Record,
		IsValid:    result.IsValid,
		Mechanisms: result.Mechanisms,
		Error:      result.Error,
	}
}

// DKIMResponse represents the result of a DKIM record check
type DKIMResponse struct {
	Domain     string            `json:"domain"`
	Selector   string            `json:"selector"`
	HasRecords bool              `json:"hasRecords"`
	Records    map[string]string `json:"records,omitempty"`
	IsValid    bool              `json:"isValid"`
	Error      string            `json:"error,omitempty"`
}

// CombinedDKIMResponse represents results from multiple DKIM selectors
type CombinedDKIMResponse struct {
	Domain    string         `json:"domain"`
	Results   []DKIMResponse `json:"results"`
	IsValid   bool           `json:"isValid"`
	Selectors []string       `json:"selectors"`
	Error     string         `json:"error,omitempty"`
}

// FromDKIMResult converts a domain DKIM result to an API response
func FromDKIMResult(result *emailauth.DKIMResult) *DKIMResponse {
	if result == nil {
		return &DKIMResponse{
			Error: "no result available",
		}
	}

	// Extract the selector from the first record key, if available
	selector := ""
	for key := range result.Records {
		parts := strings.Split(key, ".")
		if len(parts) > 0 {
			selector = parts[0]
			break
		}
	}

	return &DKIMResponse{
		Selector:   selector,
		HasRecords: result.HasRecords,
		Records:    result.Records,
		IsValid:    result.IsValid,
		Error:      result.Error,
	}
}

// DMARCResponse represents the result of a DMARC record check
type DMARCResponse struct {
	Domain          string `json:"domain"`
	HasRecord       bool   `json:"hasRecord"`
	Record          string `json:"record,omitempty"`
	IsValid         bool   `json:"isValid"`
	Policy          string `json:"policy,omitempty"`
	SubdomainPolicy string `json:"subdomainPolicy,omitempty"`
	Percentage      int    `json:"percentage,omitempty"`
	Error           string `json:"error,omitempty"`
}

// FromDMARCResult converts a domain DMARC result to an API response
func FromDMARCResult(result *emailauth.DMARCResult) *DMARCResponse {
	if result == nil {
		return &DMARCResponse{
			Error: "no result available",
		}
	}

	return &DMARCResponse{
		HasRecord:       result.HasRecord,
		Record:          result.Record,
		IsValid:         result.IsValid,
		Policy:          result.Policy,
		SubdomainPolicy: result.SubdomainPolicy,
		Percentage:      result.Percentage,
		Error:           result.Error,
	}
}

// NetworkToolResponse wraps the domain network tool result for API responses
type NetworkToolResponse struct {
	Target    string `json:"target"`
	Tool      string `json:"tool"`
	RawOutput string `json:"rawOutput,omitempty"`
	Error     string `json:"error,omitempty"`
}

// PingResponse represents the result of a ping operation
type PingResponse struct {
	Target          string   `json:"target"`
	ResolvedIP      string   `json:"resolvedIP,omitempty"`
	Success         bool     `json:"success"`
	RTTs            []string `json:"rtts,omitempty"`
	AvgRTT          string   `json:"avgRTT,omitempty"`
	MinRTT          string   `json:"minRTT,omitempty"`
	MaxRTT          string   `json:"maxRTT,omitempty"`
	PacketsSent     int      `json:"packetsSent"`
	PacketsReceived int      `json:"packetsReceived"`
	PacketLoss      float64  `json:"packetLoss"`
	Error           string   `json:"error,omitempty"`
	RawOutput       string   `json:"rawOutput,omitempty"`
}

// TracerouteHopResponse represents a single hop in a traceroute path
type TracerouteHopResponse struct {
	Number   int    `json:"number"`
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	RTT      string `json:"rtt,omitempty"`
	Error    string `json:"error,omitempty"`
}

// TracerouteResponse represents the result of a traceroute operation
type TracerouteResponse struct {
	Target        string                  `json:"target"`
	ResolvedIP    string                  `json:"resolvedIP,omitempty"`
	Hops          []TracerouteHopResponse `json:"hops,omitempty"`
	TargetReached bool                    `json:"targetReached"`
	Error         string                  `json:"error,omitempty"`
	RawOutput     string                  `json:"rawOutput,omitempty"`
}

// WHOISResponse represents the result of a WHOIS query
type WHOISResponse struct {
	Target         string   `json:"target"`
	Registrar      string   `json:"registrar,omitempty"`
	CreatedDate    string   `json:"createdDate,omitempty"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	NameServers    []string `json:"nameServers,omitempty"`
	RawData        string   `json:"rawData,omitempty"`
	Error          string   `json:"error,omitempty"`
}

// NetworkToolResult represents the result of a network tool operation
type NetworkToolResult struct {
	Tool             string              `json:"tool"`
	Target           string              `json:"target"`
	PingResult       *PingResponse       `json:"pingResult,omitempty"`
	TracerouteResult *TracerouteResponse `json:"tracerouteResult,omitempty"`
	WhoisResult      *WHOISResponse      `json:"whoisResult,omitempty"`
	Error            string              `json:"error,omitempty"`
}

// FromNetworkToolResult converts a domain NetworkToolResult to an API response
func FromNetworkToolResult(result *networktools.NetworkToolResult) *NetworkToolResult {
	if result == nil {
		return &NetworkToolResult{
			Error: "no result available",
		}
	}

	response := &NetworkToolResult{
		Tool:  string(result.ToolType),
		Error: result.Error,
	}

	// Extract the appropriate result based on the tool type
	switch result.ToolType {
	case networktools.ToolTypePing:
		if result.PingResult != nil {
			pingResponse := &PingResponse{
				Target:          result.PingResult.Target,
				ResolvedIP:      result.PingResult.ResolvedIP,
				Success:         result.PingResult.Success,
				PacketsSent:     result.PingResult.PacketsSent,
				PacketsReceived: result.PingResult.PacketsReceived,
				PacketLoss:      result.PingResult.PacketLoss,
				Error:           result.PingResult.Error,
			}

			// Convert time.Duration to strings for the response
			if result.PingResult.AvgRTT > 0 {
				pingResponse.AvgRTT = result.PingResult.AvgRTT.String()
			}
			if result.PingResult.MinRTT > 0 {
				pingResponse.MinRTT = result.PingResult.MinRTT.String()
			}
			if result.PingResult.MaxRTT > 0 {
				pingResponse.MaxRTT = result.PingResult.MaxRTT.String()
			}

			// Convert RTTs to strings
			if len(result.PingResult.RTTs) > 0 {
				pingResponse.RTTs = make([]string, len(result.PingResult.RTTs))
				for i, rtt := range result.PingResult.RTTs {
					pingResponse.RTTs[i] = rtt.String()
				}
			}

			response.PingResult = pingResponse
			response.Target = result.PingResult.Target
		}

	case networktools.ToolTypeTraceroute:
		if result.TracerouteResult != nil {
			tracerouteResponse := &TracerouteResponse{
				Target:        result.TracerouteResult.Target,
				ResolvedIP:    result.TracerouteResult.ResolvedIP,
				TargetReached: result.TracerouteResult.TargetReached,
				Error:         result.TracerouteResult.Error,
			}

			// Convert hops
			if len(result.TracerouteResult.Hops) > 0 {
				tracerouteResponse.Hops = make([]TracerouteHopResponse, len(result.TracerouteResult.Hops))
				for i, hop := range result.TracerouteResult.Hops {
					hopResponse := TracerouteHopResponse{
						Number:   hop.Number,
						IP:       hop.IP,
						Hostname: hop.Hostname,
						Error:    hop.Error,
					}

					if hop.RTT > 0 {
						hopResponse.RTT = hop.RTT.String()
					}

					tracerouteResponse.Hops[i] = hopResponse
				}
			}

			response.TracerouteResult = tracerouteResponse
			response.Target = result.TracerouteResult.Target
		}

	case networktools.ToolTypeWHOIS:
		if result.WHOISResult != nil {
			whoisResponse := &WHOISResponse{
				Target:         result.WHOISResult.Target,
				Registrar:      result.WHOISResult.Registrar,
				CreatedDate:    result.WHOISResult.CreatedDate,
				ExpirationDate: result.WHOISResult.ExpirationDate,
				NameServers:    result.WHOISResult.NameServers,
				RawData:        result.WHOISResult.RawData,
				Error:          result.WHOISResult.Error,
			}

			response.WhoisResult = whoisResponse
			response.Target = result.WHOISResult.Target
		}
	}

	return response
}
