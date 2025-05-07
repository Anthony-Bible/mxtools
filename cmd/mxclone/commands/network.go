// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"mxclone/domain/networktools"
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

		// Get the network tools service from the dependency injection container
		networkService := Container.GetNetworkToolsService()

		// Perform the ping
		result, err := networkService.ExecutePing(ctx, host, count, timeoutDuration)
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
			pingResult := networkService.WrapResult(networktools.ToolTypePing, result, nil, nil, nil)
			summary := networkService.FormatToolResult(pingResult)
			fmt.Println(summary)
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

		// Get the network tools service from the dependency injection container
		networkService := Container.GetNetworkToolsService()

		// Perform the traceroute
		result, err := networkService.ExecuteTraceroute(ctx, host, maxHops, timeoutDuration)
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
			tracerouteResult := networkService.WrapResult(networktools.ToolTypeTraceroute, nil, result, nil, nil)
			summary := networkService.FormatToolResult(tracerouteResult)
			fmt.Println(summary)
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

		// Get command flags
		timeout, _ := cmd.Flags().GetInt("timeout")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Looking up WHOIS information for %s...\n", query)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Get the network tools service from the dependency injection container
		networkService := Container.GetNetworkToolsService()

		// Perform the WHOIS lookup
		result, err := networkService.ExecuteWHOIS(ctx, query, timeoutDuration)
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
			whoisResult := networkService.WrapResult(networktools.ToolTypeWHOIS, nil, nil, result, nil)
			summary := networkService.FormatToolResult(whoisResult)
			fmt.Println(summary)
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
	WhoisCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds")

	// Add the network command to the root command
	rootCmd.AddCommand(NetworkCmd)
}
