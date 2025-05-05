package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
	"github.com/spf13/viper"
)

// InitConfig initializes the Viper configuration
func InitConfig() {
	// Set default values
	viper.SetDefault("output", "./dashboard")
	viper.SetDefault("cache-dir", "./.cache")
	viper.SetDefault("cache-ttl", "1h")
	viper.SetDefault("verbose", false)

	// Environment variables
	viper.SetEnvPrefix("GITHUB_DASHBOARD") // will convert to GITHUB_DASHBOARD_*
	viper.AutomaticEnv()

	// Check for token in environment
	if token := os.Getenv("GITHUB_TOKEN"); token != "" && viper.GetString("token") == "" {
		viper.Set("token", token)
	}
}

// GetConfig builds and validates the configuration from Viper
func GetConfig() (*types.Config, error) {
	cacheTTL, err := time.ParseDuration(viper.GetString("cache-ttl"))
	if err != nil {
		return nil, errors.New("invalid cache-ttl format: use a valid duration string (e.g., 1h, 30m)")
	}

	config := &types.Config{
		User:         viper.GetString("user"),
		Organization: viper.GetString("org"),
		OutputDir:    viper.GetString("output"),
		GithubToken:  viper.GetString("token"),
		CacheDir:     viper.GetString("cache-dir"),
		CacheTTL:     cacheTTL,
		Verbose:      viper.GetBool("verbose"),
	}

	// Validate config
	if config.User == "" && config.Organization == "" {
		return nil, errors.New("either user or organization must be specified")
	}

	if config.User != "" && config.Organization != "" {
		return nil, errors.New("only one of user or organization can be specified")
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll(config.OutputDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create repositories directory
	err = os.MkdirAll(filepath.Join(config.OutputDir, "repositories"), 0755)
	if err != nil {
		return nil, err
	}

	// Create cache directory if it doesn't exist
	err = os.MkdirAll(config.CacheDir, 0755)
	if err != nil {
		return nil, err
	}

	return config, nil
}
