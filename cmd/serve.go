package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plasmadev/codex-api-router/internal/config"
	"github.com/plasmadev/codex-api-router/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API router server",
	Long: `Start the codex-api-router HTTP server.

The server listens for Responses API requests and proxies them to z.ai's 
Chat Completions API with automatic translation.

Examples:
  # Start with default configuration
  codex-router serve

  # Start with custom port
  codex-router serve --port 9090

  # Start with API key
  codex-router serve --api-key sk-xxx

  # Start in development mode
  codex-router serve --dev

  # Start with custom backend
  codex-router serve --backend-url https://custom.api.com

The server supports:
  • Hot reload (in dev mode)
  • Graceful shutdown
  • Health monitoring
  • Metrics collection
  • Request logging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Override with command-line flags
		if port, _ := cmd.Flags().GetInt("port"); port != 0 {
			cfg.Server.Port = port
		}
		if host, _ := cmd.Flags().GetString("host"); host != "" {
			cfg.Server.Host = host
		}
		if apiKey, _ := cmd.Flags().GetString("api-key"); apiKey != "" {
			cfg.Zai.APIKey = apiKey
		}
		if backendURL, _ := cmd.Flags().GetString("backend-url"); backendURL != "" {
			cfg.Zai.BaseURL = backendURL
		}
		if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout != 0 {
			cfg.Zai.Timeout = timeout
		}
		if mode, _ := cmd.Flags().GetString("translator-mode"); mode != "" {
			cfg.Translator.Mode = mode
		}
		if dev, _ := cmd.Flags().GetBool("dev"); dev {
			cfg.Translator.Mode = "sidecar"
			cfg.Logging.Level = "debug"
		}

		// Bind flags to viper for persistence
		viper.BindPFlag("server.port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("server.host", cmd.Flags().Lookup("host"))
		viper.BindPFlag("zai.api_key", cmd.Flags().Lookup("api-key"))
		viper.BindPFlag("zai.base_url", cmd.Flags().Lookup("backend-url"))

		// Print startup banner
		printBanner(cfg)

		// Create server
		srv := server.New(cfg)

		// Setup signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		// Start server in goroutine
		errChan := make(chan error, 1)
		go func() {
			if err := srv.Start(); err != nil {
				errChan <- err
			}
		}()

		// Wait for shutdown signal or error
		select {
		case err := <-errChan:
			return fmt.Errorf("server error: %w", err)
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down gracefully...\n", sig)
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				return fmt.Errorf("shutdown error: %w", err)
			}

			fmt.Println("✓ Server shutdown complete")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server configuration flags
	serveCmd.Flags().StringP("host", "H", "", 
		"host to bind to (overrides config)")
	serveCmd.Flags().IntP("port", "p", 0, 
		"port to listen on (overrides config)")
	serveCmd.Flags().StringP("api-key", "k", "", 
		"z.ai API key (overrides config)")
	serveCmd.Flags().StringP("backend-url", "b", "", 
		"backend URL for z.ai API (overrides config)")
	serveCmd.Flags().Duration("timeout", 0, 
		"request timeout (e.g., 120s)")
	serveCmd.Flags().String("translator-mode", "", 
		"translator mode (wasm or sidecar)")
	serveCmd.Flags().BoolP("dev", "D", false, 
		"enable development mode (sidecar translator, debug logging)")
	serveCmd.Flags().Bool("tls", false, 
		"enable TLS")
	serveCmd.Flags().String("tls-cert", "", 
		"TLS certificate file")
	serveCmd.Flags().String("tls-key", "", 
		"TLS private key file")
	serveCmd.Flags().BoolP("dry-run", "n", false, 
		"validate configuration without starting server")
}

func printBanner(cfg *config.Config) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Codex API Router - Production Ready              ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Version:     %s\n", GetVersion())
	fmt.Printf("  Server:      http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  Backend:     %s\n", cfg.Zai.BaseURL)
	fmt.Printf("  Translator:  %s\n", cfg.Translator.Mode)
	fmt.Printf("  Log Level:   %s\n", cfg.Logging.Level)
	
	if cfg.Server.TLS.Enabled {
		fmt.Println("  TLS:         Enabled ✓")
	}
	
	fmt.Println()
	fmt.Println("  Endpoints:")
	fmt.Printf("    Proxy:    POST http://%s:%d/v1/responses\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("    Health:   GET  http://%s:%d/health\n", cfg.Server.Host, cfg.Server.Port)
	
	if cfg.Metrics.Enabled {
		fmt.Printf("    Metrics:  GET  http://%s:%d%s\n", cfg.Server.Host, cfg.Server.Port, cfg.Metrics.Path)
	}
	
	fmt.Println()
	fmt.Println("  Press Ctrl+C to shutdown gracefully")
	fmt.Println()
}
