package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/plasmadev/codex-api-router/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long: `Manage codex-router configuration files.

Configuration can be managed through:
  • YAML config files
  • Environment variables (CODEX_ROUTER_*)
  • Command-line flags

Priority order (highest to lowest):
  1. Command-line flags
  2. Environment variables  
  3. Config file
  4. Default values`,
}

// configInitCmd creates a new configuration file
var configInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new configuration file",
	Long: `Create a new configuration file with default values.

If no path is specified, creates the config at:
  ~/.codex-router/config.yaml

Examples:
  # Create default config
  codex-router config init

  # Create config at specific location
  codex-router config init ./my-config.yaml

  # Overwrite existing config
  codex-router config init --force`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine config path
		configPath := ""
		if len(args) > 0 {
			configPath = args[0]
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configPath = filepath.Join(home, ".codex-router", "config.yaml")
		}

		// Check if file exists
		force, _ := cmd.Flags().GetBool("force")
		if _, err := os.Stat(configPath); err == nil && !force {
			return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
		}

		// Create directory if needed
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Generate default config
		cfg := config.Default()
		
		// Interactive mode
		interactive, _ := cmd.Flags().GetBool("interactive")
		if interactive {
			if err := interactiveConfig(cfg); err != nil {
				return fmt.Errorf("interactive config failed: %w", err)
			}
		}

		// Save config
		if err := config.Save(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Created configuration file at: %s\n", configPath)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Edit the configuration file to set your z.ai API key")
		fmt.Println("  2. Or set the ZAI_API_KEY environment variable")
		fmt.Println("  3. Run 'codex-router serve' to start the server")
		
		return nil
	},
}

// configShowCmd displays current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Show the effective configuration from all sources.

This command merges configuration from:
  • Config file
  • Environment variables
  • Command-line flags
  • Default values

The output shows the final resolved configuration that will be used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get output format
		format := globalOpts.Output
		if f, _ := cmd.Flags().GetString("format"); f != "" {
			format = f
		}

		// Output configuration
		var output []byte
		var err2 error

		switch format {
		case "json":
			output, err2 = json.MarshalIndent(cfg, "", "  ")
		case "yaml":
			output, err2 = yaml.Marshal(cfg)
		default:
			output, err2 = yaml.Marshal(cfg)
		}

		if err2 != nil {
			return fmt.Errorf("failed to format config: %w", err2)
		}

		fmt.Println(string(output))
		return nil
	},
}

// configValidateCmd validates configuration
var configValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate configuration file",
	Long: `Validate a configuration file for correctness.

Checks for:
  • Valid YAML syntax
  • Required fields present
  • Valid values
  • Security best practices

Returns exit code 0 if valid, 1 if invalid.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config from specific path or default
		configPath := ""
		if len(args) > 0 {
			configPath = args[0]
		}

		cfg, err := loadConfigFromPath(configPath)
		if err != nil {
			fmt.Printf("✗ Configuration invalid: %v\n", err)
			return fmt.Errorf("validation failed")
		}

		// Additional security checks
		strict, _ := cmd.Flags().GetBool("strict")
		if strict {
			if err := validateSecurity(cfg); err != nil {
				fmt.Printf("⚠ Security warning: %v\n", err)
			}
		}

		fmt.Println("✓ Configuration is valid")
		fmt.Printf("  Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("  Backend: %s\n", cfg.Zai.BaseURL)
		fmt.Printf("  Translator: %s\n", cfg.Translator.Mode)
		
		if cfg.Server.TLS.Enabled {
			fmt.Println("  TLS: Enabled")
		}

		return nil
	},
}

// configEditCmd opens config in editor
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file in editor",
	Long: `Open the configuration file in your default editor.

The editor is determined by:
  1. $EDITOR environment variable
  2. $VISUAL environment variable
  3. System default editor

After editing, the configuration is automatically validated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find config file
		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configPath = filepath.Join(home, ".codex-router", "config.yaml")
		}

		// Check if exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("no config file found. Run 'codex-router config init' first")
		}

		// Determine editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vi" // fallback
		}

		// Open editor
		editCmd := exec.Command(editor, configPath)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		// Validate after edit
		fmt.Println("\nValidating configuration...")
		if err := configValidateCmd.RunE(configValidateCmd, []string{}); err != nil {
			return err
		}
		
		return nil
	},
}

// configSetCmd sets individual config values
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set an individual configuration value in the config file.

Keys use dot notation for nested values.

Examples:
  codex-router config set server.port 9090
  codex-router config set zai.api_key sk-xxx
  codex-router config set logging.level debug`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Load current config
		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configPath = filepath.Join(home, ".codex-router", "config.yaml")
		}

		cfg, err := loadConfigFromPath(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Set value (simplified version - in production would use reflection)
		if err := setConfigValue(cfg, key, value); err != nil {
			return fmt.Errorf("failed to set value: %w", err)
		}

		// Save
		if err := config.Save(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Set %s = %s\n", key, value)
		return nil
	},
}

// configGetCmd gets individual config values
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get an individual configuration value.

Keys use dot notation for nested values.

Examples:
  codex-router config get server.port
  codex-router config get zai.base_url`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get value
		value, err := getConfigValue(cfg, key)
		if err != nil {
			return fmt.Errorf("failed to get value: %w", err)
		}

		fmt.Println(value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)

	// Init flags
	configInitCmd.Flags().Bool("force", false, "overwrite existing config file")
	configInitCmd.Flags().BoolP("interactive", "i", false, "interactive configuration")

	// Show flags
	configShowCmd.Flags().StringP("format", "f", "yaml", "output format (yaml, json)")
	
	// Validate flags
	configValidateCmd.Flags().Bool("strict", false, "enable strict security validation")
}

// Helper functions

func loadConfigFromPath(path string) (*config.Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}
	return GetConfig()
}

func interactiveConfig(cfg *config.Config) error {
	// Simple interactive prompts (in production would use a proper library)
	fmt.Println("\nCodex Router Configuration")
	fmt.Println("==========================")
	
	fmt.Print("Server host [localhost]: ")
	var host string
	fmt.Scanln(&host)
	if host != "" {
		cfg.Server.Host = host
	}

	fmt.Print("Server port [8080]: ")
	var port int
	fmt.Scanln(&port)
	if port != 0 {
		cfg.Server.Port = port
	}

	fmt.Print("z.ai API key: ")
	var apiKey string
	fmt.Scanln(&apiKey)
	if apiKey != "" {
		cfg.Zai.APIKey = apiKey
	}

	fmt.Print("Translator mode (wasm/sidecar) [wasm]: ")
	var mode string
	fmt.Scanln(&mode)
	if mode != "" {
		cfg.Translator.Mode = mode
	}

	return nil
}

func validateSecurity(cfg *config.Config) error {
	// Security checks
	if cfg.Server.Host == "0.0.0.0" {
		return fmt.Errorf("binding to 0.0.0.0 exposes service to all interfaces")
	}
	
	if cfg.Zai.APIKey == "" {
		return fmt.Errorf("no API key configured")
	}

	if strings.HasPrefix(cfg.Zai.APIKey, "sk-test") {
		return fmt.Errorf("using test API key in production")
	}

	return nil
}

func setConfigValue(cfg *config.Config, key, value string) error {
	// Simplified implementation - production would use reflection
	switch key {
	case "server.host":
		cfg.Server.Host = value
	case "server.port":
		fmt.Sscanf(value, "%d", &cfg.Server.Port)
	case "zai.api_key":
		cfg.Zai.APIKey = value
	case "zai.base_url":
		cfg.Zai.BaseURL = value
	case "logging.level":
		cfg.Logging.Level = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	// Simplified implementation - production would use reflection
	switch key {
	case "server.host":
		return cfg.Server.Host, nil
	case "server.port":
		return fmt.Sprintf("%d", cfg.Server.Port), nil
	case "zai.api_key":
		if cfg.Zai.APIKey != "" {
			return "***" + cfg.Zai.APIKey[len(cfg.Zai.APIKey)-4:], nil
		}
		return "", nil
	case "zai.base_url":
		return cfg.Zai.BaseURL, nil
	case "logging.level":
		return cfg.Logging.Level, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
