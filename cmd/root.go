package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plasmadev/codex-api-router/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Build information (set via ldflags)
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// GlobalOptions holds global CLI options
type GlobalOptions struct {
	ConfigFile string
	Verbose    bool
	Debug      bool
	Output     string // Output format: text, json, yaml
	NoColor    bool
}

var globalOpts = GlobalOptions{}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "codex-router",
	Short: "API router for translating between Codex CLI and z.ai APIs",
	Long: `Codex API Router is a production-grade proxy service that translates between 
Codex CLI's Responses API and z.ai's Chat Completions API.

This enables Codex CLI v0.99+ (which only supports Responses API) to work 
seamlessly with z.ai (which only supports Chat Completions API).

Features:
  • Transparent API translation
  • Streaming support (SSE)
  • Configuration management
  • Health monitoring
  • Metrics collection
  • Multiple deployment modes

Quick Start:
  # Initialize configuration
  codex-router config init

  # Start the server
  codex-router serve

  # Check health
  codex-router health

For more information, see: https://github.com/plasmadev/codex-api-router`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		return initConfig()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVarP(&globalOpts.ConfigFile, "config", "c", "", 
		"config file (default is $HOME/.codex-router/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&globalOpts.Verbose, "verbose", "v", false, 
		"verbose output")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.Debug, "debug", false, 
		"debug mode (very verbose)")
	rootCmd.PersistentFlags().StringVarP(&globalOpts.Output, "output", "o", "text", 
		"output format (text, json, yaml)")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.NoColor, "no-color", false, 
		"disable colored output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig initializes configuration
func initConfig() error {
	// Set config file location
	if globalOpts.ConfigFile != "" {
		viper.SetConfigFile(globalOpts.ConfigFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Default config path
		configDir := filepath.Join(home, ".codex-router")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variable handling
	viper.SetEnvPrefix("CODEX_ROUTER")
	viper.AutomaticEnv()

	// Read config file (optional, may not exist yet)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found is okay, we'll use defaults
	}

	return nil
}

// GetConfig loads and returns the current configuration
func GetConfig() (*config.Config, error) {
	cfg := config.Default()

	// Unmarshal from viper
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables (support both ZAI_API_KEY and Z_AI_API_KEY)
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("Z_AI_API_KEY")
	}
	if apiKey != "" {
		cfg.Zai.APIKey = apiKey
		// Also set provider config
		cfg.Providers.Zai.APIKey = apiKey
		cfg.Providers.Zai.Enabled = true
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// GetVersion returns version information
func GetVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date)
}

// SaveConfig saves the configuration to a file
func SaveConfig(path string, cfg *config.Config) error {
	return config.Save(path, cfg)
}
