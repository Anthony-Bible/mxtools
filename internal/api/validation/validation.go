package validation

import (
	"fmt"
	"mxclone/internal/api/models"
	"mxclone/pkg/validation"
	"net"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult contains validation errors if any
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidateCheckRequest validates a check request
func ValidateCheckRequest(req *models.CheckRequest) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check if target is empty
	if req.Target == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "target",
			Message: "target cannot be empty",
		})
		return result
	}

	return result
}

// ValidateDNSRequest validates a DNS request
func ValidateDNSRequest(req *models.CheckRequest) *ValidationResult {
	result := ValidateCheckRequest(req)
	if !result.Valid {
		return result
	}

	// Check if target is a valid domain
	if err := validation.ValidateDomain(req.Target); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "target",
			Message: "invalid domain name: " + err.Error(),
		})
	}

	return result
}

// ValidateBlacklistRequest validates a blacklist request
func ValidateBlacklistRequest(req *models.CheckRequest) *ValidationResult {
	result := ValidateCheckRequest(req)
	if !result.Valid {
		return result
	}

	// Check if target is a valid IP or domain
	if err := validation.ValidateIP(req.Target); err != nil {
		if derr := validation.ValidateDomain(req.Target); derr != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "target",
				Message: "invalid IP or domain name",
			})
		} else {
			// It's a valid domain, attempt to resolve it
			ips, err := net.LookupIP(req.Target)
			if err != nil || len(ips) == 0 {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "target",
					Message: "could not resolve domain to IP",
				})
			}
		}
	}

	return result
}

// ValidateSMTPRequest validates an SMTP request
func ValidateSMTPRequest(req *models.CheckRequest) *ValidationResult {
	result := ValidateCheckRequest(req)
	if !result.Valid {
		return result
	}

	// Check if target is a valid domain or IP
	if err := validation.ValidateDomain(req.Target); err != nil {
		if err := validation.ValidateIP(req.Target); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "target",
				Message: "invalid domain or IP address",
			})
		}
	}

	return result
}

// ValidateEmailAuthRequest validates an email authentication request
func ValidateEmailAuthRequest(req *models.CheckRequest) *ValidationResult {
	result := ValidateCheckRequest(req)
	if !result.Valid {
		return result
	}

	// Check if target is a valid domain
	if err := validation.ValidateDomain(req.Target); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "target",
			Message: "invalid domain name: " + err.Error(),
		})
	}

	return result
}

// ValidateNetworkToolRequest validates a network tool request
func ValidateNetworkToolRequest(req *models.CheckRequest, tool string) *ValidationResult {
	result := ValidateCheckRequest(req)
	if !result.Valid {
		return result
	}

	// Validate the tool
	validTools := map[string]bool{"ping": true, "traceroute": true, "whois": true}
	if !validTools[strings.ToLower(tool)] {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "tool",
			Message: fmt.Sprintf("invalid tool: %s (valid options: ping, traceroute, whois)", tool),
		})
	}

	// Check if target is a valid domain or IP
	if err := validation.ValidateDomain(req.Target); err != nil {
		if err := validation.ValidateIP(req.Target); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "target",
				Message: "invalid domain or IP address",
			})
		}
	}

	return result
}
