package validation

import (
	"encoding/json"
	"mxclone/internal/api/models"
)

// JSONValidator adapts the validation functions to be used with the middleware
type JSONValidator struct{}

// NewJSONValidator creates a new JSON validator adapter
func NewJSONValidator() *JSONValidator {
	return &JSONValidator{}
}

// ValidateDNSRequestJSON validates a DNS request from JSON
func (v *JSONValidator) ValidateDNSRequestJSON(body []byte) (bool, map[string]interface{}) {
	var req models.CheckRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false, map[string]interface{}{
			"error": "Invalid JSON format: " + err.Error(),
		}
	}

	result := ValidateDNSRequest(&req)
	return result.Valid, v.formatErrors(result)
}

// ValidateBlacklistRequestJSON validates a blacklist request from JSON
func (v *JSONValidator) ValidateBlacklistRequestJSON(body []byte) (bool, map[string]interface{}) {
	var req models.CheckRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false, map[string]interface{}{
			"error": "Invalid JSON format: " + err.Error(),
		}
	}

	result := ValidateBlacklistRequest(&req)
	return result.Valid, v.formatErrors(result)
}

// ValidateSMTPRequestJSON validates an SMTP request from JSON
func (v *JSONValidator) ValidateSMTPRequestJSON(body []byte) (bool, map[string]interface{}) {
	var req models.CheckRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false, map[string]interface{}{
			"error": "Invalid JSON format: " + err.Error(),
		}
	}

	result := ValidateSMTPRequest(&req)
	return result.Valid, v.formatErrors(result)
}

// ValidateSMTPRelayTestRequestJSON validates an SMTP relay test request from JSON
func (v *JSONValidator) ValidateSMTPRelayTestRequestJSON(body []byte) (bool, map[string]interface{}) {
	var req models.SMTPRelayTestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false, map[string]interface{}{
			"error": "Invalid JSON format: " + err.Error(),
		}
	}

	result := ValidateSMTPRelayTestRequest(&req)
	return result.Valid, v.formatErrors(result)
}

// ValidateEmailAuthRequestJSON validates an email authentication request from JSON
func (v *JSONValidator) ValidateEmailAuthRequestJSON(body []byte) (bool, map[string]interface{}) {
	var req models.CheckRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false, map[string]interface{}{
			"error": "Invalid JSON format: " + err.Error(),
		}
	}

	result := ValidateEmailAuthRequest(&req)
	return result.Valid, v.formatErrors(result)
}

// ValidateNetworkToolRequestJSON validates a network tool request from JSON
func (v *JSONValidator) ValidateNetworkToolRequestJSON(body []byte, tool string) (bool, map[string]interface{}) {
	var req models.CheckRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return false, map[string]interface{}{
			"error": "Invalid JSON format: " + err.Error(),
		}
	}

	result := ValidateNetworkToolRequest(&req, tool)
	return result.Valid, v.formatErrors(result)
}

// ParamValidator adapts URL parameter validation functions to be used with middleware
type ParamValidator struct{}

// NewParamValidator creates a new param validator adapter
func NewParamValidator() *ParamValidator {
	return &ParamValidator{}
}

// ValidateDomainParam validates a domain parameter
func (v *ParamValidator) ValidateDomainParam(params map[string]string) (bool, map[string]interface{}) {
	domain, exists := params["domain"]
	if !exists || domain == "" {
		return false, map[string]interface{}{
			"error": "Domain parameter is required",
		}
	}

	// Create a request object to reuse existing validation
	req := &models.CheckRequest{Target: domain}
	result := ValidateDNSRequest(req)
	return result.Valid, formatErrors(result)
}

// ValidateHostParam validates a host parameter
func (v *ParamValidator) ValidateHostParam(params map[string]string) (bool, map[string]interface{}) {
	host, exists := params["host"]
	if !exists || host == "" {
		return false, map[string]interface{}{
			"error": "Host parameter is required",
		}
	}

	// Create a request object to reuse existing validation
	req := &models.CheckRequest{Target: host}
	result := ValidateSMTPRequest(req)
	return result.Valid, formatErrors(result)
}

// ValidateDKIMParams validates domain and selector parameters for DKIM checks
func (v *ParamValidator) ValidateDKIMParams(params map[string]string) (bool, map[string]interface{}) {
	domain, domainExists := params["domain"]
	selector, selectorExists := params["selector"]

	errors := make(map[string]interface{})
	valid := true

	if !domainExists || domain == "" {
		errors["domain"] = "Domain parameter is required"
		valid = false
	}

	if !selectorExists || selector == "" {
		errors["selector"] = "Selector parameter is required"
		valid = false
	}

	if !valid {
		return false, errors
	}

	// Further validate the domain format
	req := &models.CheckRequest{Target: domain}
	result := ValidateDNSRequest(req)
	return result.Valid, formatErrors(result)
}

// Helper function to format validation errors
func (v *JSONValidator) formatErrors(result *ValidationResult) map[string]interface{} {
	if result.Valid {
		return nil
	}

	errorMap := make(map[string]interface{})
	for _, err := range result.Errors {
		errorMap[err.Field] = err.Message
	}
	return errorMap
}

// Helper function to format validation errors
func formatErrors(result *ValidationResult) map[string]interface{} {
	if result.Valid {
		return nil
	}

	errorMap := make(map[string]interface{})
	for _, err := range result.Errors {
		errorMap[err.Field] = err.Message
	}
	return errorMap
}
