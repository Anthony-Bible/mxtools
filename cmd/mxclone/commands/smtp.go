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

// SMTPCmd represents the smtp command
var SMTPCmd = &cobra.Command{
	Use:   "smtp [domain]",
	Short: "Perform SMTP diagnostics",
	Long: `Perform SMTP diagnostics on a mail server.
This checks connectivity, STARTTLS support, and other SMTP capabilities.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get and validate target (domain or IP)
		domain := args[0]
		// Try to validate as domain
		if err := validation.ValidateDomain(domain); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Get command flags
		timeout, _ := cmd.Flags().GetInt("timeout")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Performing SMTP diagnostics for %s...\n", domain)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Get the SMTP service from the dependency injection container
		smtpService := Container.GetSMTPService()

		// Perform the SMTP check
		result, err := smtpService.CheckSMTP(ctx, domain, timeoutDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking SMTP for %s: %v\n", domain, err)
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
			summary := smtpService.GetSMTPSummary(result)
			fmt.Println(summary)
		}
	},
}

func init() {
	SMTPCmd.Flags().IntP("port", "p", 25, "SMTP port to check")
	SMTPCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for SMTP operations")
}
