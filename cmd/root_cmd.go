package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "codex-api-router",
	Short: "API router for Codex CLI to z.ai translation",
	Long: `Codex API Router is a proxy service that translates between Codex CLI's
Responses API and z.ai's Chat Completions API.

This enables Codex CLI v0.99+ (which only supports Responses API) to work with
z.ai (which only supports Chat Completions API).`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootOpts.Config, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().CountVarP(&rootOpts.Verbosity, "verbose", "v", "verbosity level (0=error, 1=info, 2=debug)")
}
