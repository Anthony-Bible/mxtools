// Package networktools provides auxiliary network diagnostic tools.
package networktools

import (
	"strings"
	"testing"
)

// TestParseWhoisFields tests the ParseWhoisFields function
func TestParseWhoisFields(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected map[string]string
	}{
		{
			name: "Simple response",
			response: `
Domain Name: EXAMPLE.COM
Registrar: Example Registrar, LLC
Creation Date: 1995-08-14T04:00:00Z
Registry Expiry Date: 2021-08-13T04:00:00Z
Registrar WHOIS Server: whois.example.com
Registrar URL: http://www.example.com
Updated Date: 2019-08-14T04:00:00Z
Domain Status: clientDeleteProhibited
Domain Status: clientTransferProhibited
Domain Status: clientUpdateProhibited
Registry Registrant ID: 
Registrant Name: EXAMPLE REGISTRANT
Registrant Organization: EXAMPLE ORGANIZATION
Name Server: NS1.EXAMPLE.COM
Name Server: NS2.EXAMPLE.COM
DNSSEC: signedDelegation
`,
			expected: map[string]string{
				"Domain Name":           "EXAMPLE.COM",
				"Registrar":             "Example Registrar, LLC",
				"Creation Date":         "1995-08-14T04:00:00Z",
				"Registry Expiry Date":  "2021-08-13T04:00:00Z",
				"Registrar WHOIS Server": "whois.example.com",
				"Registrar URL":         "http://www.example.com",
				"Updated Date":          "2019-08-14T04:00:00Z",
				"Domain Status":         "clientDeleteProhibited",
				"Registry Registrant ID": "",
				"Registrant Name":       "EXAMPLE REGISTRANT",
				"Registrant Organization": "EXAMPLE ORGANIZATION",
				"Name Server":           "NS1.EXAMPLE.COM",
				"DNSSEC":                "signedDelegation",
			},
		},
		{
			name: "Response with comments",
			response: `
% This is a comment
# This is another comment
Domain Name: EXAMPLE.COM
Registrar: Example Registrar, LLC
`,
			expected: map[string]string{
				"Domain Name": "EXAMPLE.COM",
				"Registrar":   "Example Registrar, LLC",
			},
		},
		{
			name: "Response with empty lines",
			response: `

Domain Name: EXAMPLE.COM

Registrar: Example Registrar, LLC

`,
			expected: map[string]string{
				"Domain Name": "EXAMPLE.COM",
				"Registrar":   "Example Registrar, LLC",
			},
		},
		{
			name:     "Empty response",
			response: "",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWhoisFields(tt.response)
			
			// Check if the number of fields matches
			if len(result) != len(tt.expected) {
				t.Errorf("ParseWhoisFields() returned %d fields, want %d", len(result), len(tt.expected))
			}
			
			// Check if each field matches
			for key, expectedValue := range tt.expected {
				if value, ok := result[key]; !ok {
					t.Errorf("ParseWhoisFields() missing field %s", key)
				} else if value != expectedValue {
					t.Errorf("ParseWhoisFields() field %s = %s, want %s", key, value, expectedValue)
				}
			}
		})
	}
}

// TestExtractCommonWhoisFields tests the ExtractCommonWhoisFields function
func TestExtractCommonWhoisFields(t *testing.T) {
	tests := []struct {
		name     string
		result   *WhoisResult
		expected map[string]string
	}{
		{
			name: "Common fields",
			result: &WhoisResult{
				Query: "example.com",
				Fields: map[string]string{
					"domain name":        "example.com",
					"registrar":          "Example Registrar, LLC",
					"creation date":      "1995-08-14T04:00:00Z",
					"expiration date":    "2021-08-13T04:00:00Z",
					"updated date":       "2019-08-14T04:00:00Z",
					"name server":        "NS1.EXAMPLE.COM",
					"status":             "clientDeleteProhibited",
					"dnssec":             "signedDelegation",
				},
			},
			expected: map[string]string{
				"Domain Name":     "example.com",
				"Registrar":       "Example Registrar, LLC",
				"Creation Date":   "1995-08-14T04:00:00Z",
				"Expiration Date": "2021-08-13T04:00:00Z",
				"Updated Date":    "2019-08-14T04:00:00Z",
				"Name Servers":    "NS1.EXAMPLE.COM",
				"Status":          "clientDeleteProhibited",
				"DNSSEC":          "signedDelegation",
			},
		},
		{
			name: "Field variations",
			result: &WhoisResult{
				Query: "example.com",
				Fields: map[string]string{
					"domain":             "example.com",
					"sponsor":            "Example Registrar, LLC",
					"created":            "1995-08-14T04:00:00Z",
					"registry expiry date": "2021-08-13T04:00:00Z",
					"last updated":       "2019-08-14T04:00:00Z",
					"nameserver":         "NS1.EXAMPLE.COM",
					"domain status":      "clientDeleteProhibited",
				},
			},
			expected: map[string]string{
				"Domain Name":     "example.com",
				"Registrar":       "Example Registrar, LLC",
				"Creation Date":   "1995-08-14T04:00:00Z",
				"Expiration Date": "2021-08-13T04:00:00Z",
				"Updated Date":    "2019-08-14T04:00:00Z",
				"Name Servers":    "NS1.EXAMPLE.COM",
				"Status":          "clientDeleteProhibited",
			},
		},
		{
			name: "No common fields",
			result: &WhoisResult{
				Query: "example.com",
				Fields: map[string]string{
					"unknown field": "value",
				},
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractCommonWhoisFields(tt.result)
			
			// Check if the number of fields matches
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractCommonWhoisFields() returned %d fields, want %d", len(result), len(tt.expected))
			}
			
			// Check if each field matches
			for key, expectedValue := range tt.expected {
				if value, ok := result[key]; !ok {
					t.Errorf("ExtractCommonWhoisFields() missing field %s", key)
				} else if value != expectedValue {
					t.Errorf("ExtractCommonWhoisFields() field %s = %s, want %s", key, value, expectedValue)
				}
			}
		})
	}
}

// TestFormatWhoisResult tests the FormatWhoisResult function
func TestFormatWhoisResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *WhoisResult
		contains []string
		notContains []string
	}{
		{
			name: "Successful WHOIS query",
			result: &WhoisResult{
				Query: "example.com",
				Fields: map[string]string{
					"domain name":   "example.com",
					"registrar":     "Example Registrar, LLC",
					"creation date": "1995-08-14T04:00:00Z",
				},
				RawResponse: "Domain Name: example.com\nRegistrar: Example Registrar, LLC\nCreation Date: 1995-08-14T04:00:00Z",
			},
			contains: []string{
				"WHOIS information for example.com",
				"Common Fields:",
				"Domain Name: example.com",
				"Registrar: Example Registrar, LLC",
				"Creation Date: 1995-08-14T04:00:00Z",
				"Raw Response:",
			},
			notContains: []string{
				"failed",
			},
		},
		{
			name: "Failed WHOIS query",
			result: &WhoisResult{
				Query: "example.com",
				Error: "Failed to connect to WHOIS server",
			},
			contains: []string{
				"WHOIS query for example.com failed: Failed to connect to WHOIS server",
			},
			notContains: []string{
				"Common Fields:",
				"Raw Response:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWhoisResult(tt.result)
			
			// Check that the output contains all expected substrings
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("FormatWhoisResult() = %v, should contain %v", result, s)
				}
			}
			
			// Check that the output does not contain any unwanted substrings
			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("FormatWhoisResult() = %v, should not contain %v", result, s)
				}
			}
		})
	}
}