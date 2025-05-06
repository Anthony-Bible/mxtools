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

// AuthCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth [domain]",
	Short: "Check email authentication",
	Long: `Check email authentication mechanisms for a domain.
This includes SPF, DKIM, and DMARC record retrieval and validation.`,
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
		timeout, _ := cmd.Flags().GetInt("timeout")
		checkDKIM, _ := cmd.Flags().GetBool("check-dkim")
		outputFormat, _ := cmd.Flags().GetString("output")

		// Get and validate selector if provided
		selector, _ := cmd.Flags().GetString("selector")
		if selector != "" {
			selector = validation.SanitizeSelector(selector)
			if err := validation.ValidateSelector(selector); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Checking email authentication for %s...\n", domain)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Get the EmailAuth service from the dependency injection container
		emailAuthService := Container.GetEmailAuthService()

		// Define DKIM selectors to check
		dkimSelectors := []string{}

		// If DKIM check is requested
		if checkDKIM {
			// If a selector is provided, use it
			if selector != "" {
				fmt.Printf("Checking email authentication with DKIM for selector %s...\n", selector)
				dkimSelectors = []string{selector}
			} else {
				// Try with common selectors
				fmt.Printf("No selector provided, trying with default selectors...\n")
				dkimSelectors = []string{"mail", "default", "google", "selector1", "dkim"}
			}
		}

		// Perform the check
		result, err := emailAuthService.CheckAll(ctx, domain, dkimSelectors, timeoutDuration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking email authentication: %v\n", err)
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
			// Text output - use the formatted summary from the service
			summary := emailAuthService.GetAuthSummary(result)
			fmt.Println(summary)
		}
	},
}

func init() {
	AuthCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for DNS operations")
	AuthCmd.Flags().StringP("selector", "s", "", "DKIM selector to check")
	AuthCmd.Flags().BoolP("check-dkim", "d", false, "Check DKIM record (requires selector)")

	// Add the command to the root command
	rootCmd.AddCommand(AuthCmd)
}
