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
)

// dnsCmd represents the dns command
var DnsCmd = &cobra.Command{
	Use:   "dns [domain]",
	Short: "Perform DNS lookups",
	Long: `Perform DNS lookups for a domain.
Supports various record types including A, AAAA, MX, TXT, CNAME, NS, SOA, PTR.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		recordType, _ := cmd.Flags().GetString("type")
		server, _ := cmd.Flags().GetString("server")
		advanced, _ := cmd.Flags().GetBool("advanced")
		all, _ := cmd.Flags().GetBool("all")
		timeout, _ := cmd.Flags().GetInt("timeout")
		retries, _ := cmd.Flags().GetInt("retries")
		outputFormat, _ := cmd.Flags().GetString("output")

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
