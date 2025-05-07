// Package errors provides error handling utilities for the API
package errors

import (
	"encoding/json"
	"mxclone/pkg/logging"
	"net/http"
)

// ErrorType defines the type of API error
type ErrorType string

// Error types
const (
	ValidationError ErrorType = "validation_error"
	NotFoundError   ErrorType = "not_found_error"
	BadRequestError ErrorType = "bad_request_error"
	InternalError   ErrorType = "internal_error"
	ServiceError    ErrorType = "service_error"
	AuthError       ErrorType = "auth_error"
	RateLimitError  ErrorType = "rate_limit_error"
	TimeoutError    ErrorType = "timeout_error"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Error       string      `json:"error"`
	Code        int         `json:"code,omitempty"`
	Type        ErrorType   `json:"type,omitempty"`
	Details     string      `json:"details,omitempty"`
	Validations interface{} `json:"validations,omitempty"`
}

// ErrorHandler manages API error responses
type ErrorHandler struct {
	logger *logging.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logging.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleValidationError handles validation errors
func (h *ErrorHandler) HandleValidationError(w http.ResponseWriter, message string, validations interface{}) {
	h.logger.Info("Validation error: %s", message)
	h.sendError(w, ErrorResponse{
		Error:       "Validation failed",
		Code:        http.StatusBadRequest,
		Type:        ValidationError,
		Details:     message,
		Validations: validations,
	}, http.StatusBadRequest)
}

// HandleBadRequest handles bad request errors
func (h *ErrorHandler) HandleBadRequest(w http.ResponseWriter, message string) {
	h.logger.Info("Bad request: %s", message)
	h.sendError(w, ErrorResponse{
		Error:   "Bad request",
		Code:    http.StatusBadRequest,
		Type:    BadRequestError,
		Details: message,
	}, http.StatusBadRequest)
}

// HandleNotFound handles not found errors
func (h *ErrorHandler) HandleNotFound(w http.ResponseWriter, message string) {
	h.logger.Info("Resource not found: %s", message)
	h.sendError(w, ErrorResponse{
		Error:   "Not found",
		Code:    http.StatusNotFound,
		Type:    NotFoundError,
		Details: message,
	}, http.StatusNotFound)
}

// HandleInternalError handles internal server errors
func (h *ErrorHandler) HandleInternalError(w http.ResponseWriter, err error) {
	h.logger.Error("Internal server error: %v", err)
	h.sendError(w, ErrorResponse{
		Error:   "Internal server error",
		Code:    http.StatusInternalServerError,
		Type:    InternalError,
		Details: "An unexpected error occurred",
	}, http.StatusInternalServerError)
}

// HandleServiceError handles errors from underlying services
func (h *ErrorHandler) HandleServiceError(w http.ResponseWriter, err error, details string) {
	h.logger.Error("Service error: %v - %s", err, details)
	h.sendError(w, ErrorResponse{
		Error:   "Service error",
		Code:    http.StatusInternalServerError,
		Type:    ServiceError,
		Details: details,
	}, http.StatusInternalServerError)
}

// HandleTimeoutError handles timeout errors
func (h *ErrorHandler) HandleTimeoutError(w http.ResponseWriter, service string) {
	h.logger.Error("Service timeout: %s", service)
	h.sendError(w, ErrorResponse{
		Error:   "Service timeout",
		Code:    http.StatusGatewayTimeout,
		Type:    TimeoutError,
		Details: "The request to " + service + " timed out",
	}, http.StatusGatewayTimeout)
}

// HandleRateLimitError handles rate limit errors
func (h *ErrorHandler) HandleRateLimitError(w http.ResponseWriter) {
	h.logger.Info("Rate limit exceeded")
	h.sendError(w, ErrorResponse{
		Error:   "Rate limit exceeded",
		Code:    http.StatusTooManyRequests,
		Type:    RateLimitError,
		Details: "Too many requests, please try again later",
	}, http.StatusTooManyRequests)
}

// sendError sends the error response to the client
func (h *ErrorHandler) sendError(w http.ResponseWriter, err ErrorResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		h.logger.Error("Failed to encode error response: %v", err)
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(errorMsg string, code int, errType ErrorType, details string) ErrorResponse {
	return ErrorResponse{
		Error:   errorMsg,
		Code:    code,
		Type:    errType,
		Details: details,
	}
}
