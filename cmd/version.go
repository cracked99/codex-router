package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the application version
	Version = "0.1.0"
	// Commit is the git commit hash
	Commit = "unknown"
	// BuildTime is the build timestamp
	BuildTime = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit hash, and build time of codex-api-router.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("codex-api-router version %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built at: %s\n", BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
