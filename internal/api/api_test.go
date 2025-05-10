package api_test

import (
	"bytes"
	"encoding/json"
	"mxclone/domain/dns"
	"mxclone/internal/api"
	"mxclone/internal/api/models"
	"mxclone/pkg/logging"
	"net/http"
	"net/http/httptest"
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

func (m *MockDNSService) LookupAll(ctx interface{}, domain string) (*dns.DNSResult, error) {
	m.lookupAllCalled = true
	m.target = domain
	return m.result, m.err
}

func (m *MockDNSService) Lookup(ctx interface{}, domain string, recordType dns.RecordType) (*dns.DNSResult, error) {
	m.lookupCalled = true
	m.target = domain
	m.recordType = recordType
	return m.result, m.err
}

// MockDNSBLService is a mock implementation of input.DNSBLPort
type MockDNSBLService struct {
	// Add mock fields as needed for testing
}

func (m *MockDNSBLService) CheckSingleBlacklist(ctx interface{}, ip string, zone string, timeout time.Duration) (interface{}, error) {
	return nil, nil
}

func (m *MockDNSBLService) CheckMultipleBlacklists(ctx interface{}, ip string, zones []string, timeout time.Duration) interface{} {
	return nil
}

func (m *MockDNSBLService) GetBlacklistSummary(result interface{}) string {
	return ""
}

func (m *MockDNSBLService) CheckDNSBLHealth(ctx interface{}, zone string, timeout time.Duration) (bool, error) {
	return true, nil
}

func (m *MockDNSBLService) CheckMultipleDNSBLHealth(ctx interface{}, zones []string, timeout time.Duration) map[string]bool {
	return nil
}

// MockSMTPService is a mock implementation of input.SMTPPort
type MockSMTPService struct {
	// Add mock fields as needed
}

func (m *MockSMTPService) CheckSMTP(ctx interface{}, server string, port int, domain string, timeout time.Duration) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{"status": "smtp check ok"}, nil
}


	mockDNSBLService := &MockDNSBLService{}
	mockSMTPService := &MockSMTPService{}
	mockEmailAuthService := &MockEmailAuthService{}
	mockNetworkToolsService := &MockNetworkToolsService{}

	// Create a logger
	logger := logging.NewLogger("test")

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

func (m *MockEmailAuthService) CheckDMARC(ctx interface{}, domain string) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{"dmarc_record": "v=DMARC1; p=none;"}, nil
}

// MockNetworkToolsService is a mock implementation of input.NetworkToolsPort
type MockNetworkToolsService struct {
	// Add mock fields as needed
}

func (m *MockNetworkToolsService) Ping(ctx interface{}, host string, count int, timeout time.Duration) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{"ping_status": "ok", "rtt_avg": "10ms"}, nil
}

func (m *MockNetworkToolsService) Traceroute(ctx interface{}, host string, maxHops int, timeout time.Duration) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{"traceroute_hops": []string{"hop1", "hop2"}}, nil
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

	// Create a logger
	logger := logging.NewLogger("test")

	// Create API server
	server := api.NewServer(
		mockDNSService,
		mockDNSBLService,
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

	// Run test cases
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
