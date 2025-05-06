// Package secondary contains the secondary adapters (implementing output ports)
package secondary

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"mxclone/domain/networktools"
)

// NetworkToolsRepository implements the NetworkTools repository output port
type NetworkToolsRepository struct {
	// Configuration options could be injected here
}

// NewNetworkToolsRepository creates a new NetworkTools repository
func NewNetworkToolsRepository() *NetworkToolsRepository {
	return &NetworkToolsRepository{}
}

// ExecutePingCommand executes the actual ping command/operation
func (r *NetworkToolsRepository) ExecutePingCommand(ctx context.Context, target string, count int, timeout time.Duration) (string, []time.Duration, int, int, error) {
	// Prepare ping command
	// Note: parameters may vary by OS (Linux vs macOS vs Windows)
	// This implementation assumes a Linux/Unix-like system
	cmd := exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(count), "-W", strconv.Itoa(int(timeout.Seconds())), target)

	// Execute ping command
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// If the command failed but we have output, we might still be able to parse it
		if len(outputStr) == 0 {
			return "", nil, count, 0, err
		}
	}

	// Parse ping output to extract RTT values
	rtts := make([]time.Duration, 0)
	packetsSent := count
	packetsReceived := 0

	// Parse each line of ping output
	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	for scanner.Scan() {
		line := scanner.Text()

		// Look for lines like "64 bytes from 8.8.8.8: icmp_seq=1 ttl=116 time=15.6 ms"
		if strings.Contains(line, "bytes from") && strings.Contains(line, "time=") {
			packetsReceived++

			// Extract round-trip time
			rttStr := regexp.MustCompile(`time=([0-9.]+) ms`).FindStringSubmatch(line)
			if len(rttStr) > 1 {
				rttValue, parseErr := strconv.ParseFloat(rttStr[1], 64)
				if parseErr == nil {
					rtts = append(rtts, time.Duration(rttValue*float64(time.Millisecond)))
				}
			}
		}

		// Look for packet statistics summary
		if strings.Contains(line, "packets transmitted") {
			// Try to extract more accurate sent/received counts
			parts := strings.Split(line, ", ")
			if len(parts) >= 2 {
				sentParts := strings.Split(parts[0], " ")
				if len(sentParts) >= 1 {
					if sent, parseErr := strconv.Atoi(sentParts[0]); parseErr == nil {
						packetsSent = sent
					}
				}

				recvParts := strings.Split(parts[1], " ")
				if len(recvParts) >= 1 {
					if recv, parseErr := strconv.Atoi(recvParts[0]); parseErr == nil {
						packetsReceived = recv
					}
				}
			}
		}
	}

	return outputStr, rtts, packetsSent, packetsReceived, err
}

// ResolveDomain resolves a domain name to an IP address
func (r *NetworkToolsRepository) ResolveDomain(ctx context.Context, domain string) (string, error) {
	// Check if the input is already an IP address
	if net.ParseIP(domain) != nil {
		return domain, nil
	}

	// Create a resolver with a timeout
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	// Resolve the domain name
	ips, err := resolver.LookupHost(ctx, domain)
	if err != nil {
		return "", err
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for domain: %s", domain)
	}

	return ips[0], nil
}

// ExecuteTracerouteCommand executes the actual traceroute command/operation
func (r *NetworkToolsRepository) ExecuteTracerouteCommand(ctx context.Context, target string, maxHops int, timeout time.Duration) (string, []networktools.TracerouteHop, bool, error) {
	// Prepare traceroute command
	// Note: parameters may vary by OS
	cmd := exec.CommandContext(ctx, "traceroute", "-m", strconv.Itoa(maxHops), "-w", strconv.Itoa(int(timeout.Seconds())), target)

	// Execute traceroute command
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// If the command failed but we have output, we might still be able to parse it
		if len(outputStr) == 0 {
			return "", nil, false, err
		}
	}

	// Parse traceroute output to extract hops
	hops := make([]networktools.TracerouteHop, 0)
	targetReached := false

	// Parse each line of traceroute output
	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip header line
		if strings.HasPrefix(line, "traceroute to") {
			continue
		}

		// Parse hop line
		// Example: " 1  router.local (192.168.1.1)  1.123 ms  0.987 ms  0.897 ms"
		hopMatch := regexp.MustCompile(`^\s*(\d+)\s+([\w.-]+|\*)\s+(?:\(([\d.]+)\))?\s+(.+)$`).FindStringSubmatch(line)
		if len(hopMatch) >= 4 {
			hopNumber, _ := strconv.Atoi(hopMatch[1])
			hostname := hopMatch[2]
			ipAddress := hopMatch[3]

			// Handle case where hostname is "*" but IP is present
			if hostname == "*" && ipAddress != "" {
				hostname = ""
			}

			// Handle case where IP is not in parentheses
			if ipAddress == "" && net.ParseIP(hostname) != nil {
				ipAddress = hostname
				hostname = ""
			}

			// Extract RTT from the first timing
			rttStr := regexp.MustCompile(`([\d.]+)\s*ms`).FindStringSubmatch(hopMatch[4])
			var rtt time.Duration
			if len(rttStr) > 1 {
				rttValue, parseErr := strconv.ParseFloat(rttStr[1], 64)
				if parseErr == nil {
					rtt = time.Duration(rttValue * float64(time.Millisecond))
				}
			}

			// Try to resolve hostname if we only have IP
			if hostname == "" && ipAddress != "" {
				resolvedHostname, _ := r.ResolveIP(ctx, ipAddress)
				hostname = resolvedHostname
			}

			hop := networktools.TracerouteHop{
				Number:   hopNumber,
				IP:       ipAddress,
				Hostname: hostname,
				RTT:      rtt,
			}

			hops = append(hops, hop)

			// Check if this hop is the target
			resolvedIP, _ := r.ResolveDomain(ctx, target)
			if ipAddress == resolvedIP || hostname == target {
				targetReached = true
			}
		}
	}

	return outputStr, hops, targetReached, err
}

// ResolveIP resolves an IP address to a hostname
func (r *NetworkToolsRepository) ResolveIP(ctx context.Context, ip string) (string, error) {
	// Create a resolver with a timeout
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 2 * time.Second,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	// Try to do a reverse lookup
	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil {
		return "", err
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no hostname found for IP: %s", ip)
	}

	// Remove trailing dot from hostname
	return strings.TrimSuffix(names[0], "."), nil
}

// ExecuteWHOISCommand executes the actual WHOIS command/operation
func (r *NetworkToolsRepository) ExecuteWHOISCommand(ctx context.Context, target string, timeout time.Duration) (string, error) {
	// Set up a timeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute WHOIS command
	cmd := exec.CommandContext(ctxWithTimeout, "whois", target)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// ParseWHOISData parses raw WHOIS data to extract structured information
func (r *NetworkToolsRepository) ParseWHOISData(rawData string) (string, string, string, []string, error) {
	var registrar string
	var createdDate string
	var expirationDate string
	nameServers := make([]string, 0)

	// Parse each line of the WHOIS output
	scanner := bufio.NewScanner(strings.NewReader(rawData))
	for scanner.Scan() {
		line := scanner.Text()

		// Extract registrar information
		if strings.Contains(line, "Registrar:") || strings.Contains(line, "Registrar URL:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 && registrar == "" {
				registrar = strings.TrimSpace(parts[1])
			}
		}

		// Extract creation date
		if strings.Contains(line, "Creation Date:") ||
			strings.Contains(line, "Created:") ||
			strings.Contains(line, "Registration Date:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 && createdDate == "" {
				createdDate = strings.TrimSpace(parts[1])
			}
		}

		// Extract expiration date
		if strings.Contains(line, "Expiration Date:") ||
			strings.Contains(line, "Expiry Date:") ||
			strings.Contains(line, "Registry Expiry Date:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 && expirationDate == "" {
				expirationDate = strings.TrimSpace(parts[1])
			}
		}

		// Extract name servers
		if strings.Contains(line, "Name Server:") || strings.Contains(line, "NS:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				ns := strings.TrimSpace(parts[1])
				if ns != "" {
					nameServers = append(nameServers, strings.ToLower(ns))
				}
			}
		}
	}

	return registrar, createdDate, expirationDate, nameServers, nil
}
