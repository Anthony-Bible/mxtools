// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"mxclone/pkg/networktools"
	"mxclone/pkg/validation"
)

// NetworkCmd represents the network command
var NetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network diagnostic tools",
	Long:  `Network diagnostic tools including ping, traceroute, and whois.`,
}

// PingCmd represents the ping command
var PingCmd = &cobra.Command{
	Use:   "ping [host]",
	Short: "Ping a host",
	Long:  `Send ICMP echo requests to a host and measure response time.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate host
		host := args[0]
		if err := validation.ValidateHost(host); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Get command flags
		count, _ := cmd.Flags().GetInt("count")
		timeout, _ := cmd.Flags().GetInt("timeout")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Pinging %s...\n", host)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Perform the ping
		result, err := networktools.PingWithPrivilegeCheck(ctx, host, count, timeoutDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Output the result
		if outputFormat == "json" {
			jsonOutput, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		} else {
			// Text output
			fmt.Println(networktools.FormatPingResult(result))
		}
	},
}

// TracerouteCmd represents the traceroute command
var TracerouteCmd = &cobra.Command{
	Use:   "traceroute [host]",
	Short: "Trace the route to a host",
	Long:  `Trace the route packets take to a host.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate host
		host := args[0]
		if err := validation.ValidateHost(host); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Get command flags
		maxHops, _ := cmd.Flags().GetInt("max-hops")
		timeout, _ := cmd.Flags().GetInt("timeout")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Tracing route to %s...\n", host)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Perform the traceroute
		//result, err := networktools.TracerouteWithPrivilegeCheck(ctx, host, maxHops, timeoutDuration)
		result, err := networktools.Traceroute(ctx, host, maxHops, timeoutDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Output the result
		if outputFormat == "json" {
			jsonOutput, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		} else {
			// Text output
			fmt.Println(networktools.FormatTracerouteResult(result))
		}
	},
}

// WhoisCmd represents the whois command
var WhoisCmd = &cobra.Command{
	Use:   "whois [domain/ip]",
	Short: "Look up WHOIS information",
	Long:  `Look up WHOIS information for a domain or IP address.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate query (domain or IP)
		query := args[0]
		// Try to validate as IP first, then as domain if that fails
		ipErr := validation.ValidateIP(query)
		domainErr := validation.ValidateDomain(query)
		if ipErr != nil && domainErr != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid domain or IP address\n")
			os.Exit(1)
		}

		// Get and validate server if provided
		server, _ := cmd.Flags().GetString("server")
		if server != "" {
			if err := validation.ValidateHost(server); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		// Get other command flags
		timeout, _ := cmd.Flags().GetInt("timeout")
		followReferral, _ := cmd.Flags().GetBool("follow-referral")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Looking up WHOIS information for %s...\n", query)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		var result *networktools.WhoisResult
		var err error

		if followReferral {
			// Perform the WHOIS query with referral
			result, err = networktools.WhoisWithReferral(ctx, query, timeoutDuration)
		} else {
			// Perform the WHOIS query
			whoisServer := networktools.DefaultWhoisServer
			if server != "" {
				whoisServer.Host = server
			}
			result, err = networktools.Whois(ctx, query, whoisServer, timeoutDuration)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Output the result
		if outputFormat == "json" {
			jsonOutput, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		} else {
			// Text output
			fmt.Println(networktools.FormatWhoisResult(result))
		}
	},
}

func init() {
	// Add subcommands to the network command
	NetworkCmd.AddCommand(PingCmd)
	NetworkCmd.AddCommand(TracerouteCmd)
	NetworkCmd.AddCommand(WhoisCmd)

	// Ping command flags
	PingCmd.Flags().IntP("count", "c", 4, "Number of pings to send")
	PingCmd.Flags().IntP("timeout", "t", 5, "Timeout in seconds")

	// Traceroute command flags
	TracerouteCmd.Flags().IntP("max-hops", "m", 30, "Maximum number of hops")
	TracerouteCmd.Flags().IntP("timeout", "t", 5, "Timeout in seconds")

	// Whois command flags
	WhoisCmd.Flags().StringP("server", "s", "", "WHOIS server to query (default: whois.iana.org)")
	WhoisCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds")
	WhoisCmd.Flags().BoolP("follow-referral", "f", true, "Follow referrals to other WHOIS servers")

	// Add the network command to the root command
	rootCmd.AddCommand(NetworkCmd)
}
