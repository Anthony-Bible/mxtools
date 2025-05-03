// Package smtp provides SMTP diagnostic functionality.
package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"mxclone/pkg/dns"
	"mxclone/pkg/types"
)

// DefaultPorts is a list of default SMTP ports to check.
var DefaultPorts = []int{25, 465, 587}

// Connect establishes a connection to an SMTP server and returns the result.
func Connect(ctx context.Context, host string, port int, timeout time.Duration) (*types.SMTPResult, error) {
	result := &types.SMTPResult{}

	// Measure response time
	startTime := time.Now()

	// Create a connection with timeout
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Construct the address
	address := fmt.Sprintf("%s:%d", host, port)

	// Connect to the server
	conn, err := dialer.DialContext(ctx, "tcp", address)

	// Calculate response time
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.ConnectSuccess = false
		result.ConnectError = err.Error()
		return result, err
	}

	// Close the connection when done
	defer conn.Close()

	result.ConnectSuccess = true
	return result, nil
}

// CheckSTARTTLS checks if the SMTP server supports STARTTLS.
func CheckSTARTTLS(ctx context.Context, host string, port int, timeout time.Duration) (*types.SMTPResult, error) {
	result := &types.SMTPResult{}

	// Measure response time
	startTime := time.Now()

	// Create a connection with timeout
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Construct the address
	address := fmt.Sprintf("%s:%d", host, port)

	// Connect to the server
	conn, err := dialer.DialContext(ctx, "tcp", address)

	// Calculate response time
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.ConnectSuccess = false
		result.ConnectError = err.Error()
		return result, err
	}

	// Close the connection when done
	defer conn.Close()

	result.ConnectSuccess = true

	// Create an SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		result.ConnectSuccess = false
		result.ConnectError = err.Error()
		return result, err
	}

	// Check if the server supports STARTTLS
	supportsStartTLS := false
	if ok, _ := client.Extension("STARTTLS"); ok {
		supportsStartTLS = true

		// Try to start TLS
		tlsConfig := &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		}

		err = client.StartTLS(tlsConfig)
		if err != nil {
			supportsStartTLS = true // Still supports it, but there was an error
			result.STARTTLSError = err.Error()
		}
	}

	// Set the result
	result.SupportsSTARTTLS = &supportsStartTLS

	return result, nil
}

// CheckOpenRelay checks if the SMTP server is an open relay.
// This is a simplified check and should be used with caution.
func CheckOpenRelay(ctx context.Context, host string, port int, timeout time.Duration) (*types.SMTPResult, error) {
	result := &types.SMTPResult{}

	// Measure response time
	startTime := time.Now()

	// Create a connection with timeout
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Construct the address
	address := fmt.Sprintf("%s:%d", host, port)

	// Connect to the server
	conn, err := dialer.DialContext(ctx, "tcp", address)

	// Calculate response time
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.ConnectSuccess = false
		result.ConnectError = err.Error()
		return result, err
	}

	// Close the connection when done
	defer conn.Close()

	result.ConnectSuccess = true

	// Create an SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		result.ConnectSuccess = false
		result.ConnectError = err.Error()
		return result, err
	}

	// Try to send a test email to check if the server is an open relay
	// This is a simplified check and may not be accurate in all cases

	// Set a from address that is clearly not from the server's domain
	err = client.Mail("test@example.com")
	if err != nil {
		// Server rejected the from address, not an open relay
		isOpenRelay := false
		result.IsOpenRelay = &isOpenRelay
		return result, nil
	}

	// Set a to address that is clearly not from the server's domain
	err = client.Rcpt("test@example.org")
	if err != nil {
		// Server rejected the to address, not an open relay
		isOpenRelay := false
		result.IsOpenRelay = &isOpenRelay
		return result, nil
	}

	// If we got this far, the server might be an open relay
	// In a real implementation, we would check more thoroughly
	isOpenRelay := true
	result.IsOpenRelay = &isOpenRelay

	return result, nil
}

// VerifyPTR checks if the SMTP server has a valid PTR record.
func VerifyPTR(ctx context.Context, host string, timeout time.Duration) (bool, string, error) {
	// Resolve the host to an IP address
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, "", fmt.Errorf("failed to resolve host: %w", err)
	}

	if len(ips) == 0 {
		return false, "", fmt.Errorf("no IP addresses found for host: %s", host)
	}

	// Use the first IP address
	ip := ips[0]

	// Perform a PTR lookup
	result, err := dns.LookupWithRetry(ctx, ip.String(), "PTR", 2, timeout)
	if err != nil {
		return false, "", fmt.Errorf("PTR lookup failed: %w", err)
	}

	// Check if there are any PTR records
	if len(result.Lookups["PTR"]) == 0 {
		return false, "", nil
	}

	// Get the PTR record
	ptr := result.Lookups["PTR"][0]

	// Check if the PTR record matches the host
	// This is a simplified check and may not be accurate in all cases
	if strings.Contains(ptr, host) || strings.Contains(host, ptr) {
		return true, ptr, nil
	}

	return false, ptr, nil
}

// CheckSMTP performs a comprehensive SMTP check.
func CheckSMTP(ctx context.Context, host string, ports []int, timeout time.Duration) (*types.SMTPResult, error) {
	// If no ports are specified, use the default ports
	if len(ports) == 0 {
		ports = DefaultPorts
	}

	// Try each port until we find one that works
	for _, port := range ports {
		// Connect to the server
		result, err := Connect(ctx, host, port, timeout)
		if err != nil {
			continue
		}

		// Check STARTTLS support
		starttlsResult, err := CheckSTARTTLS(ctx, host, port, timeout)
		if err == nil && starttlsResult.SupportsSTARTTLS != nil {
			result.SupportsSTARTTLS = starttlsResult.SupportsSTARTTLS
			result.STARTTLSError = starttlsResult.STARTTLSError
		}

		// Check if the server is an open relay
		relayResult, err := CheckOpenRelay(ctx, host, port, timeout)
		if err == nil && relayResult.IsOpenRelay != nil {
			result.IsOpenRelay = relayResult.IsOpenRelay
			result.RelayCheckError = relayResult.RelayCheckError
		}

		// Return the result for the first successful connection
		return result, nil
	}

	// If we get here, we couldn't connect to any of the ports
	return &types.SMTPResult{
		ConnectSuccess: false,
		ConnectError:   "Failed to connect to any SMTP port. This may be expected for properly configured mail servers that restrict connections.",
	}, nil
}

// CheckSMTPWithPTR performs a comprehensive SMTP check including PTR verification.
func CheckSMTPWithPTR(ctx context.Context, host string, ports []int, timeout time.Duration) (*types.SMTPResult, error) {
	// Perform the basic SMTP check
	result, err := CheckSMTP(ctx, host, ports, timeout)
	if err != nil {
		return result, err
	}

	// If the connection was successful, verify the PTR record
	if result.ConnectSuccess {
		// Verify the PTR record
		ptrValid, ptrRecord, ptrErr := VerifyPTR(ctx, host, timeout)

		// Add the PTR verification result to the SMTP result
		// We'll use the Options field to store this information
		if ptrErr != nil {
			result.ConnectError = fmt.Sprintf("PTR verification error: %s", ptrErr.Error())
		} else if !ptrValid {
			result.ConnectError = fmt.Sprintf("Invalid PTR record: %s", ptrRecord)
		}
	}

	return result, nil
}
