// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
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

		fmt.Printf("Performing DNS lookup for %s (type: %s)...\n", domain, recordType)

		// In a real implementation, this would use the engine to perform the lookup
		// For now, we'll just print a placeholder message
		fmt.Printf("DNS lookup for %s completed.\n", domain)
	},
}

func init() {
	DnsCmd.Flags().StringP("type", "t", "A", "Record type (A, AAAA, MX, TXT, CNAME, NS, SOA, PTR)")
}
