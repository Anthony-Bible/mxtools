// Package commands contains the CLI commands for the MXToolbox clone.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"mxclone/internal/config"
	"mxclone/pkg/orchestration"
)

var (
	cfgFile string
	cfg     *config.Config
	engine  *orchestration.Engine
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mxclone",
	Short: "MXToolbox clone - A comprehensive email and network diagnostic tool",
	Long: `MXToolbox clone is a comprehensive email and network diagnostic tool
that provides DNS lookups, blacklist checks, SMTP diagnostics, email
authentication validation, and more.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load configuration
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Initialize engine
		engine = orchestration.NewEngine(cfg.WorkerCount)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func init() {
	// Here you will define your flags and configuration settings.
	cobra.OnInitialize(initConfig)

	// Persistent flags for the root command
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mxclone/config.yaml)")
	rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format (text, json)")

	// Local flags for the root command
	rootCmd.Flags().BoolP("version", "v", false, "Display version information")

	// Add subcommands
	rootCmd.AddCommand(DnsCmd)
	rootCmd.AddCommand(BlacklistCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".mxclone" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".mxclone")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}