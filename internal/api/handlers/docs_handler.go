package handlers

import (
	"encoding/json"
	"mxclone/pkg/logging"
	"net/http"
	"strings"
)

// APIEndpointDoc represents documentation for a specific API endpoint
type APIEndpointDoc struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Parameters  []APIParameterDoc `json:"parameters,omitempty"`
	Request     interface{}       `json:"request,omitempty"`
	Response    interface{}       `json:"response,omitempty"`
	Example     string            `json:"example,omitempty"`
}

// APIParameterDoc represents documentation for an API parameter
type APIParameterDoc struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // "path", "query", "header", "body"
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Example     interface{} `json:"example,omitempty"`
}

// APIGroupDoc represents documentation for a group of related API endpoints
type APIGroupDoc struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Endpoints   []APIEndpointDoc `json:"endpoints"`
}

// APIDocs represents the full API documentation
type APIDocs struct {
	Version     string        `json:"version"`
	Description string        `json:"description"`
	BaseURL     string        `json:"baseUrl"`
	Groups      []APIGroupDoc `json:"groups"`
}

// DocsHandler handles API documentation requests
type DocsHandler struct {
	logger *logging.Logger
	docs   *APIDocs
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler(logger *logging.Logger) *DocsHandler {
	handler := &DocsHandler{
		logger: logger,
		docs:   createDocs(),
	}
	return handler
}

// HandleDocs handles the API documentation endpoint
func (h *DocsHandler) HandleDocs(w http.ResponseWriter, r *http.Request) {
	// Check if a specific group is requested
	group := r.URL.Query().Get("group")
	if group != "" {
		h.handleGroupDocs(w, r, group)
		return
	}

	// Return all documentation if no specific group requested
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.docs)
}

// handleGroupDocs returns documentation for a specific group
func (h *DocsHandler) handleGroupDocs(w http.ResponseWriter, r *http.Request, group string) {
	group = strings.ToLower(group)

	// Find the requested group
	for _, g := range h.docs.Groups {
		if strings.ToLower(g.Name) == group {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(g)
			return
		}
	}

	// Group not found
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Documentation group not found",
	})
}

// createDocs creates the API documentation
func createDocs() *APIDocs {
	docs := &APIDocs{
		Version:     "v1",
		Description: "MXClone API provides DNS, blacklist, SMTP, email authentication, and network diagnostics",
		BaseURL:     "/api/v1",
		Groups:      createGroups(),
	}
	return docs
}

// createGroups creates documentation for all API groups
func createGroups() []APIGroupDoc {
	return []APIGroupDoc{
		createDNSGroup(),
		createBlacklistGroup(),
		createSMTPGroup(),
		createEmailAuthGroup(),
		createNetworkToolsGroup(),
		createHealthGroup(),
	}
}

// createDNSGroup creates documentation for DNS endpoints
func createDNSGroup() APIGroupDoc {
	return APIGroupDoc{
		Name:        "dns",
		Description: "DNS lookup operations",
		Endpoints: []APIEndpointDoc{
			{
				Path:        "/dns",
				Method:      "POST",
				Description: "Perform DNS lookups for multiple record types",
				Request: map[string]interface{}{
					"target": "example.com",
					"option": "mx",
				},
				Response: map[string]interface{}{
					"records": map[string]interface{}{
						"A":   []string{"192.0.2.1"},
						"MX":  []string{"10 mail.example.com."},
						"TXT": []string{"v=spf1 include:_spf.example.com ~all"},
					},
					"timing": "245.3ms",
				},
				Example: "curl -X POST http://localhost:8080/api/v1/dns -d '{\"target\":\"example.com\"}'",
			},
		},
	}
}

// createBlacklistGroup creates documentation for blacklist endpoints
func createBlacklistGroup() APIGroupDoc {
	return APIGroupDoc{
		Name:        "blacklist",
		Description: "IP and domain blacklist checking",
		Endpoints: []APIEndpointDoc{
			{
				Path:        "/blacklist",
				Method:      "POST",
				Description: "Check if an IP or domain is listed in DNS blacklists",
				Request: map[string]interface{}{
					"target": "192.0.2.1",
				},
				Response: map[string]interface{}{
					"ip": "192.0.2.1",
					"listedOn": map[string]string{
						"bl.spamcop.net":  "Listed - see details at spamcop.net",
						"dnsbl.sorbs.net": "Listed as suspicious source",
					},
				},
				Example: "curl -X POST http://localhost:8080/api/v1/blacklist -d '{\"target\":\"192.0.2.1\"}'",
			},
		},
	}
}

// createSMTPGroup creates documentation for SMTP endpoints
func createSMTPGroup() APIGroupDoc {
	return APIGroupDoc{
		Name:        "smtp",
		Description: "SMTP server diagnostics",
		Endpoints: []APIEndpointDoc{
			{
				Path:        "/smtp",
				Method:      "POST",
				Description: "Perform comprehensive SMTP server checks for a domain",
				Request: map[string]interface{}{
					"target": "example.com",
				},
				Response: map[string]interface{}{
					"connected":        true,
					"supportsStartTLS": true,
				},
				Example: "curl -X POST http://localhost:8080/api/v1/smtp -d '{\"target\":\"example.com\"}'",
			},
			{
				Path:        "/smtp/connect/{host}",
				Method:      "POST",
				Description: "Test basic SMTP connectivity to a server",
				Parameters: []APIParameterDoc{
					{
						Name:        "host",
						In:          "path",
						Description: "Domain or IP of SMTP server",
						Required:    true,
						Type:        "string",
						Example:     "smtp.example.com",
					},
				},
				Response: map[string]interface{}{
					"host":             "smtp.example.com",
					"port":             25,
					"connected":        true,
					"latency":          "45.3ms",
					"supportsStartTLS": true,
					"authMethods":      []string{"LOGIN", "PLAIN"},
					"banner":           "220 smtp.example.com ESMTP Service ready",
				},
				Example: "curl -X POST http://localhost:8080/api/v1/smtp/connect/smtp.example.com",
			},
			{
				Path:        "/smtp/starttls/{host}",
				Method:      "POST",
				Description: "Test if an SMTP server supports STARTTLS upgrade",
				Parameters: []APIParameterDoc{
					{
						Name:        "host",
						In:          "path",
						Description: "Domain or IP of SMTP server",
						Required:    true,
						Type:        "string",
						Example:     "smtp.example.com",
					},
				},
				Response: map[string]interface{}{
					"host":      "example.com",
					"mxRecords": []string{"smtp.example.com"},
					"connectionStatus": map[string]interface{}{
						"smtp.example.com": map[string]interface{}{
							"connected":        true,
							"supportsStartTLS": true,
						},
					},
				},
				Example: "curl -X POST http://localhost:8080/api/v1/smtp/starttls/example.com",
			},
			{
				Path:        "/smtp/relay-test",
				Method:      "POST",
				Description: "Test if an SMTP server is configured as an open relay",
				Request: map[string]interface{}{
					"host":           "smtp.example.com",
					"fromAddress":    "sender@example.com",
					"toAddress":      "recipient@example.net",
					"authentication": false,
				},
				Response: map[string]interface{}{
					"host":         "smtp.example.com",
					"port":         25,
					"isOpenRelay":  false,
					"authRequired": true,
					"responseCode": 550,
					"responseText": "Relay access denied",
					"testDetails":  "Server requires authentication before relaying mail",
				},
				Example: "curl -X POST http://localhost:8080/api/v1/smtp/relay-test -d '{\"host\":\"smtp.example.com\",\"fromAddress\":\"sender@example.com\",\"toAddress\":\"recipient@example.net\"}'",
			},
		},
	}
}

// createEmailAuthGroup creates documentation for email authentication endpoints
func createEmailAuthGroup() APIGroupDoc {
	return APIGroupDoc{
		Name:        "email-auth",
		Description: "Email authentication (SPF, DKIM, DMARC)",
		Endpoints: []APIEndpointDoc{
			{
				Path:        "/auth/spf/{domain}",
				Method:      "POST",
				Description: "Validates SPF record for a domain",
				Parameters: []APIParameterDoc{
					{
						Name:        "domain",
						In:          "path",
						Description: "Domain to check",
						Required:    true,
						Type:        "string",
						Example:     "example.com",
					},
				},
				Response: map[string]interface{}{
					"domain":     "example.com",
					"hasRecord":  true,
					"record":     "v=spf1 include:_spf.example.com ~all",
					"isValid":    true,
					"mechanisms": []string{"include:_spf.example.com", "~all"},
				},
				Example: "curl -X POST http://localhost:8080/api/v1/auth/spf/example.com",
			},
			{
				Path:        "/auth/dkim/{domain}/{selector}",
				Method:      "POST",
				Description: "Validates DKIM record for a domain and selector",
				Parameters: []APIParameterDoc{
					{
						Name:        "domain",
						In:          "path",
						Description: "Domain to check",
						Required:    true,
						Type:        "string",
						Example:     "example.com",
					},
					{
						Name:        "selector",
						In:          "path",
						Description: "DKIM selector",
						Required:    true,
						Type:        "string",
						Example:     "default",
					},
				},
				Response: map[string]interface{}{
					"domain":     "example.com",
					"selector":   "default",
					"hasRecords": true,
					"records": map[string]string{
						"default._domainkey.example.com": "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0...",
					},
					"isValid": true,
				},
				Example: "curl -X POST http://localhost:8080/api/v1/auth/dkim/example.com/default",
			},
			{
				Path:        "/auth/dmarc/{domain}",
				Method:      "POST",
				Description: "Validates DMARC record for a domain",
				Parameters: []APIParameterDoc{
					{
						Name:        "domain",
						In:          "path",
						Description: "Domain to check",
						Required:    true,
						Type:        "string",
						Example:     "example.com",
					},
				},
				Response: map[string]interface{}{
					"domain":          "example.com",
					"hasRecord":       true,
					"record":          "v=DMARC1; p=reject; sp=none; pct=100; rua=mailto:dmarc@example.com",
					"isValid":         true,
					"policy":          "reject",
					"subdomainPolicy": "none",
					"percentage":      100,
				},
				Example: "curl -X POST http://localhost:8080/api/v1/auth/dmarc/example.com",
			},
		},
	}
}

// createNetworkToolsGroup creates documentation for network tools endpoints
func createNetworkToolsGroup() APIGroupDoc {
	return APIGroupDoc{
		Name:        "network",
		Description: "Network diagnostic tools (ping, traceroute, WHOIS)",
		Endpoints: []APIEndpointDoc{
			{
				Path:        "/network/ping/{host}",
				Method:      "POST",
				Description: "Sends ICMP echo requests to a host",
				Parameters: []APIParameterDoc{
					{
						Name:        "host",
						In:          "path",
						Description: "Host to ping",
						Required:    true,
						Type:        "string",
						Example:     "example.com",
					},
					{
						Name:        "count",
						In:          "query",
						Description: "Number of ping packets to send",
						Required:    false,
						Type:        "integer",
						Example:     4,
					},
					{
						Name:        "timeout",
						In:          "query",
						Description: "Timeout in seconds",
						Required:    false,
						Type:        "integer",
						Example:     10,
					},
				},
				Response: map[string]interface{}{
					"target":          "example.com",
					"resolvedIP":      "192.0.2.1",
					"success":         true,
					"rtts":            []string{"24.5ms", "25.1ms", "26.3ms", "25.2ms"},
					"avgRTT":          "25.3ms",
					"minRTT":          "24.5ms",
					"maxRTT":          "26.3ms",
					"packetsSent":     4,
					"packetsReceived": 4,
					"packetLoss":      0,
				},
				Example: "curl -X POST http://localhost:8080/api/v1/network/ping/example.com?count=4",
			},
			{
				Path:        "/network/traceroute/{host}",
				Method:      "POST",
				Description: "Performs traceroute to a host",
				Parameters: []APIParameterDoc{
					{
						Name:        "host",
						In:          "path",
						Description: "Host to trace",
						Required:    true,
						Type:        "string",
						Example:     "example.com",
					},
					{
						Name:        "maxHops",
						In:          "query",
						Description: "Maximum number of hops",
						Required:    false,
						Type:        "integer",
						Example:     30,
					},
					{
						Name:        "timeout",
						In:          "query",
						Description: "Timeout in seconds",
						Required:    false,
						Type:        "integer",
						Example:     30,
					},
				},
				Response: map[string]interface{}{
					"target":     "example.com",
					"resolvedIP": "192.0.2.1",
					"hops": []map[string]interface{}{
						{
							"number":   1,
							"ip":       "192.168.1.1",
							"hostname": "router.local",
							"rtt":      "2.3ms",
						},
						{
							"number":   2,
							"ip":       "203.0.113.1",
							"hostname": "isp-router.example.net",
							"rtt":      "14.7ms",
						},
					},
					"targetReached": true,
				},
				Example: "curl -X POST http://localhost:8080/api/v1/network/traceroute/example.com?maxHops=30",
			},
			{
				Path:        "/network/whois/{domain}",
				Method:      "POST",
				Description: "Performs WHOIS lookup for a domain",
				Parameters: []APIParameterDoc{
					{
						Name:        "domain",
						In:          "path",
						Description: "Domain to lookup",
						Required:    true,
						Type:        "string",
						Example:     "example.com",
					},
				},
				Response: map[string]interface{}{
					"target":         "example.com",
					"registrar":      "Example Registrar, Inc.",
					"createdDate":    "1995-08-14T04:00:00Z",
					"expirationDate": "2024-08-13T04:00:00Z",
					"nameServers":    []string{"ns1.example.com", "ns2.example.com"},
				},
				Example: "curl -X POST http://localhost:8080/api/v1/network/whois/example.com",
			},
		},
	}
}

// createHealthGroup creates documentation for health endpoints
func createHealthGroup() APIGroupDoc {
	return APIGroupDoc{
		Name:        "health",
		Description: "System health and status",
		Endpoints: []APIEndpointDoc{
			{
				Path:        "/health",
				Method:      "GET",
				Description: "Check API health",
				Response: map[string]interface{}{
					"status":  "ok",
					"version": "v1",
				},
				Example: "curl http://localhost:8080/api/v1/health",
			},
		},
	}
}
