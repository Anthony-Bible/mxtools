// Package emailauth provides functionality for checking email authentication mechanisms.
package emailauth

import (
	"testing"
)

// TestParseSPFRecord tests the ParseSPFRecord function
func TestParseSPFRecord(t *testing.T) {
	tests := []struct {
		name     string
		record   string
		expected *SPFRecord
		wantErr  bool
	}{
		{
			name:    "Empty record",
			record:  "",
			wantErr: true,
		},
		{
			name:    "Invalid record (no v=spf1)",
			record:  "invalid record",
			wantErr: true,
		},
		{
			name:   "Simple record",
			record: "v=spf1 -all",
			expected: &SPFRecord{
				Raw:        "v=spf1 -all",
				Version:    "v=spf1",
				Mechanisms: []string{"-all"},
				Modifiers:  map[string]string{},
				All:        "-all",
			},
			wantErr: false,
		},
		{
			name:   "Complex record",
			record: "v=spf1 ip4:192.168.0.1/16 include:example.com redirect=example.net -all",
			expected: &SPFRecord{
				Raw:        "v=spf1 ip4:192.168.0.1/16 include:example.com redirect=example.net -all",
				Version:    "v=spf1",
				Mechanisms: []string{"ip4:192.168.0.1/16", "include:example.com", "-all"},
				Modifiers:  map[string]string{"redirect": "example.net"},
				Redirect:   "example.net",
				All:        "-all",
			},
			wantErr: false,
		},
		{
			name:   "Record with multiple modifiers",
			record: "v=spf1 ip4:192.168.0.1 exp=explain.example.com redirect=example.net ~all",
			expected: &SPFRecord{
				Raw:        "v=spf1 ip4:192.168.0.1 exp=explain.example.com redirect=example.net ~all",
				Version:    "v=spf1",
				Mechanisms: []string{"ip4:192.168.0.1", "~all"},
				Modifiers:  map[string]string{"exp": "explain.example.com", "redirect": "example.net"},
				Redirect:   "example.net",
				All:        "~all",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSPFRecord(tt.record)
			
			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSPFRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If we expect an error, we don't need to check the result
			if tt.wantErr {
				return
			}
			
			// Check if the result matches the expected result
			if result.Version != tt.expected.Version {
				t.Errorf("ParseSPFRecord().Version = %v, want %v", result.Version, tt.expected.Version)
			}
			
			if result.All != tt.expected.All {
				t.Errorf("ParseSPFRecord().All = %v, want %v", result.All, tt.expected.All)
			}
			
			if result.Redirect != tt.expected.Redirect {
				t.Errorf("ParseSPFRecord().Redirect = %v, want %v", result.Redirect, tt.expected.Redirect)
			}
			
			// Check mechanisms
			if len(result.Mechanisms) != len(tt.expected.Mechanisms) {
				t.Errorf("ParseSPFRecord().Mechanisms has %d items, want %d", len(result.Mechanisms), len(tt.expected.Mechanisms))
			} else {
				for i, mech := range result.Mechanisms {
					if mech != tt.expected.Mechanisms[i] {
						t.Errorf("ParseSPFRecord().Mechanisms[%d] = %v, want %v", i, mech, tt.expected.Mechanisms[i])
					}
				}
			}
			
			// Check modifiers
			if len(result.Modifiers) != len(tt.expected.Modifiers) {
				t.Errorf("ParseSPFRecord().Modifiers has %d items, want %d", len(result.Modifiers), len(tt.expected.Modifiers))
			} else {
				for key, expectedValue := range tt.expected.Modifiers {
					if value, ok := result.Modifiers[key]; !ok {
						t.Errorf("ParseSPFRecord().Modifiers missing key %s", key)
					} else if value != expectedValue {
						t.Errorf("ParseSPFRecord().Modifiers[%s] = %v, want %v", key, value, expectedValue)
					}
				}
			}
		})
	}
}

// TestValidateSPF tests the ValidateSPF function
func TestValidateSPF(t *testing.T) {
	tests := []struct {
		name     string
		spf      *SPFRecord
		expected string
		wantErr  bool
	}{
		{
			name:     "Nil record",
			spf:      nil,
			expected: "",
			wantErr:  true,
		},
		{
			name: "Invalid version",
			spf: &SPFRecord{
				Version: "invalid",
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "No all mechanism or redirect",
			spf: &SPFRecord{
				Version: "v=spf1",
			},
			expected: "neutral",
			wantErr:  true,
		},
		{
			name: "Pass all",
			spf: &SPFRecord{
				Version: "v=spf1",
				All:     "+all",
			},
			expected: "pass",
			wantErr:  false,
		},
		{
			name: "Fail all",
			spf: &SPFRecord{
				Version: "v=spf1",
				All:     "-all",
			},
			expected: "fail",
			wantErr:  false,
		},
		{
			name: "Softfail all",
			spf: &SPFRecord{
				Version: "v=spf1",
				All:     "~all",
			},
			expected: "softfail",
			wantErr:  false,
		},
		{
			name: "Neutral all",
			spf: &SPFRecord{
				Version: "v=spf1",
				All:     "?all",
			},
			expected: "neutral",
			wantErr:  false,
		},
		{
			name: "Redirect",
			spf: &SPFRecord{
				Version:  "v=spf1",
				Redirect: "example.com",
			},
			expected: "redirect",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSPF(tt.spf)
			
			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSPF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Check if the result matches the expected result
			if result != tt.expected {
				t.Errorf("ValidateSPF() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseDMARCRecord tests the ParseDMARCRecord function
func TestParseDMARCRecord(t *testing.T) {
	tests := []struct {
		name     string
		record   string
		expected *DMARCRecord
		wantErr  bool
	}{
		{
			name:    "Empty record",
			record:  "",
			wantErr: true,
		},
		{
			name:    "Invalid record (no v=DMARC1)",
			record:  "invalid record",
			wantErr: true,
		},
		{
			name:   "Simple record",
			record: "v=DMARC1; p=none;",
			expected: &DMARCRecord{
				Raw:     "v=DMARC1; p=none;",
				Version: "v=DMARC1",
				Policy:  "none",
				Pct:     100, // Default
			},
			wantErr: false,
		},
		{
			name:   "Complex record",
			record: "v=DMARC1; p=reject; sp=quarantine; pct=50; rua=mailto:dmarc@example.com; ruf=mailto:forensics@example.com; adkim=s; aspf=r;",
			expected: &DMARCRecord{
				Raw:       "v=DMARC1; p=reject; sp=quarantine; pct=50; rua=mailto:dmarc@example.com; ruf=mailto:forensics@example.com; adkim=s; aspf=r;",
				Version:   "v=DMARC1",
				Policy:    "reject",
				SubPolicy: "quarantine",
				Pct:       50,
				RUA:       "mailto:dmarc@example.com",
				RUF:       "mailto:forensics@example.com",
				ADKIM:     "s",
				ASPF:      "r",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDMARCRecord(tt.record)
			
			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDMARCRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If we expect an error, we don't need to check the result
			if tt.wantErr {
				return
			}
			
			// Check if the result matches the expected result
			if result.Version != tt.expected.Version {
				t.Errorf("ParseDMARCRecord().Version = %v, want %v", result.Version, tt.expected.Version)
			}
			
			if result.Policy != tt.expected.Policy {
				t.Errorf("ParseDMARCRecord().Policy = %v, want %v", result.Policy, tt.expected.Policy)
			}
			
			if result.SubPolicy != tt.expected.SubPolicy {
				t.Errorf("ParseDMARCRecord().SubPolicy = %v, want %v", result.SubPolicy, tt.expected.SubPolicy)
			}
			
			if result.Pct != tt.expected.Pct {
				t.Errorf("ParseDMARCRecord().Pct = %v, want %v", result.Pct, tt.expected.Pct)
			}
			
			if result.RUA != tt.expected.RUA {
				t.Errorf("ParseDMARCRecord().RUA = %v, want %v", result.RUA, tt.expected.RUA)
			}
			
			if result.RUF != tt.expected.RUF {
				t.Errorf("ParseDMARCRecord().RUF = %v, want %v", result.RUF, tt.expected.RUF)
			}
			
			if result.ADKIM != tt.expected.ADKIM {
				t.Errorf("ParseDMARCRecord().ADKIM = %v, want %v", result.ADKIM, tt.expected.ADKIM)
			}
			
			if result.ASPF != tt.expected.ASPF {
				t.Errorf("ParseDMARCRecord().ASPF = %v, want %v", result.ASPF, tt.expected.ASPF)
			}
		})
	}
}

// TestParseDKIMRecord tests the ParseDKIMRecord function
func TestParseDKIMRecord(t *testing.T) {
	tests := []struct {
		name     string
		record   string
		expected *DKIMRecord
		wantErr  bool
	}{
		{
			name:    "Empty record",
			record:  "",
			wantErr: true,
		},
		{
			name:   "Simple record",
			record: "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+W6scd3XWwvC/hPRksfDYFi3ztgyS9OSqnnjtNQeDdTSD1DRx/xFar2wjmzxp2+SnJ5pspaF77VZveN3P/HVmXZVghr3asoV9WBx/uW1nDIUxU35L4juXiTwsMAbgMyh3NqIKTNKyMDy4P8vpEhtH1iv/BrwMdBjHDVCycB8WnwIDAQAB;",
			expected: &DKIMRecord{
				Raw:       "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+W6scd3XWwvC/hPRksfDYFi3ztgyS9OSqnnjtNQeDdTSD1DRx/xFar2wjmzxp2+SnJ5pspaF77VZveN3P/HVmXZVghr3asoV9WBx/uW1nDIUxU35L4juXiTwsMAbgMyh3NqIKTNKyMDy4P8vpEhtH1iv/BrwMdBjHDVCycB8WnwIDAQAB;",
				Version:   "DKIM1",
				KeyType:   "rsa",
				PublicKey: "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+W6scd3XWwvC/hPRksfDYFi3ztgyS9OSqnnjtNQeDdTSD1DRx/xFar2wjmzxp2+SnJ5pspaF77VZveN3P/HVmXZVghr3asoV9WBx/uW1nDIUxU35L4juXiTwsMAbgMyh3NqIKTNKyMDy4P8vpEhtH1iv/BrwMdBjHDVCycB8WnwIDAQAB",
			},
			wantErr: false,
		},
		{
			name:   "Complex record",
			record: "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+W6scd3XWwvC/hPRksfDYFi3ztgyS9OSqnnjtNQeDdTSD1DRx/xFar2wjmzxp2+SnJ5pspaF77VZveN3P/HVmXZVghr3asoV9WBx/uW1nDIUxU35L4juXiTwsMAbgMyh3NqIKTNKyMDy4P8vpEhtH1iv/BrwMdBjHDVCycB8WnwIDAQAB; n=Notes; s=email; t=s;",
			expected: &DKIMRecord{
				Raw:       "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+W6scd3XWwvC/hPRksfDYFi3ztgyS9OSqnnjtNQeDdTSD1DRx/xFar2wjmzxp2+SnJ5pspaF77VZveN3P/HVmXZVghr3asoV9WBx/uW1nDIUxU35L4juXiTwsMAbgMyh3NqIKTNKyMDy4P8vpEhtH1iv/BrwMdBjHDVCycB8WnwIDAQAB; n=Notes; s=email; t=s;",
				Version:   "DKIM1",
				KeyType:   "rsa",
				PublicKey: "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+W6scd3XWwvC/hPRksfDYFi3ztgyS9OSqnnjtNQeDdTSD1DRx/xFar2wjmzxp2+SnJ5pspaF77VZveN3P/HVmXZVghr3asoV9WBx/uW1nDIUxU35L4juXiTwsMAbgMyh3NqIKTNKyMDy4P8vpEhtH1iv/BrwMdBjHDVCycB8WnwIDAQAB",
				Notes:     "Notes",
				Service:   "email",
				Flags:     "s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDKIMRecord(tt.record)
			
			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDKIMRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// If we expect an error, we don't need to check the result
			if tt.wantErr {
				return
			}
			
			// Check if the result matches the expected result
			if result.Version != tt.expected.Version {
				t.Errorf("ParseDKIMRecord().Version = %v, want %v", result.Version, tt.expected.Version)
			}
			
			if result.KeyType != tt.expected.KeyType {
				t.Errorf("ParseDKIMRecord().KeyType = %v, want %v", result.KeyType, tt.expected.KeyType)
			}
			
			if result.PublicKey != tt.expected.PublicKey {
				t.Errorf("ParseDKIMRecord().PublicKey = %v, want %v", result.PublicKey, tt.expected.PublicKey)
			}
			
			if result.Notes != tt.expected.Notes {
				t.Errorf("ParseDKIMRecord().Notes = %v, want %v", result.Notes, tt.expected.Notes)
			}
			
			if result.Service != tt.expected.Service {
				t.Errorf("ParseDKIMRecord().Service = %v, want %v", result.Service, tt.expected.Service)
			}
			
			if result.Flags != tt.expected.Flags {
				t.Errorf("ParseDKIMRecord().Flags = %v, want %v", result.Flags, tt.expected.Flags)
			}
		})
	}
}

// TestAnalyzeEmailHeader tests the AnalyzeEmailHeader function
func TestAnalyzeEmailHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected map[string]string
	}{
		{
			name:     "Empty header",
			header:   "",
			expected: map[string]string{},
		},
		{
			name:     "Header without Authentication-Results",
			header:   "From: sender@example.com\nTo: recipient@example.com\nSubject: Test",
			expected: map[string]string{},
		},
		{
			name:     "Header with Authentication-Results",
			header:   "From: sender@example.com\nTo: recipient@example.com\nSubject: Test\nAuthentication-Results: example.com; spf=pass; dkim=pass; dmarc=pass",
			expected: map[string]string{
				"SPF":   "pass",
				"DKIM":  "pass",
				"DMARC": "pass",
			},
		},
		{
			name:     "Header with partial Authentication-Results",
			header:   "From: sender@example.com\nTo: recipient@example.com\nSubject: Test\nAuthentication-Results: example.com; spf=pass; dkim=fail",
			expected: map[string]string{
				"SPF":  "pass",
				"DKIM": "fail",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AnalyzeEmailHeader(tt.header)
			
			// Check if there was an error
			if err != nil {
				t.Errorf("AnalyzeEmailHeader() error = %v", err)
				return
			}
			
			// Check if the result matches the expected result
			if len(result) != len(tt.expected) {
				t.Errorf("AnalyzeEmailHeader() returned %d results, want %d", len(result), len(tt.expected))
			}
			
			for key, expectedValue := range tt.expected {
				if value, ok := result[key]; !ok {
					t.Errorf("AnalyzeEmailHeader() missing key %s", key)
				} else if value != expectedValue {
					t.Errorf("AnalyzeEmailHeader()[%s] = %v, want %v", key, value, expectedValue)
				}
			}
		})
	}
}