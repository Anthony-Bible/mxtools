package handlers

import (
	"encoding/json"
	"io/ioutil"
	"mxclone/internal/api/models"
	apivalidation "mxclone/internal/api/validation"
	"mxclone/ports/input"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// SMTPHandler encapsulates handlers for SMTP operations
type SMTPHandler struct {
	smtpService input.SMTPPort // Using the interface (port) instead of direct implementation
}

// NewSMTPHandler creates a new SMTP handler with the given SMTP service
func NewSMTPHandler(smtpService input.SMTPPort) *SMTPHandler {
	return &SMTPHandler{
		smtpService: smtpService,
	}
}

// HandleSMTPConnect handles SMTP connection check requests
func (h *SMTPHandler) HandleSMTPConnect(w http.ResponseWriter, r *http.Request) {
	host := r.PathValue("host")

	if host == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Host path parameter is required", // Updated error message
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Default port is 25 if not specified
	port := 25
	portStr := r.URL.Query().Get("port")
	if portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.APIError{
				Error:   "Invalid port parameter",
				Code:    http.StatusBadRequest,
				Details: err.Error(),
			})
			return
		}
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

	// Use the SMTP service through the port interface
	result, err := h.smtpService.TestSMTPConnection(r.Context(), host, port, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "SMTP connection check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert domain result to API response
	response := models.FromSMTPConnectionResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSMTPStartTLS handles SMTP STARTTLS check requests
func (h *SMTPHandler) HandleSMTPStartTLS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	host := vars["host"]

	if host == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Host path parameter is required", // Updated error message
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

	// Use the SMTP service through the port interface
	result, err := h.smtpService.CheckSMTP(r.Context(), host, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "SMTP STARTTLS check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Get only the STARTTLS-related information from the result
	response := models.FromSMTPStartTLSResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSMTPRelayTest handles SMTP open relay test requests
func (h *SMTPHandler) HandleSMTPRelayTest(w http.ResponseWriter, r *http.Request) {
	// Read and parse request body
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

	var req models.SMTPRelayTestRequest
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
	validationResult := apivalidation.ValidateSMTPRelayTestRequest(&req)
	if !validationResult.Valid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Validation failed",
			"code":        http.StatusBadRequest,
			"validations": validationResult.Errors,
		})
		return
	}

	// Use the SMTP service through the port interface
	// Note: Currently, there's no direct method for relay testing in the port
	// This would require extending the SMTPPort interface or using a specialized service
	result, err := h.smtpService.CheckSMTP(r.Context(), req.Host, time.Duration(req.Timeout)*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "SMTP relay test failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// For now, extract relay test info from the comprehensive result
	// In a real implementation, you would have a dedicated relay test method
	response := models.FromSMTPRelayTestResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSMTPCheck handles SMTP check requests
func (h *SMTPHandler) HandleSMTPCheck(w http.ResponseWriter, r *http.Request) {
	// Read and parse request body
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
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Invalid JSON format",
			Code:    http.StatusBadRequest,
			Details: err.Error(),
		})
		return
	}

	// Validate the request
	validationResult := apivalidation.ValidateSMTPRequest(&req)
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

	// Use the SMTP service through the port interface
	result, err := h.smtpService.CheckSMTP(r.Context(), req.Target, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "SMTP check failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert domain result to API response
	response := models.FromSMTPResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
