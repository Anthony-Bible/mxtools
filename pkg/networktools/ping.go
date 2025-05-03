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

const (
	// ProtocolICMP is the ICMP protocol number
	ProtocolICMP = 1
	// ProtocolIPv6ICMP is the IPv6 ICMP protocol number
	ProtocolIPv6ICMP      = 58
	UDPTracerouteBasePort = 33434
)

// PingResult represents the result of a ping operation.
type PingResult struct {
	Target       string        `json:"target"`
	Sent         int           `json:"sent"`
	Received     int           `json:"received"`
	PacketLoss   float64       `json:"packetLoss"`
	MinRTT       time.Duration `json:"minRtt,omitempty"`
	MaxRTT       time.Duration `json:"maxRtt,omitempty"`
	AvgRTT       time.Duration `json:"avgRtt,omitempty"`
	Error        string        `json:"error,omitempty"`
	IsPrivileged bool          `json:"isPrivileged"`
}

// Ping sends ICMP echo requests to the specified target.
func Ping(ctx context.Context, target string, count int, timeout time.Duration) (*PingResult, error) {
	result := &PingResult{
		Target: target,
		Sent:   count,
	}

	// Resolve the target to an IP address
	ipAddr, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to resolve target: %s", err.Error())
		return result, err
	}
	// Prepare dstAddr as the target address for WriteTo (type net.Addr)
	var dstAddr net.Addr = ipAddr

	// Determine if we're dealing with IPv4 or IPv6
	var network string
	if ipAddr.IP.To4() != nil {
		network = "ip4:icmp"
	} else {
		network = "ip6:ipv6-icmp"
	}

	// Try to create a privileged ICMP connection
	conn, err := icmp.ListenPacket(network, "")
	if err != nil {
		// If we can't create a privileged connection, try unprivileged
		if err != nil {
			// If we can't create a privileged connection, try unprivileged
			var localAddr string
			if network == "ip4:icmp" {
				network = "udp4"
				localAddr = "0.0.0.0"                 // Use ephemeral port for IPv4
				dstAddr = &net.UDPAddr{IP: ipAddr.IP} // Use UDPAddr for UDP
			} else {
				network = "udp6"
				localAddr = "[::]:0"                  // Use ephemeral port for IPv6
				dstAddr = &net.UDPAddr{IP: ipAddr.IP} // Use UDPAddr for UDP
			}
			conn, err = icmp.ListenPacket(network, localAddr)
			if err != nil {
				result.Error = fmt.Sprintf("Failed to create ICMP connection: %s", err.Error())
				return result, err
			}
			result.IsPrivileged = false
		}
		result.IsPrivileged = false
	} else {
		result.IsPrivileged = true
	}
	defer conn.Close()

	// Set the connection deadline
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		result.Error = fmt.Sprintf("Failed to set deadline: %s", err.Error())
		return result, err
	}

	// Create the ICMP message
	var msgType icmp.Type
	if ipAddr.IP.To4() != nil {
		msgType = ipv4.ICMPTypeEcho
	} else {
		msgType = ipv6.ICMPTypeEchoRequest
	}

	// Get the current process ID for the ICMP identifier
	pid := os.Getpid() & 0xffff

	// Variables to track statistics
	var minRTT, maxRTT, totalRTT time.Duration
	received := 0

	// Send the specified number of pings
	for i := 0; i < count; i++ {
		// Create the ICMP message
		msg := icmp.Message{
			Type: msgType,
			Code: 0,
			Body: &icmp.Echo{
				ID:   pid,
				Seq:  i,
				Data: []byte("PING"),
			},
		}

		// Marshal the message
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to marshal ICMP message: %s", err.Error())
			return result, err
		}

		// Record the start time
		start := time.Now()

		// Send the message
		_, err = conn.WriteTo(msgBytes, dstAddr)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to send ICMP message: %s", err.Error())
			return result, err
		}

		// Create a buffer to receive the response
		reply := make([]byte, 1500)

		// Wait for the response
		n, _, err := conn.ReadFrom(reply)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout, continue to the next ping
				continue
			}
			result.Error = fmt.Sprintf("Failed to receive ICMP message: %s", err.Error())
			return result, err
		}
		// print reply

		// Calculate the round-trip time
		rtt := time.Since(start)

		// Parse the ICMP message
		var proto int
		if ipAddr.IP.To4() != nil {
			proto = ProtocolICMP
		} else {
			proto = ProtocolIPv6ICMP
		}

		parsed, err := icmp.ParseMessage(proto, reply[:n])
		if err != nil {
			result.Error = fmt.Sprintf("Failed to parse ICMP message: %s", err.Error())
			return result, err
		}

		// Check if it's an echo reply
		switch parsed.Type {
		case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
			echo, ok := parsed.Body.(*icmp.Echo)
			if !ok {
				continue
			}
			// Check if it's a reply to our request
			if (!result.IsPrivileged || echo.ID == pid) && echo.Seq == i {
				received++
				totalRTT += rtt
				if minRTT == 0 || rtt < minRTT {
					minRTT = rtt
				}
				if rtt > maxRTT {
					maxRTT = rtt
				}
			}
		}

		// Sleep a bit before sending the next ping
		time.Sleep(time.Second)
	}

	// Calculate statistics
	result.Received = received
	if count > 0 {
		result.PacketLoss = 100 - (float64(received) / float64(count) * 100)
	}
	if received > 0 {
		result.MinRTT = minRTT
		result.MaxRTT = maxRTT
		result.AvgRTT = totalRTT / time.Duration(received)
	}

	return result, nil
}

// PingWithPrivilegeCheck performs a ping operation and handles privilege requirements.
func PingWithPrivilegeCheck(ctx context.Context, target string, count int, timeout time.Duration) (*PingResult, error) {
	result, err := Ping(ctx, target, count, timeout)
	if err != nil && !result.IsPrivileged {
		// If the ping failed and we're not privileged, return a more helpful error
		result.Error = fmt.Sprintf("Ping requires elevated privileges. Try running as root or with sudo. Error: %s", result.Error)
	}
	return result, err
}

// FormatPingResult formats a ping result as a string.
func FormatPingResult(result *PingResult) string {
	if result.Error != "" {
		return fmt.Sprintf("Ping to %s failed: %s", result.Target, result.Error)
	}

	output := fmt.Sprintf("PING %s\n", result.Target)
	output += fmt.Sprintf("Sent: %d, Received: %d, Packet Loss: %.1f%%\n", result.Sent, result.Received, result.PacketLoss)
	if result.Received > 0 {
		output += fmt.Sprintf("Round-trip min/avg/max: %s/%s/%s\n", result.MinRTT, result.AvgRTT, result.MaxRTT)
	}
	if !result.IsPrivileged {
		output += "Note: Running in unprivileged mode. For better results, run with elevated privileges.\n"
	}
	return output
}
