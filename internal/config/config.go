// Package config provides configuration functionality for the MXToolbox clone.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	// General settings
	WorkerCount int    `mapstructure:"worker_count"`
	LogLevel    string `mapstructure:"log_level"`
	CacheDir    string `mapstructure:"cache_dir"`

	// DNS settings
	DNSTimeout     int      `mapstructure:"dns_timeout"`
	DNSRetries     int      `mapstructure:"dns_retries"`
	DNSResolvers   []string `mapstructure:"dns_resolvers"`
	DNSCacheTTL    int      `mapstructure:"dns_cache_ttl"`

	// Blacklist settings
	BlacklistZones    []string `mapstructure:"blacklist_zones"`
	BlacklistTimeout  int      `mapstructure:"blacklist_timeout"`
	BlacklistCacheTTL int      `mapstructure:"blacklist_cache_ttl"`

	// SMTP settings
	SMTPTimeout int `mapstructure:"smtp_timeout"`
	SMTPPorts   []int `mapstructure:"smtp_ports"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		WorkerCount: 10,
		LogLevel:    "info",
		CacheDir:    filepath.Join(os.TempDir(), "mxclone"),

		DNSTimeout:   5,
		DNSRetries:   2,
		DNSResolvers: []string{"8.8.8.8:53", "1.1.1.1:53"},
		DNSCacheTTL:  300, // 5 minutes

		BlacklistZones: []string{
			"zen.spamhaus.org",
			"bl.spamcop.net",
			"dnsbl.sorbs.net",
		},
		BlacklistTimeout:  10,
		BlacklistCacheTTL: 1800, // 30 minutes

		SMTPTimeout: 10,
		SMTPPorts:   []int{25, 465, 587},
	}
}

// LoadConfig loads the configuration from file and environment variables.
func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()

	// Set default values
	defaultConfig := DefaultConfig()
	v.SetDefault("worker_count", defaultConfig.WorkerCount)
	v.SetDefault("log_level", defaultConfig.LogLevel)
	v.SetDefault("cache_dir", defaultConfig.CacheDir)
	v.SetDefault("dns_timeout", defaultConfig.DNSTimeout)
	v.SetDefault("dns_retries", defaultConfig.DNSRetries)
	v.SetDefault("dns_resolvers", defaultConfig.DNSResolvers)
	v.SetDefault("dns_cache_ttl", defaultConfig.DNSCacheTTL)
	v.SetDefault("blacklist_zones", defaultConfig.BlacklistZones)
	v.SetDefault("blacklist_timeout", defaultConfig.BlacklistTimeout)
	v.SetDefault("blacklist_cache_ttl", defaultConfig.BlacklistCacheTTL)
	v.SetDefault("smtp_timeout", defaultConfig.SMTPTimeout)
	v.SetDefault("smtp_ports", defaultConfig.SMTPPorts)

	// Set config file name and path
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Look for config in the following directories
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.mxclone")
		v.AddConfigPath("/etc/mxclone")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read environment variables
	v.SetEnvPrefix("MXCLONE")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return config, nil
}