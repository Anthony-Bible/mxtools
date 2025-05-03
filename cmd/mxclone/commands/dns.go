// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"mxclone/pkg/dns"
	"mxclone/pkg/types"
	"mxclone/pkg/validation"
)

// dnsCmd represents the dns command
var DnsCmd = &cobra.Command{
	Use:   "dns [domain]",
	Short: "Perform DNS lookups",
	Long: `Perform DNS lookups for a domain.
Supports various record types including A, AAAA, MX, TXT, CNAME, NS, SOA, PTR.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate domain
		domain := args[0]
		domain = validation.SanitizeDomain(domain)
		if err := validation.ValidateDomain(domain); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Get command flags
		advanced, _ := cmd.Flags().GetBool("advanced")
		all, _ := cmd.Flags().GetBool("all")
		timeout, _ := cmd.Flags().GetInt("timeout")
		retries, _ := cmd.Flags().GetInt("retries")
		outputFormat, _ := cmd.Flags().GetString("output")

		// Get and validate record type
		recordType, _ := cmd.Flags().GetString("type")
		recordType = validation.SanitizeDNSRecordType(recordType)
		if !all {
			if err := validation.ValidateDNSRecordType(recordType); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		// Get and validate server if provided
		server, _ := cmd.Flags().GetString("server")
		if server != "" {
			if err := validation.ValidateServer(server); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Performing DNS lookup for %s (type: %s)...\n", domain, recordType)

		ctx := context.Background()
		var result *types.DNSResult
		var err error

		// Set timeout duration
		timeoutDuration := time.Duration(timeout) * time.Second

		if all {
			// Lookup all record types
			if advanced {
				result, err = dns.AdvancedLookupAll(ctx, domain, server)
			} else {
				result, err = dns.LookupAll(ctx, domain)
			}
		} else {
			// Lookup specific record type
			if advanced {
				result, err = dns.AdvancedLookupWithRetry(ctx, domain, recordType, server, retries, timeoutDuration)
			} else {
				result, err = dns.LookupWithRetry(ctx, domain, recordType, retries, timeoutDuration)
			}
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
			fmt.Printf("DNS lookup results for %s:\n", domain)
			for recordType, records := range result.Lookups {
				fmt.Printf("\n%s records:\n", recordType)
				for _, record := range records {
					fmt.Printf("  %s\n", record)
				}
			}
			if result.Error != "" {
				fmt.Printf("\nErrors: %s\n", result.Error)
			}
		}
	},
}

func init() {
	DnsCmd.Flags().StringP("type", "t", "A", "Record type (A, AAAA, MX, TXT, CNAME, NS, SOA, PTR)")
	DnsCmd.Flags().StringP("server", "s", "", "DNS server to query (e.g., 8.8.8.8)")
	DnsCmd.Flags().BoolP("advanced", "a", false, "Use advanced DNS lookup (miekg/dns)")
	DnsCmd.Flags().BoolP("all", "l", false, "Lookup all record types")
	DnsCmd.Flags().IntP("timeout", "T", 5, "Timeout in seconds")
	DnsCmd.Flags().IntP("retries", "r", 2, "Number of retries")
}
