package commands

import (
	"mxclone/internal"
	"mxclone/internal/api"
	"mxclone/internal/config"

	"github.com/spf13/cobra"
)

// ApiCmd starts the API server
var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Long:  `Start the HTTP API server for diagnostics`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadConfig("")
		if err != nil {
			Container.GetLogger().Fatal("Failed to load configuration: %v", err)
		}

		// Initialize the job store with the loaded configuration
		internal.InitJobStore(cfg)

		// Get services and logger from the shared DI container
		dnsService := Container.GetDNSService()
		dnsblService := Container.GetDNSBLService()
		smtpService := Container.GetSMTPService()
		emailAuthService := Container.GetEmailAuthService()
		networkToolsService := Container.GetNetworkToolsService()
		logger := Container.GetLogger()

		// Start API server with dependencies
		err = api.StartAPIServer(
			dnsService,
			dnsblService,
			smtpService,
			emailAuthService,
			networkToolsService,
			logger,
		)

		if err != nil {
			logger.Fatal("Failed to start API server: %v", err)
		}
	},
}
