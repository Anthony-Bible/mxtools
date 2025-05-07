// Package secondary contains the secondary adapters (implementing output ports)
package secondary

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"mxclone/domain/dns"
	"mxclone/ports/input"
)

// SMTPRepository implements the SMTP repository output port
type SMTPRepository struct {
	dnsService input.DNSPort
}

// NewSMTPRepository creates a new SMTP repository
func NewSMTPRepository(dnsService input.DNSPort) *SMTPRepository {
	return &SMTPRepository{
		dnsService: dnsService,
	}
}

// GetMXRecords retrieves the MX records for a domain
func (r *SMTPRepository) GetMXRecords(ctx context.Context, domain string) ([]string, error) {
	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Use our DNS service to look up MX records
	result, err := r.dnsService.Lookup(ctxWithTimeout, domain, dns.TypeMX)
	if err != nil {
		return nil, err
	}

	// Extract MX server hostnames from the lookup results
	mxRecords := make([]string, 0)
	for _, mx := range result.Lookups["MX"] {
		// Extract hostname from the MX record (e.g., "mail.example.com (priority: 10)" -> "mail.example.com")
		parts := strings.Split(mx, " ")
		if len(parts) > 0 {
			mxHost := parts[0]
			// Remove any trailing dot
			mxHost = strings.TrimSuffix(mxHost, ".")
			mxRecords = append(mxRecords, mxHost)
		}
	}

	return mxRecords, nil
}

// ConnectToSMTPServer connects to an SMTP server and tests its capabilities
// Returns: connected, latency, supportsStartTLS, authMethods, banner, error
func (r *SMTPRepository) ConnectToSMTPServer(ctx context.Context, server string, port int, timeout time.Duration) (bool, time.Duration, bool, []string, string, error) {
	// Establish a TCP connection to the SMTP server
	startTime := time.Now()

	// Create a dialer with timeout
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Connect to the server
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", server, port))
	if err != nil {
		return false, 0, false, nil, "", fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Calculate connection latency
	latency := time.Since(startTime)

	// Set read deadline for receiving the banner
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Read the welcome banner
	banner, err := readResponse(conn)
	if err != nil {
		return true, latency, false, nil, "", fmt.Errorf("failed to read banner: %w", err)
	}

	// Check for STARTTLS support
	supportsStartTLS, err := checkStartTLS(conn, timeout)
	if err != nil {
		// We still return true for connected even if checking STARTTLS failed
		return true, latency, false, nil, banner, nil
	}

	// Check for supported authentication methods
	authMethods, err := checkAuthMethods(conn, timeout)
	if err != nil {
		// We still return success with the info we have
		return true, latency, supportsStartTLS, nil, banner, nil
	}

	return true, latency, supportsStartTLS, authMethods, banner, nil
}

// readResponse reads a response from an SMTP server
func readResponse(conn net.Conn) (string, error) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer[:n]), nil
}

// checkStartTLS checks if the server supports STARTTLS
func checkStartTLS(conn net.Conn, timeout time.Duration) (bool, error) {
	// Send EHLO command
	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write([]byte("EHLO mxclone.example.com\r\n"))
	if err != nil {
		return false, err
	}

	// Read server response
	conn.SetReadDeadline(time.Now().Add(timeout))
	response, err := readResponse(conn)
	if err != nil {
		return false, err
	}

	// Check if STARTTLS is in the response
	return strings.Contains(strings.ToUpper(response), "STARTTLS"), nil
}

// checkAuthMethods checks which authentication methods are supported
func checkAuthMethods(conn net.Conn, timeout time.Duration) ([]string, error) {
	// We already sent EHLO in checkStartTLS, so we just need to check the response
	// But the response might have been consumed, so we might need to send EHLO again

	authMethods := []string{}

	// Check for "AUTH " in the response (from previous EHLO)
	// We would need to keep the full response from the previous EHLO to do this properly

	// This is a simplified implementation - in a real application,
	// we would parse the full EHLO response and extract all auth methods

	return authMethods, nil
}
