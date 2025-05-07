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

// ValidateSMTPRelayTestRequest validates an SMTP relay test request
func ValidateSMTPRelayTestRequest(req *models.SMTPRelayTestRequest) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check if host is empty
	if req.Host == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "host",
			Message: "host cannot be empty",
		})
	} else {
		// Check if host is a valid domain or IP
		if err := validation.ValidateDomain(req.Host); err != nil {
			if err := validation.ValidateIP(req.Host); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   "host",
					Message: "invalid domain or IP address",
				})
			}
		}
	}

	// Check if fromAddress is a valid email
	if req.FromAddress == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "fromAddress",
			Message: "fromAddress cannot be empty",
		})
	} else if err := validation.ValidateEmail(req.FromAddress); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "fromAddress",
			Message: "invalid email address: " + err.Error(),
		})
	}

	// Check if toAddress is a valid email
	if req.ToAddress == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "toAddress",
			Message: "toAddress cannot be empty",
		})
	} else if err := validation.ValidateEmail(req.ToAddress); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "toAddress",
			Message: "invalid email address: " + err.Error(),
		})
	}

	// Check port is valid (if specified)
	if req.Port < 0 || req.Port > 65535 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "port",
			Message: "port must be between 0 and 65535",
		})
	}

	// If authentication is required, validate credentials
	if req.Authentication {
		if req.Username == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "username",
				Message: "username is required when authentication is enabled",
			})
		}

		if req.Password == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "password",
				Message: "password is required when authentication is enabled",
			})
		}
	}

	return result
}
