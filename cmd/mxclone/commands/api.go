package commands

import (
	"github.com/spf13/cobra"
	"mxclone/internal/api"
)

// ApiCmd starts the API server
var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Long:  `Start the HTTP API server for diagnostics`,
	Run: func(cmd *cobra.Command, args []string) {
		api.StartAPIServer()
	},
}
