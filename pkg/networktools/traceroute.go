// Package networktools provides auxiliary network diagnostic tools.
package networktools

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// TracerouteHop represents a single hop in a traceroute.
type TracerouteHop struct {
	TTL      int           `json:"ttl"`
	Address  string        `json:"address,omitempty"`
	Hostname string        `json:"hostname,omitempty"`
	RTT      time.Duration `json:"rtt,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// TracerouteResult represents the result of a traceroute operation.
type TracerouteResult struct {
	Target       string          `json:"target"`
	Hops         []TracerouteHop `json:"hops"`
	Error        string          `json:"error,omitempty"`
	IsPrivileged bool            `json:"isPrivileged"`
}

// Traceroute performs a traceroute to the specified target.
func Traceroute(ctx context.Context, target string, maxHops int, timeout time.Duration) (*TracerouteResult, error) {
	result := &TracerouteResult{
		Target: target,
		Hops:   make([]TracerouteHop, 0, maxHops),
	}

	// Resolve the target to an IP address
	ipAddr, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to resolve target: %s", err.Error())
		return result, err
	}
	var (
		network string
		proto   int
		isIPv4  bool
	)
	if ipAddr.IP.To4() != nil {
		network = "ip4:icmp"
		proto = ProtocolICMP
		isIPv4 = true
	} else {
		network = "ip6:ipv6-icmp"
		proto = ProtocolIPv6ICMP
		isIPv4 = false
	}

	// Try to create a privileged ICMP connection
	conn, err := icmp.ListenPacket(network, "")
	isPrivileged := true
	if err != nil {
		// Fallback to UDP for non-privileged mode
		//var localAddr string
		if isIPv4 {
			network = "udp4"
			//localAddr = "0.0.0.0"

		} else {
			network = "udp6"
			//localAddr = "[::]:0"
		}
		conn, err = icmp.ListenPacket(network, "")
		if err != nil {
			result.Error = fmt.Sprintf("Failed to create ICMP/UDP connection: %s", err.Error())
			return result, err
		}
		isPrivileged = false
	}
	result.IsPrivileged = isPrivileged
	defer conn.Close()

	// Create the ICMP message
	var msgType icmp.Type
	if isIPv4 {
		msgType = ipv4.ICMPTypeEcho
	} else {
		msgType = ipv6.ICMPTypeEchoRequest
	}

	// Perform the traceroute
	for ttl := 1; ttl <= maxHops; ttl++ {
		hop := TracerouteHop{TTL: ttl}

		// Set TTL/hop limit
		if isIPv4 {
			err = conn.IPv4PacketConn().SetTTL(ttl)
		} else {
			err = conn.IPv6PacketConn().SetHopLimit(ttl)
		}
		if err != nil {
			hop.Error = fmt.Sprintf("Failed to set TTL: %s", err.Error())
			result.Hops = append(result.Hops, hop)
			// Stop if this hop is the final address
			if ipAddr.String() == hop.Address || ttl == maxHops {
				break
			}
			continue
		}

		// Set the connection deadline
		if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			hop.Error = fmt.Sprintf("Failed to set deadline: %s", err.Error())
			result.Hops = append(result.Hops, hop)
			// Stop if this hop is the final address
			if ipAddr.String() == hop.Address || ttl == maxHops {
				break
			}
			continue
		}

		var (
			msgBytes []byte
			sendAddr net.Addr
		)
		msg := icmp.Message{
			Type: msgType,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  ttl,
				Data: []byte("TRACEROUTE"),
			},
		}
		msgBytes, err = msg.Marshal(nil)
		if isPrivileged {
			if err != nil {
				hop.Error = fmt.Sprintf("Failed to marshal ICMP message: %s", err.Error())
				result.Hops = append(result.Hops, hop)
				// Stop if this hop is the final address
				if ipAddr.String() == hop.Address || ttl == maxHops {
					break
				}
				continue
			}
			sendAddr = ipAddr
		} else {
			// UDP traceroute: send to incrementing port
			port := UDPTracerouteBasePort + ttl
			sendAddr = &net.UDPAddr{IP: ipAddr.IP, Port: port}
		}

		// Record the start time
		start := time.Now()
		_, err = conn.WriteTo(msgBytes, sendAddr)
		if err != nil {
			hop.Error = fmt.Sprintf("Failed to send packet: %s", err.Error())
			result.Hops = append(result.Hops, hop)
			// Stop if this hop is the final address
			if ipAddr.String() == hop.Address || ttl == maxHops {
				break
			}
			continue
		}

		// Create a buffer to receive the response
		reply := make([]byte, 1500)

		// Wait for the response
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				hop.Error = "Request timed out"
				result.Hops = append(result.Hops, hop)
				// Stop if this hop is the final address
				if ipAddr.String() == hop.Address || ttl == maxHops {
					break
				}
				continue
			}
			hop.Error = fmt.Sprintf("Failed to receive reply: %s", err.Error())
			result.Hops = append(result.Hops, hop)
			// Stop if this hop is the final address
			if ipAddr.String() == hop.Address || ttl == maxHops {
				break
			}
			continue
		}
		hop.RTT = time.Since(start)

		// Get responder address
		switch addr := peer.(type) {
		case *net.IPAddr:
			hop.Address = addr.String()
		case *net.UDPAddr:
			hop.Address = addr.IP.String()
		default:
			hop.Address = peer.String()
		}

		names, err := net.LookupAddr(hop.Address)
		if err == nil && len(names) > 0 {
			hop.Hostname = names[0]
		}

		parsed, err := icmp.ParseMessage(proto, reply[:n])
		if err != nil {
			hop.Error = fmt.Sprintf("Failed to parse ICMP message: %s", err.Error())
			result.Hops = append(result.Hops, hop)
			// Stop if this hop is the final address
			if hop.Address == ipAddr.String() || ttl == maxHops {
				break
			}
			continue
		}

		// Privileged: stop on EchoReply. Unprivileged: stop on Port Unreachable.
		if isPrivileged {
			if (isIPv4 && parsed.Type == ipv4.ICMPTypeEchoReply) ||
				(!isIPv4 && parsed.Type == ipv6.ICMPTypeEchoReply) {
				result.Hops = append(result.Hops, hop)
				break
			}
		} else {
			if isIPv4 && parsed.Type == ipv4.ICMPTypeDestinationUnreachable {
				// Port unreachable means we've reached the destination
				result.Hops = append(result.Hops, hop)
				break
			}
			if !isIPv4 && parsed.Type == ipv6.ICMPTypeDestinationUnreachable {
				result.Hops = append(result.Hops, hop)
				break
			}
		}

		result.Hops = append(result.Hops, hop)
		// Stop if this hop is the final address
		if hop.Address == ipAddr.String() || ttl == maxHops {
			break
		}
	}

	return result, nil
}

// TracerouteWithPrivilegeCheck performs a traceroute operation and handles privilege requirements.
func TracerouteWithPrivilegeCheck(ctx context.Context, target string, maxHops int, timeout time.Duration) (*TracerouteResult, error) {
	result, err := Traceroute(ctx, target, maxHops, timeout)
	if err != nil && !result.IsPrivileged {
		// If the traceroute failed and we're not privileged, return a more helpful error
		result.Error = fmt.Sprintf("Traceroute requires elevated privileges. Try running as root or with sudo. Error: %s", result.Error)
	}
	return result, err
}

// FormatTracerouteResult formats a traceroute result as a string.
func FormatTracerouteResult(result *TracerouteResult) string {
	if result.Error != "" {
		return fmt.Sprintf("Traceroute to %s failed: %s", result.Target, result.Error)
	}

	output := fmt.Sprintf("Traceroute to %s\n", result.Target)
	for _, hop := range result.Hops {
		if hop.Error != "" {
			output += fmt.Sprintf("%2d  %s\n", hop.TTL, hop.Error)
		} else {
			if hop.Hostname != "" {
				output += fmt.Sprintf("%2d  %s (%s)  %s\n", hop.TTL, hop.Hostname, hop.Address, hop.RTT)
			} else {
				output += fmt.Sprintf("%2d  %s  %s\n", hop.TTL, hop.Address, hop.RTT)
			}
		}
	}
	return output
}
