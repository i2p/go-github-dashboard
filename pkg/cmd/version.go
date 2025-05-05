// pkg/cmd/version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information
var (
	Version   = "0.1.0"
	BuildDate = "unknown"
	Commit    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version, build date, and commit hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-github-dashboard version %s\n", Version)
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Commit: %s\n", Commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
