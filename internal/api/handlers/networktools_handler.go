package handlers

import (
	"encoding/json"
	"io"
	"mxclone/domain/networktools"
	"mxclone/internal/api/models"
	apivalidation "mxclone/internal/api/validation"
	"mxclone/ports/input"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// NetworkToolsHandler encapsulates handlers for network diagnostic tools
type NetworkToolsHandler struct {
	networkToolsService input.NetworkToolsPort // Using the interface (port) instead of direct implementation
}

// NewNetworkToolsHandler creates a new network tools handler with the given service
func NewNetworkToolsHandler(networkToolsService input.NetworkToolsPort) *NetworkToolsHandler {
	return &NetworkToolsHandler{
		networkToolsService: networkToolsService,
	}
}

// HandlePing handles ping requests
func (h *NetworkToolsHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	// Extract host from URL path parameter set by the router
	host, ok := r.Context().Value("host").(string)
	// Fallback to query parameter if not found in context
	if !ok || host == "" {
		host = r.URL.Query().Get("host")
	}

	if host == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Host parameter is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Get optional count parameter, default to 4
	count := 4
	countStr := r.URL.Query().Get("count")
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count < 1 || count > 20 {
			count = 4 // Reset to default for invalid input
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

	// Use the network tools service through the port interface
	result, err := h.networkToolsService.ExecutePing(r.Context(), host, count, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Ping failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Wrap the ping result in a network tool result
	genericResult := h.networkToolsService.WrapResult(networktools.ToolTypePing, result, nil, nil, nil)

	// Convert result to API response
	response := models.FromNetworkToolResult(genericResult)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleTraceroute handles traceroute requests
func (h *NetworkToolsHandler) HandleTraceroute(w http.ResponseWriter, r *http.Request) {
	// Extract host from URL path parameter set by the router
	host, ok := r.Context().Value("host").(string)
	// Fallback to query parameter if not found in context
	if !ok || host == "" {
		host = r.URL.Query().Get("host")
	}

	if host == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Host parameter is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Get optional maxHops parameter, default to 30
	maxHops := 30
	maxHopsStr := r.URL.Query().Get("maxHops")
	if maxHopsStr != "" {
		var err error
		maxHops, err = strconv.Atoi(maxHopsStr)
		if err != nil || maxHops < 1 || maxHops > 64 {
			maxHops = 30 // Reset to default for invalid input
		}
	}

	// Default timeout is 30 seconds
	timeout := 30 * time.Second
	timeoutStr := r.URL.Query().Get("timeout")
	if timeoutStr != "" {
		var err error
		timeoutDuration, err := time.ParseDuration(timeoutStr)
		if err == nil {
			timeout = timeoutDuration
		}
	}

	// Use the network tools service through the port interface
	result, err := h.networkToolsService.ExecuteTraceroute(r.Context(), host, maxHops, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Traceroute failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Wrap the traceroute result in a network tool result
	genericResult := h.networkToolsService.WrapResult(networktools.ToolTypeTraceroute, nil, result, nil, nil)

	// Convert result to API response
	response := models.FromNetworkToolResult(genericResult)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleWhois handles WHOIS lookup requests
func (h *NetworkToolsHandler) HandleWhois(w http.ResponseWriter, r *http.Request) {
	// Extract domain from URL path parameter set by the router
	domain, ok := r.Context().Value("domain").(string)
	// Fallback to query parameter if not found in context
	if !ok || domain == "" {
		domain = r.URL.Query().Get("domain")
	}

	if domain == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Domain parameter is required",
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

	// Use the network tools service through the port interface
	result, err := h.networkToolsService.ExecuteWHOIS(r.Context(), domain, timeout)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "WHOIS lookup failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Wrap the WHOIS result in a network tool result
	genericResult := h.networkToolsService.WrapResult(networktools.ToolTypeWHOIS, nil, nil, result, nil)

	// Convert result to API response
	response := models.FromNetworkToolResult(genericResult)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleNetworkTools handles legacy network tools requests
func (h *NetworkToolsHandler) HandleNetworkTools(w http.ResponseWriter, r *http.Request) {
	// Read and parse request body
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

	// Determine which tool to use (default to ping if not specified)
	toolType := strings.ToLower(req.Option)
	if toolType == "" {
		toolType = "ping"
	}

	// Validate the request
	validationResult := apivalidation.ValidateNetworkToolRequest(&req, toolType)
	if !validationResult.Valid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "Validation failed",
			"code":        http.StatusBadRequest,
			"validations": validationResult.Errors,
		})
		return
	}

	// Set up options for the tool
	options := make(map[string]interface{})
	options["timeout"] = 10 * time.Second // Default timeout

	// Add tool-specific options
	switch toolType {
	case "ping":
		options["count"] = 4
	case "traceroute":
		options["maxHops"] = 30
	}

	// Parse query parameters for options
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			switch key {
			case "timeout":
				if timeout, err := time.ParseDuration(values[0]); err == nil {
					options["timeout"] = timeout
				}
			case "count":
				if count, err := strconv.Atoi(values[0]); err == nil && count > 0 {
					options["count"] = count
				}
			case "maxHops":
				if maxHops, err := strconv.Atoi(values[0]); err == nil && maxHops > 0 {
					options["maxHops"] = maxHops
				}
			}
		}
	}

	// Determine tool type enum
	var toolTypeEnum networktools.ToolType
	switch toolType {
	case "ping":
		toolTypeEnum = networktools.ToolTypePing
	case "traceroute":
		toolTypeEnum = networktools.ToolTypeTraceroute
	case "whois":
		toolTypeEnum = networktools.ToolTypeWHOIS
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIError{
			Error: "Invalid tool type",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Use the network tools service through the port interface
	result, err := h.networkToolsService.ExecuteNetworkTool(r.Context(), toolTypeEnum, req.Target, options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIError{
			Error:   "Network tool execution failed",
			Code:    http.StatusInternalServerError,
			Details: err.Error(),
		})
		return
	}

	// Convert domain result to API response
	response := models.FromNetworkToolResult(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
