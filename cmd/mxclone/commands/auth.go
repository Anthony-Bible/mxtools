// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"mxclone/pkg/emailauth"
	"mxclone/pkg/types"
)

// AuthCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth [domain]",
	Short: "Check email authentication",
	Long: `Check email authentication mechanisms for a domain.
This includes SPF, DKIM, and DMARC record retrieval and validation.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		timeout, _ := cmd.Flags().GetInt("timeout")
		selector, _ := cmd.Flags().GetString("selector")
		checkDKIM, _ := cmd.Flags().GetBool("check-dkim")
		headerFile, _ := cmd.Flags().GetString("header-file")
		outputFormat, _ := cmd.Flags().GetString("output")

		fmt.Printf("Checking email authentication for %s...\n", domain)

		ctx := context.Background()
		timeoutDuration := time.Duration(timeout) * time.Second

		// Determine which check to perform based on flags
		var result *types.AuthResult
		var err error

		// If DKIM check is requested and a selector is provided
		if checkDKIM && selector != "" {
			fmt.Printf("Checking email authentication with DKIM for selector %s...\n", selector)
			result, err = emailauth.CheckEmailAuthWithDKIM(ctx, domain, selector, timeoutDuration)
		} else if headerFile != "" {
			// If header file is provided
			fmt.Printf("Checking email authentication and analyzing header from %s...\n", headerFile)
			headerData, err := os.ReadFile(headerFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading header file: %v\n", err)
				os.Exit(1)
			}
			result, err = emailauth.CheckEmailAuthWithHeader(ctx, domain, string(headerData), timeoutDuration)
		} else {
			// Basic email authentication check
			result, err = emailauth.CheckEmailAuth(ctx, domain, timeoutDuration)
		}

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
			// Text output
			fmt.Printf("\nEmail authentication results for %s:\n", domain)

			// SPF results
			fmt.Println("\nSPF:")
			if result.SPFRecord != "" {
				fmt.Printf("  Record: %s\n", result.SPFRecord)
				if result.SPFResult != "" {
					fmt.Printf("  Result: %s\n", result.SPFResult)
				}
			} else {
				fmt.Printf("  Error: %s\n", result.SPFError)
			}

			// DMARC results
			fmt.Println("\nDMARC:")
			if result.DMARCRecord != "" {
				fmt.Printf("  Record: %s\n", result.DMARCRecord)
				if result.DMARCPolicy != "" {
					fmt.Printf("  Policy: %s\n", result.DMARCPolicy)
				}
			} else {
				fmt.Printf("  Error: %s\n", result.DMARCError)
			}

			// DKIM results (if available)
			if result.DKIMRecord != "" || result.DKIMError != "" {
				fmt.Println("\nDKIM:")
				if result.DKIMRecord != "" {
					fmt.Printf("  Record: %s\n", result.DKIMRecord)
					if result.DKIMResult != "" {
						fmt.Printf("  Result: %s\n", result.DKIMResult)
					}
				} else {
					fmt.Printf("  Error: %s\n", result.DKIMError)
				}
			}

			// Header authentication results (if available)
			if result.HeaderAuth != nil && len(result.HeaderAuth) > 0 {
				fmt.Println("\nEmail Header Authentication:")
				for mechanism, value := range result.HeaderAuth {
					fmt.Printf("  %s: %s\n", mechanism, value)
				}
			}
		}
	},
}

func init() {
	AuthCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds for DNS operations")
	AuthCmd.Flags().StringP("selector", "s", "", "DKIM selector to check")
	AuthCmd.Flags().BoolP("check-dkim", "d", false, "Check DKIM record (requires selector)")
	AuthCmd.Flags().StringP("header-file", "f", "", "File containing email headers to analyze")

	// Add the command to the root command
	rootCmd.AddCommand(AuthCmd)
}
