package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mxclone/domain/dns"
	"mxclone/domain/dnsbl"
	"mxclone/domain/emailauth"
	"mxclone/domain/networktools"
	"mxclone/domain/smtp"
	"mxclone/internal/api"
	"mxclone/internal/api/models"
	"mxclone/pkg/logging"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// MockDNSService is a mock implementation of input.DNSPort
type MockDNSService struct {
	lookupAllCalled bool
	lookupCalled    bool
	target          string
	recordType      dns.RecordType
	result          *dns.DNSResult
	err             error
}

func (m *MockDNSService) LookupAll(ctx context.Context, domain string) (*dns.DNSResult, error) {
	m.lookupAllCalled = true
	m.target = domain
	return m.result, m.err
}

func (m *MockDNSService) Lookup(ctx context.Context, domain string, recordType dns.RecordType) (*dns.DNSResult, error) {
	m.lookupCalled = true
	m.target = domain
	m.recordType = recordType
	return m.result, m.err
}

// MockDNSBLService is a mock implementation of input.DNSBLPort
type MockDNSBLService struct {
	// Add mock fields as needed for testing
}

func (m *MockDNSBLService) CheckSingleBlacklist(ctx context.Context, ip string, zone string, timeout time.Duration) (*dnsbl.BlacklistResult, error) {
	return nil, nil
}

func (m *MockDNSBLService) CheckMultipleBlacklists(ctx context.Context, ip string, zones []string, timeout time.Duration) (*dnsbl.BlacklistResult, error) {
	return nil, nil
}

func (m *MockDNSBLService) GetBlacklistSummary(result *dnsbl.BlacklistResult) string {
	return ""
}

func (m *MockDNSBLService) CheckDNSBLHealth(ctx context.Context, zone string, timeout time.Duration) (bool, error) {
	return true, nil
}

func (m *MockDNSBLService) CheckMultipleDNSBLHealth(ctx context.Context, zones []string, timeout time.Duration) map[string]bool {
	return nil
}

// MockSMTPService is a mock implementation of input.SMTPPort
type MockSMTPService struct {
	// Add mock fields as needed
}

func (m *MockSMTPService) CheckSMTP(ctx context.Context, domain string, timeout time.Duration) (*smtp.SMTPResult, error) {
	return &smtp.SMTPResult{Domain: domain, MXRecords: []string{"mx1.example.com"}, ConnectionResults: map[string]*smtp.ConnectionResult{}, Banner: "smtp check ok"}, nil
}

func (m *MockSMTPService) TestSMTPConnection(ctx context.Context, server string, port int, timeout time.Duration) (*smtp.ConnectionResult, error) {
	return &smtp.ConnectionResult{Server: server, Connected: true}, nil
}

func (m *MockSMTPService) GetSMTPSummary(result *smtp.SMTPResult) string {
	return "SMTP check summary"
}

// Add missing MockEmailAuthService struct
type MockEmailAuthService struct {
	// Add mock fields as needed
}

func (m *MockEmailAuthService) CheckSPF(ctx context.Context, domain string, timeout time.Duration) (*emailauth.SPFResult, error) {
	return &emailauth.SPFResult{Record: "v=spf1 include:_spf.example.com ~all"}, nil
}

func (m *MockEmailAuthService) CheckDKIM(ctx context.Context, domain string, selectors []string, timeout time.Duration) (*emailauth.DKIMResult, error) {
	return &emailauth.DKIMResult{HasRecords: true, Records: map[string]string{"selector1": "dkimrecord"}}, nil
}

func (m *MockEmailAuthService) CheckDMARC(ctx context.Context, domain string, timeout time.Duration) (*emailauth.DMARCResult, error) {
	return &emailauth.DMARCResult{Record: "v=DMARC1; p=none;"}, nil
}

func (m *MockEmailAuthService) CheckAll(ctx context.Context, domain string, dkimSelectors []string, timeout time.Duration) (*emailauth.AuthResult, error) {
	return &emailauth.AuthResult{Domain: domain}, nil
}

func (m *MockEmailAuthService) GetAuthSummary(result *emailauth.AuthResult) string {
	return "Auth summary"
}

// MockNetworkToolsService is a mock implementation of input.NetworkToolsPort
type MockNetworkToolsService struct {
	// Add mock fields as needed
}

func (m *MockNetworkToolsService) ExecutePing(ctx context.Context, target string, count int, timeout time.Duration) (*networktools.PingResult, error) {
	return &networktools.PingResult{Target: target, Success: true, RTTs: []time.Duration{10 * time.Millisecond}}, nil
}

func (m *MockNetworkToolsService) ExecuteTraceroute(ctx context.Context, target string, maxHops int, timeout time.Duration) (*networktools.TracerouteResult, error) {
	return &networktools.TracerouteResult{Target: target, Hops: []networktools.TracerouteHop{{Number: 1, IP: "1.1.1.1", RTT: 10 * time.Millisecond}}}, nil
}

func (m *MockNetworkToolsService) ExecuteWHOIS(ctx context.Context, target string, timeout time.Duration) (*networktools.WHOISResult, error) {
	return &networktools.WHOISResult{Target: target, RawData: "whois data"}, nil
}

func (m *MockNetworkToolsService) ExecuteNetworkTool(ctx context.Context, toolType networktools.ToolType, target string, options map[string]interface{}) (*networktools.NetworkToolResult, error) {
	return &networktools.NetworkToolResult{ToolType: toolType}, nil
}

func (m *MockNetworkToolsService) WrapResult(toolType networktools.ToolType, pingResult *networktools.PingResult, tracerouteResult *networktools.TracerouteResult, whoisResult *networktools.WHOISResult, err error) *networktools.NetworkToolResult {
	return &networktools.NetworkToolResult{ToolType: toolType, PingResult: pingResult, TracerouteResult: tracerouteResult, WHOISResult: whoisResult}
}

func (m *MockNetworkToolsService) FormatToolResult(result *networktools.NetworkToolResult) string {
	return "Network tool result summary"
}

func (m *MockNetworkToolsService) ResolveDomain(ctx context.Context, domain string) (string, error) {
	return "127.0.0.1", nil
}

func (m *MockNetworkToolsService) TracerouteHop(ctx context.Context, target string, ttl int, timeout time.Duration) (networktools.TracerouteHop, bool, error) {
	return networktools.TracerouteHop{Number: ttl, IP: "1.1.1.1", RTT: 10 * time.Millisecond}, true, nil
}

func TestDNSHandler(t *testing.T) {
	// Create mock services
	mockDNSService := &MockDNSService{
		result: &dns.DNSResult{
			Lookups: map[string][]string{
				string(dns.TypeA):  {"192.0.2.1"},
				string(dns.TypeMX): {"10 mail.example.com."},
			},
		},
	}
	mockDNSBLService := &MockDNSBLService{}
	mockSMTPService := &MockSMTPService{}
	mockEmailAuthService := &MockEmailAuthService{}
	mockNetworkToolsService := &MockNetworkToolsService{}

	// Create a logger
	logger := logging.NewLogger("test", logging.LevelDebug, os.Stderr)

	// Create API server
	server := api.NewServer(
		mockDNSService,
		mockDNSBLService,
		mockSMTPService,
		mockEmailAuthService,
		mockNetworkToolsService,
		logger,
	)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route requests to the server
		if r.URL.Path == "/api/v1/dns" {
			// Strip version prefix for v1 router
			r.URL.Path = "/dns"
			server.ServeHTTP(w, r)
		} else {
			server.ServeHTTP(w, r)
		}
	}))
	defer ts.Close()

	// Test cases
	testCases := []struct {
		name           string
		endpoint       string
		method         string
		body           interface{}
		expectedStatus int
		validateResult func(t *testing.T, body []byte)
	}{
		{
			name:     "DNS Lookup - Valid",
			endpoint: "/api/v1/dns",
			method:   "POST",
			body: models.CheckRequest{
				Target: "example.com",
			},
			expectedStatus: http.StatusOK,
			validateResult: func(t *testing.T, body []byte) {
				var response models.DNSResponse
				err := json.Unmarshal(body, &response)
				if err != nil {
					t.Errorf("Error unmarshaling response: %v", err)
					return // Stop further processing if unmarshaling fails
				}

				// Check that records exist
				// response.Records is of type map[string][]string as per models.DNSResponse
				if response.Records == nil {
					t.Errorf("Expected records map, but it was nil")
				} else {
					if _, ok := response.Records["A"]; !ok {
						t.Errorf("Expected A records, got none")
					}
					if _, ok := response.Records["MX"]; !ok {
						t.Errorf("Expected MX records, got none")
					}
				}

				// Check that timing is reported
				if response.Timing == "" {
					t.Errorf("Expected timing to be reported")
				}
			},
		},
		{
			name:     "DNS Lookup - Invalid Domain",
			endpoint: "/api/v1/dns",
			method:   "POST",
			body: models.CheckRequest{
				Target: "invalid domain with spaces",
			},
			expectedStatus: http.StatusBadRequest,
			validateResult: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				if err != nil {
					t.Errorf("Error unmarshaling response: %v", err)
				}

				// Check that we got a validation error
				if _, ok := response["validations"]; !ok {
					t.Errorf("Expected validation errors, got none")
				}
			},
		},
	}

	// Run test
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal request body
			jsonBody, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatalf("Error marshaling request body: %v", err)
			}

			// Create request
			req, err := http.NewRequest(tc.method, ts.URL+tc.endpoint, bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Read response body
			body := new(bytes.Buffer)
			body.ReadFrom(resp.Body)

			// Validate result
			if tc.validateResult != nil {
				tc.validateResult(t, body.Bytes())
			}
		})
	}
}
