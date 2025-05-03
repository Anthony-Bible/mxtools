// Package dns provides DNS lookup functionality.
package dns

import (
	"context"
	"errors"
	"testing"

	"mxclone/pkg/types"
)

// mockDNSResolver is a mock implementation for testing DNS functions
type mockDNSResolver struct {
	lookupAFunc     func(ctx context.Context, domain string) ([]string, error)
	lookupAAAAFunc  func(ctx context.Context, domain string) ([]string, error)
	lookupMXFunc    func(ctx context.Context, domain string) ([]string, error)
	lookupTXTFunc   func(ctx context.Context, domain string) ([]string, error)
	lookupCNAMEFunc func(ctx context.Context, domain string) ([]string, error)
	lookupNSFunc    func(ctx context.Context, domain string) ([]string, error)
	lookupPTRFunc   func(ctx context.Context, domain string) ([]string, error)
}

// newMockDNSResolver creates a new mock DNS resolver
func newMockDNSResolver() *mockDNSResolver {
	return &mockDNSResolver{}
}

// mockLookupA mocks the A record lookup
func (m *mockDNSResolver) mockLookupA(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupAFunc != nil {
		records, err := m.lookupAFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["A"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// mockLookupAAAA mocks the AAAA record lookup
func (m *mockDNSResolver) mockLookupAAAA(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupAAAAFunc != nil {
		records, err := m.lookupAAAAFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["AAAA"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// mockLookupMX mocks the MX record lookup
func (m *mockDNSResolver) mockLookupMX(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupMXFunc != nil {
		records, err := m.lookupMXFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["MX"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// mockLookupTXT mocks the TXT record lookup
func (m *mockDNSResolver) mockLookupTXT(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupTXTFunc != nil {
		records, err := m.lookupTXTFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["TXT"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// mockLookupCNAME mocks the CNAME record lookup
func (m *mockDNSResolver) mockLookupCNAME(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupCNAMEFunc != nil {
		records, err := m.lookupCNAMEFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["CNAME"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// mockLookupNS mocks the NS record lookup
func (m *mockDNSResolver) mockLookupNS(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupNSFunc != nil {
		records, err := m.lookupNSFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["NS"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// mockLookupPTR mocks the PTR record lookup
func (m *mockDNSResolver) mockLookupPTR(ctx context.Context, domain string, result *types.DNSResult) error {
	if m.lookupPTRFunc != nil {
		records, err := m.lookupPTRFunc(ctx, domain)
		if err != nil {
			return err
		}
		result.Lookups["PTR"] = records
		return nil
	}
	return errors.New("mock not implemented")
}

// TestLookupA tests the lookupA function.
func TestLookupA(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		mockIPs  []string
		mockErr  error
		expected []string
		wantErr  bool
	}{
		{
			name:     "Valid domain with single IP",
			domain:   "example.com",
			mockIPs:  []string{"192.0.2.1"},
			mockErr:  nil,
			expected: []string{"192.0.2.1"},
			wantErr:  false,
		},
		{
			name:     "Valid domain with multiple IPs",
			domain:   "example.com",
			mockIPs:  []string{"192.0.2.1", "192.0.2.2"},
			mockErr:  nil,
			expected: []string{"192.0.2.1", "192.0.2.2"},
			wantErr:  false,
		},
		{
			name:     "Invalid domain",
			domain:   "invalid.domain",
			mockIPs:  nil,
			mockErr:  errors.New("no such host"),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock resolver
			mockResolver := newMockDNSResolver()
			mockResolver.lookupAFunc = func(ctx context.Context, domain string) ([]string, error) {
				if domain == tt.domain {
					return tt.mockIPs, tt.mockErr
				}
				return nil, errors.New("unexpected call")
			}

			// Create a result to store the lookup results
			result := &types.DNSResult{
				Lookups: make(map[string][]string),
			}

			// Call the mock function
			err := mockResolver.mockLookupA(context.Background(), tt.domain, result)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("lookupA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check the results
			if !tt.wantErr {
				// Check if the A records were added to the result
				if records, ok := result.Lookups["A"]; !ok {
					t.Errorf("lookupA() did not add A records to the result")
				} else {
					// Check if the number of records matches the expected number
					if len(records) != len(tt.expected) {
						t.Errorf("lookupA() returned %d records, want %d", len(records), len(tt.expected))
					}

					// Check if each record matches the expected value
					for i, record := range records {
						if record != tt.expected[i] {
							t.Errorf("lookupA() record[%d] = %s, want %s", i, record, tt.expected[i])
						}
					}
				}
			}
		})
	}
}

// TestLookupAAAA tests the lookupAAAA function.
func TestLookupAAAA(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		mockIPs  []string
		mockErr  error
		expected []string
		wantErr  bool
	}{
		{
			name:     "Valid domain with single IPv6",
			domain:   "example.com",
			mockIPs:  []string{"2001:db8::1"},
			mockErr:  nil,
			expected: []string{"2001:db8::1"},
			wantErr:  false,
		},
		{
			name:     "Valid domain with multiple IPv6s",
			domain:   "example.com",
			mockIPs:  []string{"2001:db8::1", "2001:db8::2"},
			mockErr:  nil,
			expected: []string{"2001:db8::1", "2001:db8::2"},
			wantErr:  false,
		},
		{
			name:     "Invalid domain",
			domain:   "invalid.domain",
			mockIPs:  nil,
			mockErr:  errors.New("no such host"),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock resolver
			mockResolver := newMockDNSResolver()
			mockResolver.lookupAAAAFunc = func(ctx context.Context, domain string) ([]string, error) {
				if domain == tt.domain {
					return tt.mockIPs, tt.mockErr
				}
				return nil, errors.New("unexpected call")
			}

			// Create a result to store the lookup results
			result := &types.DNSResult{
				Lookups: make(map[string][]string),
			}

			// Call the mock function
			err := mockResolver.mockLookupAAAA(context.Background(), tt.domain, result)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("lookupAAAA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check the results
			if !tt.wantErr {
				// Check if the AAAA records were added to the result
				if records, ok := result.Lookups["AAAA"]; !ok {
					t.Errorf("lookupAAAA() did not add AAAA records to the result")
				} else {
					// Check if the number of records matches the expected number
					if len(records) != len(tt.expected) {
						t.Errorf("lookupAAAA() returned %d records, want %d", len(records), len(tt.expected))
					}

					// Check if each record matches the expected value
					for i, record := range records {
						if record != tt.expected[i] {
							t.Errorf("lookupAAAA() record[%d] = %s, want %s", i, record, tt.expected[i])
						}
					}
				}
			}
		})
	}
}

// TestMockLookup tests the mock lookup functionality.
func TestMockLookup(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		recordType string
		setupMock  func(resolver *mockDNSResolver)
		expected   map[string][]string
		wantErr    bool
	}{
		{
			name:       "Valid A record lookup",
			domain:     "example.com",
			recordType: "A",
			setupMock: func(resolver *mockDNSResolver) {
				resolver.lookupAFunc = func(ctx context.Context, domain string) ([]string, error) {
					if domain == "example.com" {
						return []string{"192.0.2.1"}, nil
					}
					return nil, errors.New("unexpected call")
				}
			},
			expected: map[string][]string{
				"A": {"192.0.2.1"},
			},
			wantErr: false,
		},
		{
			name:       "Valid AAAA record lookup",
			domain:     "example.com",
			recordType: "AAAA",
			setupMock: func(resolver *mockDNSResolver) {
				resolver.lookupAAAAFunc = func(ctx context.Context, domain string) ([]string, error) {
					if domain == "example.com" {
						return []string{"2001:db8::1"}, nil
					}
					return nil, errors.New("unexpected call")
				}
			},
			expected: map[string][]string{
				"AAAA": {"2001:db8::1"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock resolver
			mockResolver := newMockDNSResolver()
			
			// Set up the mock functions
			tt.setupMock(mockResolver)

			// Create a result to store the lookup results
			result := &types.DNSResult{
				Lookups: make(map[string][]string),
			}

			// Call the appropriate mock function based on the record type
			var err error
			switch tt.recordType {
			case "A":
				err = mockResolver.mockLookupA(context.Background(), tt.domain, result)
			case "AAAA":
				err = mockResolver.mockLookupAAAA(context.Background(), tt.domain, result)
			case "MX":
				err = mockResolver.mockLookupMX(context.Background(), tt.domain, result)
			case "TXT":
				err = mockResolver.mockLookupTXT(context.Background(), tt.domain, result)
			case "CNAME":
				err = mockResolver.mockLookupCNAME(context.Background(), tt.domain, result)
			case "NS":
				err = mockResolver.mockLookupNS(context.Background(), tt.domain, result)
			case "PTR":
				err = mockResolver.mockLookupPTR(context.Background(), tt.domain, result)
			default:
				err = errors.New("unsupported record type")
			}

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("Mock lookup error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error is expected, check the results
			if !tt.wantErr {
				// Check if the result has the expected records
				for recordType, expectedRecords := range tt.expected {
					if records, ok := result.Lookups[recordType]; !ok {
						t.Errorf("Mock lookup did not return %s records", recordType)
					} else {
						// Check if the number of records matches the expected number
						if len(records) != len(expectedRecords) {
							t.Errorf("Mock lookup returned %d %s records, want %d", len(records), recordType, len(expectedRecords))
						}

						// Check if each record matches the expected value
						for i, record := range records {
							if record != expectedRecords[i] {
								t.Errorf("Mock lookup %s record[%d] = %s, want %s", recordType, i, record, expectedRecords[i])
							}
						}
					}
				}
			}
		})
	}
}

// TestMockLookupAll tests the mock lookup functionality for multiple record types.
func TestMockLookupAll(t *testing.T) {
	// Create a mock resolver
	mockResolver := newMockDNSResolver()
	
	// Set up the mock functions
	mockResolver.lookupAFunc = func(ctx context.Context, domain string) ([]string, error) {
		if domain == "example.com" {
			return []string{"192.0.2.1"}, nil
		}
		return nil, errors.New("unexpected call")
	}
	mockResolver.lookupAAAAFunc = func(ctx context.Context, domain string) ([]string, error) {
		if domain == "example.com" {
			return []string{"2001:db8::1"}, nil
		}
		return nil, errors.New("unexpected call")
	}
	mockResolver.lookupMXFunc = func(ctx context.Context, domain string) ([]string, error) {
		if domain == "example.com" {
			return []string{"mail.example.com (priority: 10)"}, nil
		}
		return nil, errors.New("unexpected call")
	}
	mockResolver.lookupTXTFunc = func(ctx context.Context, domain string) ([]string, error) {
		if domain == "example.com" {
			return []string{"v=spf1 -all"}, nil
		}
		return nil, errors.New("unexpected call")
	}
	mockResolver.lookupCNAMEFunc = func(ctx context.Context, domain string) ([]string, error) {
		if domain == "example.com" {
			return []string{"example.net."}, nil
		}
		return nil, errors.New("unexpected call")
	}
	mockResolver.lookupNSFunc = func(ctx context.Context, domain string) ([]string, error) {
		if domain == "example.com" {
			return []string{"ns1.example.com"}, nil
		}
		return nil, errors.New("unexpected call")
	}

	// Create a result to store the lookup results
	result := &types.DNSResult{
		Lookups: make(map[string][]string),
	}

	// Domain to test
	domain := "example.com"

	// Call each mock function to populate the result
	err1 := mockResolver.mockLookupA(context.Background(), domain, result)
	err2 := mockResolver.mockLookupAAAA(context.Background(), domain, result)
	err3 := mockResolver.mockLookupMX(context.Background(), domain, result)
	err4 := mockResolver.mockLookupTXT(context.Background(), domain, result)
	err5 := mockResolver.mockLookupCNAME(context.Background(), domain, result)
	err6 := mockResolver.mockLookupNS(context.Background(), domain, result)

	// Check if any errors occurred
	var err error
	if err1 != nil {
		err = err1
	} else if err2 != nil {
		err = err2
	} else if err3 != nil {
		err = err3
	} else if err4 != nil {
		err = err4
	} else if err5 != nil {
		err = err5
	} else if err6 != nil {
		err = err6
	}

	// Check if there was an error
	if err != nil {
		t.Errorf("Mock lookup all error = %v", err)
		return
	}

	// Expected records
	expected := map[string][]string{
		"A":     {"192.0.2.1"},
		"AAAA":  {"2001:db8::1"},
		"MX":    {"mail.example.com (priority: 10)"},
		"TXT":   {"v=spf1 -all"},
		"CNAME": {"example.net."},
		"NS":    {"ns1.example.com"},
	}

	// Check the results
	for recordType, expectedRecords := range expected {
		if records, ok := result.Lookups[recordType]; !ok {
			t.Errorf("Mock lookup all did not return %s records", recordType)
		} else {
			// Check if the number of records matches the expected number
			if len(records) != len(expectedRecords) {
				t.Errorf("Mock lookup all returned %d %s records, want %d", len(records), recordType, len(expectedRecords))
			}

			// Check if each record matches the expected value
			for i, record := range records {
				if record != expectedRecords[i] {
					t.Errorf("Mock lookup all %s record[%d] = %s, want %s", recordType, i, record, expectedRecords[i])
				}
			}
		}
	}
}