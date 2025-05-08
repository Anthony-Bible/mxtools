package handlers

import (
	"encoding/json"
	"io"
	"mxclone/domain/dns"
	"mxclone/internal/api/models"
	apivalidation "mxclone/internal/api/validation"
	"mxclone/ports/input"
	"net/http"
)

// DNSHandler encapsulates handlers for DNS operations
type DNSHandler struct {
	dnsService input.DNSPort // Using the interface (port) instead of direct implementation
}

// NewDNSHandler creates a new DNS handler with the given DNS service
func NewDNSHandler(dnsService input.DNSPort) *DNSHandler {
	return &DNSHandler{
		dnsService: dnsService,
	}
}

// HandleDNSLookup handles DNS lookup requests
func (h *DNSHandler) HandleDNSLookup(w http.ResponseWriter, r *http.Request) {
	var domain string
	var recordType dns.RecordType
	if domain = r.PathValue("domain"); domain == "" {

		// If not from path, read from request body
		body, err := io.ReadAll(r.Body)
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
		validationResult := apivalidation.ValidateDNSRequest(&req)
		if !validationResult.Valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":       "Validation failed",
				"code":        http.StatusBadRequest,
				"validations": validationResult.Errors,
			})
			return
		}

		domain = req.Target
		if req.Option != "" {
			recordType = dns.RecordType(req.Option)
		}
	}

	// Use the DNS service through the port interface
	var result *dns.DNSResult
	var err error

	if recordType != "" {
		result, err = h.dnsService.Lookup(r.Context(), domain, recordType)
	} else {
		result, err = h.dnsService.LookupAll(r.Context(), domain)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "DNS lookup failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert domain result to API response
	response := models.FromDNSResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
