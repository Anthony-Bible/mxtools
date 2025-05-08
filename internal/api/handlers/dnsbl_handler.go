package handlers

import (
	"encoding/json"
	"io/ioutil"
	"mxclone/internal/api/models"
	"mxclone/pkg/validation"
	"mxclone/ports/input"
	"net"
	"net/http"
	"time"
)

// DNSBLHandler encapsulates handlers for DNSBL operations
type DNSBLHandler struct {
	dnsblService input.DNSBLPort // Using the interface (port) instead of direct implementation
}

// NewDNSBLHandler creates a new DNSBL handler with the given DNSBL service
func NewDNSBLHandler(dnsblService input.DNSBLPort) *DNSBLHandler {
	return &DNSBLHandler{
		dnsblService: dnsblService,
	}
}

// HandleDNSBLCheck handles DNSBL check requests
func (h *DNSBLHandler) HandleDNSBLCheck(w http.ResponseWriter, r *http.Request) {
	var ip string

	// Check if target is provided in the context (from path parameter)
	if ip = r.PathValue("target"); ip == "" {

		// If not from path, read from request body
		body, err := ioutil.ReadAll(r.Body)
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
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		ip = req.Target
	}

	// Resolve domain to IP if necessary
	if err := validation.ValidateIP(ip); err != nil {
		if derr := validation.ValidateDomain(ip); derr != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid IP or domain"})
			return
		}
		ips, err := net.LookupIP(ip)
		if err != nil || len(ips) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "could not resolve domain to IP"})
			return
		}
		ip = ips[0].String()
	}

	// Standard blacklist zones - could be moved to configuration
	zones := []string{"bl.spamcop.net", "dnsbl.sorbs.net"}

	// Use the DNSBL service through the port interface
	result, err := h.dnsblService.CheckMultipleBlacklists(r.Context(), ip, zones, 10*time.Second)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "DNSBL check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert to API response model
	response := models.FromBlacklistResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
