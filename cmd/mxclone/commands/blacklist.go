// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"mxclone/pkg/validation"
)

// BlacklistCmd represents the blacklist command
var BlacklistCmd = &cobra.Command{
	Use:   "blacklist [ip]",
	Short: "Check if an IP is blacklisted",
	Long: `Check if an IP address is listed on various DNS-based blacklists (DNSBLs).
This helps determine if an IP address has a poor reputation for sending spam or engaging in malicious activities.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate IP address
		ip := args[0]
		ip = validation.SanitizeIP(ip)
		if err := validation.ValidateIP(ip); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Get command flags
		all, _ := cmd.Flags().GetBool("all")
		timeout, _ := cmd.Flags().GetInt("timeout")
		checkHealth, _ := cmd.Flags().GetBool("check-health")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Checking if IP %s is blacklisted...\n", ip)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Get the list of blacklist zones to check
		// Default blacklist zones
		defaultZones := []string{
			"bl.spamcop.net",
			"dnsbl.sorbs.net",
		}

		var zones []string
		if all {
			// Use all configured zones
			zones = defaultZones
		} else {
			// Use only the default zones (first 3 or fewer)
			maxZones := 3
			if len(defaultZones) < maxZones {
				maxZones = len(defaultZones)
			}
			zones = defaultZones[:maxZones]
		}

		// Get the DNSBL service from the dependency injection container
		dnsblService := Container.GetDNSBLService()

		// Check if we should check the health of the blacklists first
		if checkHealth {
			fmt.Println("Checking health of blacklist servers...")
			healthStatus := dnsblService.CheckMultipleDNSBLHealth(ctx, zones, timeoutDuration)

			// Filter out unhealthy zones
			var healthyZones []string
			for zone, healthy := range healthStatus {
				if healthy {
					healthyZones = append(healthyZones, zone)
				} else {
					fmt.Printf("Warning: Blacklist %s appears to be unavailable\n", zone)
				}
			}
			zones = healthyZones
		}

		// Check the IP against the blacklists
		result, err := dnsblService.CheckMultipleBlacklists(ctx, ip, zones, timeoutDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking blacklists: %v\n", err)
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
			summary := dnsblService.GetBlacklistSummary(result)
			fmt.Println(summary)
		}
	},
}

func init() {
	BlacklistCmd.Flags().BoolP("all", "a", false, "Check all available blacklists")
	BlacklistCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for each blacklist check")
	BlacklistCmd.Flags().BoolP("check-health", "c", false, "Check health of blacklist servers before querying")
}
