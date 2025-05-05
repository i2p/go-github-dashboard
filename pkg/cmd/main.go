// pkg/cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/go-i2p/go-github-dashboard/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "go-github-dashboard",
	Short: "Generate a static GitHub dashboard",
	Long: `A pure Go command-line application that generates a static 
GitHub dashboard by aggregating repository data from GitHub API 
and RSS feeds, organizing content in a repository-by-repository structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		// The root command will just show help
		cmd.Help()
	},
}

// Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)

	// Persistent flags for all commands
	rootCmd.PersistentFlags().StringP("user", "u", "", "GitHub username to generate dashboard for")
	rootCmd.PersistentFlags().StringP("org", "o", "", "GitHub organization to generate dashboard for")
	rootCmd.PersistentFlags().StringP("output", "d", "./dashboard", "Output directory for the dashboard")
	rootCmd.PersistentFlags().StringP("token", "t", "", "GitHub API token (optional, increases rate limits)")
	rootCmd.PersistentFlags().String("cache-dir", "./.cache", "Directory for caching API responses")
	rootCmd.PersistentFlags().String("cache-ttl", "1h", "Cache time-to-live duration (e.g., 1h, 30m)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("org", rootCmd.PersistentFlags().Lookup("org"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("cache-dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	viper.BindPFlag("cache-ttl", rootCmd.PersistentFlags().Lookup("cache-ttl"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}
