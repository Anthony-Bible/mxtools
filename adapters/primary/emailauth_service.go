// Package primary contains the primary adapters (implementing input ports)
package primary

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mxclone/domain/emailauth"
	"mxclone/ports/output"
)

// EmailAuthAdapter implements the EmailAuth input port
type EmailAuthAdapter struct {
	authService *emailauth.Service
	repository  output.EmailAuthRepository
}

// NewEmailAuthAdapter creates a new EmailAuth adapter
func NewEmailAuthAdapter(repository output.EmailAuthRepository) *EmailAuthAdapter {
	return &EmailAuthAdapter{
		authService: emailauth.NewService(),
		repository:  repository,
	}
}

// CheckSPF checks SPF (Sender Policy Framework) records for a domain
func (a *EmailAuthAdapter) CheckSPF(ctx context.Context, domain string, timeout time.Duration) (*emailauth.SPFResult, error) {
	// Get SPF record
	record, hasRecord, err := a.repository.GetSPFRecord(ctx, domain, timeout)
	if err != nil {
		return a.authService.ProcessSPFResult(false, "", false, nil, err), err
	}

	// If no SPF record found, return result with hasRecord=false
	if !hasRecord {
		return a.authService.ProcessSPFResult(false, "", false, nil, nil), nil
	}

	// Validate SPF record and extract mechanisms
	isValid, mechanisms, err := a.repository.ValidateSPFRecord(record)

	// Process and return the result
	return a.authService.ProcessSPFResult(true, record, isValid, mechanisms, err), nil
}

// CheckDKIM checks DKIM (DomainKeys Identified Mail) records for a domain
func (a *EmailAuthAdapter) CheckDKIM(ctx context.Context, domain string, selectors []string, timeout time.Duration) (*emailauth.DKIMResult, error) {
	// If no selectors provided, use common ones
	if len(selectors) == 0 {
		selectors = []string{"default", "selector1", "selector2", "dkim", "mail"}
	}

	// Get DKIM records for each selector
	records := make(map[string]string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, selector := range selectors {
		wg.Add(1)
		go func(sel string) {
			defer wg.Done()

			// Get DKIM record for this selector
			record, found, err := a.repository.GetDKIMRecord(ctx, domain, sel, timeout)

			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("error checking DKIM selector %s: %w", sel, err)
				}
				mu.Unlock()
				return
			}

			if found {
				mu.Lock()
				records[sel] = record
				mu.Unlock()
			}
		}(selector)
	}

	// Wait for all checks to complete
	wg.Wait()

	// Check if we found any DKIM records
	hasRecords := len(records) > 0

	// Validate all found DKIM records
	isValid := true
	for selector, record := range records {
		valid, err := a.repository.ValidateDKIMRecord(record)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("error validating DKIM record for selector %s: %w", selector, err)
			}
			isValid = false
		} else if !valid {
			isValid = false
		}
	}

	// Process and return the result
	return a.authService.ProcessDKIMResult(hasRecords, records, isValid, firstErr), nil
}

// CheckDMARC checks DMARC (Domain-based Message Authentication) records for a domain
func (a *EmailAuthAdapter) CheckDMARC(ctx context.Context, domain string, timeout time.Duration) (*emailauth.DMARCResult, error) {
	// Get DMARC record
	record, hasRecord, err := a.repository.GetDMARCRecord(ctx, domain, timeout)
	if err != nil {
		return a.authService.ProcessDMARCResult(false, "", false, "", "", 0, err), err
	}

	// If no DMARC record found, return result with hasRecord=false
	if !hasRecord {
		return a.authService.ProcessDMARCResult(false, "", false, "", "", 0, nil), nil
	}

	// Parse DMARC record
	isValid, policy, subdomainPolicy, percentage, err := a.repository.ParseDMARCRecord(record)

	// Process and return the result
	return a.authService.ProcessDMARCResult(true, record, isValid, policy, subdomainPolicy, percentage, err), nil
}

// CheckAll performs SPF, DKIM, and DMARC checks for a domain
func (a *EmailAuthAdapter) CheckAll(ctx context.Context, domain string, dkimSelectors []string, timeout time.Duration) (*emailauth.AuthResult, error) {
	var spfResult *emailauth.SPFResult
	var dkimResult *emailauth.DKIMResult
	var dmarcResult *emailauth.DMARCResult
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// Check SPF in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		spfResult, err = a.CheckSPF(ctx, domain, timeout)
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = fmt.Errorf("SPF check error: %w", err)
			}
			mu.Unlock()
		}
	}()

	// Check DKIM in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		dkimResult, err = a.CheckDKIM(ctx, domain, dkimSelectors, timeout)
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = fmt.Errorf("DKIM check error: %w", err)
			}
			mu.Unlock()
		}
	}()

	// Check DMARC in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		dmarcResult, err = a.CheckDMARC(ctx, domain, timeout)
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = fmt.Errorf("DMARC check error: %w", err)
			}
			mu.Unlock()
		}
	}()

	// Wait for all checks to complete
	wg.Wait()

	// Process and return the aggregated result
	return a.authService.ProcessAuthResult(domain, spfResult, dkimResult, dmarcResult, firstErr), firstErr
}

// GetAuthSummary returns a human-readable summary of email authentication checks
func (a *EmailAuthAdapter) GetAuthSummary(result *emailauth.AuthResult) string {
	return a.authService.FormatAuthSummary(result)
}
