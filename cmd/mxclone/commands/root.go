// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mxclone/internal/di"
)

var (
	// Container is the dependency injection container for the application
	Container *di.Container

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "mxclone",
		Short: "A clone of MXToolbox with DNS and SMTP diagnostic capabilities",
		Long: `MXClone is a command-line tool that provides DNS and SMTP diagnostic capabilities,
similar to what MXToolbox offers on their website.

The tool can check DNS records, verify email authentication (SPF, DKIM, DMARC),
check if an IP is blacklisted, and test SMTP connectivity.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Initialize the dependency injection container
	Container = di.NewContainer()

	// Execute the root command
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add shared global flags here
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format (text, json)")
}
