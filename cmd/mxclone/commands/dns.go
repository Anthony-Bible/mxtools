// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"mxclone/domain/dns"
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
		all, _ := cmd.Flags().GetBool("all")
		timeout, _ := cmd.Flags().GetInt("timeout")
		outputFormat, _ := cmd.Flags().GetString("output")

		// Get and validate record type
		recordTypeStr, _ := cmd.Flags().GetString("type")
		recordTypeStr = validation.SanitizeDNSRecordType(recordTypeStr)
		if !all {
			if err := validation.ValidateDNSRecordType(recordTypeStr); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		recordType := dns.RecordType(recordTypeStr)

		// Get and validate server if provided
		server, _ := cmd.Flags().GetString("server")
		if server != "" {
			if err := validation.ValidateServer(server); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Performing DNS lookup for %s (type: %s)...\n", domain, recordType)

		// Get the DNS service from the dependency injection container
		dnsService := Container.GetDNSService()

		ctx := context.Background()
		var result *dns.DNSResult
		var err error

		// Set timeout duration
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		if all {
			// Lookup all record types
			result, err = dnsService.LookupAll(timeoutCtx, domain)
		} else {
			// Lookup specific record type
			result, err = dnsService.Lookup(timeoutCtx, domain, recordType)
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
