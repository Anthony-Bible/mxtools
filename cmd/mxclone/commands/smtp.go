// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mxclone/pkg/dns"
	"mxclone/pkg/smtp"
	"mxclone/pkg/types"
	"mxclone/pkg/validation"
)

// SMTPCmd represents the smtp command
var SMTPCmd = &cobra.Command{
	Use:   "smtp [domain/ip]",
	Short: "Perform SMTP diagnostics",
	Long: `Perform SMTP diagnostics on a mail server.
This checks connectivity, STARTTLS support, open relay status, and more.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate target (domain or IP)
		target := args[0]
		// Try to validate as IP first, then as domain if that fails
		ipErr := validation.ValidateIP(target)
		domainErr := validation.ValidateDomain(target)
		if ipErr != nil && domainErr != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid domain or IP address\n")
			os.Exit(1)
		}

		// Get and validate port if provided
		port, _ := cmd.Flags().GetInt("port")
		if port != 0 {
			if err := validation.ValidatePort(port); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		// Get other command flags
		timeout, _ := cmd.Flags().GetInt("timeout")
		checkRelay, _ := cmd.Flags().GetBool("check-relay")
		checkPTR, _ := cmd.Flags().GetBool("check-ptr")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Performing SMTP diagnostics for %s...\n", target)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Determine if the target is a domain or an IP
		// If it's a domain, try to get MX records first
		var hosts []string
		mxResult, err := dns.LookupWithRetry(ctx, target, "MX", 2, timeoutDuration)
		if err == nil && len(mxResult.Lookups["MX"]) > 0 {
			// Extract the hostnames from the MX records
			for _, record := range mxResult.Lookups["MX"] {
				// Extract the hostname from the MX record (format: "hostname (priority: N)")
				hostname := record
				if idx := strings.Index(record, " (priority:"); idx > 0 {
					hostname = record[:idx]
				}
				hosts = append(hosts, hostname)
			}
			fmt.Printf("Found %d MX records for %s\n", len(hosts), target)
		} else {
			// No MX records found, use the target as the host
			hosts = []string{target}
		}

		// Check each host
		var results []*types.SMTPResult
		for _, host := range hosts {
			fmt.Printf("Checking SMTP server: %s\n", host)

			var result *types.SMTPResult
			var err error

			// Determine which ports to check
			var ports []int
			if port != 0 {
				// Use the specified port
				ports = []int{port}
			} else {
				// Use the default ports
				ports = smtp.DefaultPorts
			}

			// Perform the SMTP check
			if checkPTR {
				result, err = smtp.CheckSMTPWithPTR(ctx, host, ports, timeoutDuration)
			} else {
				result, err = smtp.CheckSMTP(ctx, host, ports, timeoutDuration)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", host, err)
				continue
			}

			// If we're not checking for open relay, skip that part
			if !checkRelay {
				result.IsOpenRelay = nil
				result.RelayCheckError = ""
			}

			results = append(results, result)
		}

		// Output the results
		if outputFormat == "json" {
			jsonOutput, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
		} else {
			// Text output
			for i, result := range results {
				host := hosts[i]
				fmt.Printf("\nSMTP diagnostics for %s:\n", host)
				fmt.Printf("  Connection: %t\n", result.ConnectSuccess)
				if !result.ConnectSuccess {
					fmt.Printf("  Connection error: %s\n", result.ConnectError)
					fmt.Printf("  Note: Many properly configured mail servers restrict connections for security reasons.\n")
					fmt.Printf("        This is often a sign of good security practices rather than a problem.\n")
					continue
				}

				fmt.Printf("  Response time: %s\n", result.ResponseTime)

				if result.SupportsSTARTTLS != nil {
					fmt.Printf("  Supports STARTTLS: %t\n", *result.SupportsSTARTTLS)
					if result.STARTTLSError != "" {
						fmt.Printf("  STARTTLS error: %s\n", result.STARTTLSError)
					}
				}

				if result.IsOpenRelay != nil {
					fmt.Printf("  Open relay: %t\n", *result.IsOpenRelay)
					if result.RelayCheckError != "" {
						fmt.Printf("  Relay check error: %s\n", result.RelayCheckError)
					}
				}
			}
		}
	},
}

func init() {
	SMTPCmd.Flags().IntP("port", "p", 0, "SMTP port to check (default: check 25, 465, 587)")
	SMTPCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for SMTP operations")
	SMTPCmd.Flags().BoolP("check-relay", "r", false, "Check if the server is an open relay (use with caution)")
	SMTPCmd.Flags().BoolP("check-ptr", "P", false, "Check PTR records for the server")

	// Add the command to the root command
	rootCmd.AddCommand(SMTPCmd)
}
