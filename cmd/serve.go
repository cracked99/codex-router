package cmd

import (
	"github.com/plasmadev/codex-api-router/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long: `Start the codex-api-router HTTP server that listens for Responses API
requests and proxies them to z.ai's Chat Completions API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get configuration from viper
		cfg := server.Config{
			Host:    viper.GetString("server.host"),
			Port:    viper.GetInt("server.port"),
			ZaiKey:  viper.GetString("zai.api_key"),
			ZaiURL:  viper.GetString("zai.base_url"),
		}

		// Override with command-line flags if provided
		if host, _ := cmd.Flags().GetString("host"); host != "" {
			cfg.Host = host
		}
		if port, _ := cmd.Flags().GetInt("port"); port != 0 {
			cfg.Port = port
		}

		// Start the server
		return server.Start(&cfg)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server flags
	serveCmd.Flags().StringP("host", "H", "localhost", "Host to bind to")
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")

	// Z.ai flags
	serveCmd.Flags().String("zai-url", "https://api.z.ai/api/paas/v4", "z.ai base URL")
	serveCmd.Flags().String("zai-api-key", "", "z.ai API key (overrides ZAI_API_KEY env var)")

	// Bind to viper
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("zai.base_url", serveCmd.Flags().Lookup("zai-url"))
	viper.BindPFlag("zai.api_key", serveCmd.Flags().Lookup("zai-api-key"))

	// Set defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("zai.base_url", "https://api.z.ai/api/paas/v4")
}
