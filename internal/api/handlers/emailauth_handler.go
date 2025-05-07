package handlers

import (
	"encoding/json"
	"io" // Changed from ioutil
	"mxclone/domain/emailauth"
	"mxclone/internal/api/models"
	apivalidation "mxclone/internal/api/validation"
	"mxclone/ports/input"
	"net/http"
	"strings"
	"time"
)

// EmailAuthHandler encapsulates handlers for email authentication operations
type EmailAuthHandler struct {
	emailAuthService input.EmailAuthPort // Using the interface (port) instead of direct implementation
}

// NewEmailAuthHandler creates a new email authentication handler with the given service
func NewEmailAuthHandler(emailAuthService input.EmailAuthPort) *EmailAuthHandler {
	return &EmailAuthHandler{
		emailAuthService: emailAuthService,
	}
}

// HandleSPFCheck handles SPF record check requests
func (h *EmailAuthHandler) HandleSPFCheck(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	if domain == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Domain path parameter is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Default timeout is 10 seconds
	timeout := 10 * time.Second
	timeoutStr := r.URL.Query().Get("timeout")
	if timeoutStr != "" {
		var err error
		timeoutDuration, err := time.ParseDuration(timeoutStr)
		if err == nil {
			timeout = timeoutDuration
		}
	}

	// Use the email authentication service through the port interface
	result, err := h.emailAuthService.CheckSPF(r.Context(), domain, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "SPF check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert result to API response
	response := models.FromSPFResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleDKIMCheck handles DKIM record check requests
func (h *EmailAuthHandler) HandleDKIMCheck(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	selector := r.PathValue("selector")

	if domain == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Domain path parameter is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Default timeout is 10 seconds
	timeout := 10 * time.Second
	timeoutStr := r.URL.Query().Get("timeout")
	if timeoutStr != "" {
		var err error
		timeoutDuration, err := time.ParseDuration(timeoutStr)
		if err == nil {
			timeout = timeoutDuration
		}
	}

	// If no selector is provided, try all default selectors
	if selector == "" {
		defaultSelectors := []string{"mail", "google"}
		var combinedResults []*emailauth.DKIMResult
		var lastError error
		foundAnyValid := false

		// Try each default selector and collect all results
		for _, defaultSelector := range defaultSelectors {
			result, err := h.emailAuthService.CheckDKIM(r.Context(), domain, []string{defaultSelector}, timeout)
			if err == nil && result != nil {
				combinedResults = append(combinedResults, result)
				if result.IsValid {
					foundAnyValid = true

				}
			} else {
				lastError = err
			}
		}

		// If we couldn't find any valid DKIM record with default selectors
		if !foundAnyValid {
			errorMsg := "DKIM check failed: No valid DKIM records found with default selectors"
			if lastError != nil {
				errorMsg = lastError.Error()
			}

			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.APIError{
				Error:   errorMsg,
				Code:    http.StatusNotFound,
				Details: "Tried selectors: mail, google",
			})
			return
		}

		// Create a combined response with all results
		combinedResponse := models.CombinedDKIMResponse{
			Domain:    domain,
			Results:   make([]models.DKIMResponse, 0, len(combinedResults)),
			IsValid:   foundAnyValid,
			Selectors: defaultSelectors,
		}

		for i, result := range combinedResults {
			response := models.FromDKIMResult(result)
			response.Selector = defaultSelectors[i]
			combinedResponse.Results = append(combinedResponse.Results, *response)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(combinedResponse)
		return
	} else {
		// Use the selector provided in the request
		result, err := h.emailAuthService.CheckDKIM(r.Context(), domain, []string{selector}, timeout)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.APIError{
				Error:   "DKIM check failed",
				Code:    http.StatusInternalServerError,
				Details: err.Error(),
			})
			return
		}

		// Convert result to API response
		dkimResult := result
		response := models.FromDKIMResult(dkimResult)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// HandleDMARCCheck handles DMARC record check requests
func (h *EmailAuthHandler) HandleDMARCCheck(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")

	if domain == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Domain path parameter is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Default timeout is 10 seconds
	timeout := 10 * time.Second
	timeoutStr := r.URL.Query().Get("timeout")
	if timeoutStr != "" {
		var err error
		timeoutDuration, err := time.ParseDuration(timeoutStr)
		if err == nil {
			timeout = timeoutDuration
		}
	}

	// Use the email authentication service through the port interface
	result, err := h.emailAuthService.CheckDMARC(r.Context(), domain, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "DMARC check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert result to API response
	response := models.FromDMARCResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleEmailAuth handles email authentication check requests
func (h *EmailAuthHandler) HandleEmailAuth(w http.ResponseWriter, r *http.Request) {
	// Read and parse request body
	body, err := io.ReadAll(r.Body) // Changed from ioutil.ReadAll
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Invalid request body",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	var req models.CheckRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Invalid JSON format",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	// Validate the request
	validationResult := apivalidation.ValidateEmailAuthRequest(&req)
	if !validationResult.Valid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Validation failed",
			"code":        http.StatusBadRequest,
			"validations": validationResult.Errors,
		})
		return
	}

	// Default timeout is 10 seconds
	timeout := 10 * time.Second

	// Default DKIM selectors if not specified
	dkimSelectors := []string{"default", "dkim", "mail"}
	if req.Option != "" {
		// If specific selectors were provided in the option field
		selectors := strings.Split(req.Option, ",")
		if len(selectors) > 0 {
			dkimSelectors = selectors
		}
	}

	// Use the email authentication service through the port interface
	result, err := h.emailAuthService.CheckAll(r.Context(), req.Target, dkimSelectors, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Email authentication check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Build a simplified response with overall results
	response := &models.EmailAuthResponse{
		Domain: req.Target,
		SPF:    result.SPF != nil && result.SPF.IsValid,
		DKIM:   result.DKIM != nil && result.DKIM.IsValid,
		DMARC:  result.DMARC != nil && result.DMARC.IsValid,
		AllPassed: result.SPF != nil && result.SPF.IsValid &&
			result.DKIM != nil && result.DKIM.IsValid &&
			result.DMARC != nil && result.DMARC.IsValid,
	}

	if err != nil {
		response.Error = err.Error()
	} else if result.Error != "" {
		response.Error = result.Error
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
