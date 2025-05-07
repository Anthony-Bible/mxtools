package commands

import (
	"mxclone/internal/api"

	"github.com/spf13/cobra"
)

// ApiCmd starts the API server
var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Long:  `Start the HTTP API server for diagnostics`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get services and logger from the shared DI container
		dnsService := Container.GetDNSService()
		dnsblService := Container.GetDNSBLService()
		logger := Container.GetLogger()

		// Start API server with dependencies
		err := api.StartAPIServer(
			dnsService,
			dnsblService,
			logger,
		)

		if err != nil {
			logger.Fatal("Failed to start API server: %v", err)
		}
	},
}
