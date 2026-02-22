package cmd

import (
	"fmt"
	
	"github.com/spf13/cobra"
)

// providerCmd represents provider management commands
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Provider management commands",
	Long: `Manage LLM providers for the router.

Supports multiple providers including:
  • z.ai (primary)
  • OpenAI
  • Anthropic (coming soon)

Commands:
  list       List all providers
  health     Check provider health
  enable     Enable a provider
  disable    Disable a provider
  test       Test a provider
  metrics    Show provider metrics`,
}

// providerListCmd lists all configured providers
var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured providers",
	Long: `List all configured providers with their status.

Shows:
  • Provider name and type
  • Enabled/disabled status
  • Priority
  • Supported models
  • Health status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("Configured Providers:")
		fmt.Println("=====================")
		
		for name, provider := range cfg.Providers.GetProviders() {
			status := "disabled"
			if provider.Enabled {
				status = "enabled"
			}
			
			fmt.Printf("\n%s (%s):\n", name, provider.Type)
			fmt.Printf("  Status: %s\n", status)
			fmt.Printf("  Priority: %d\n", provider.Priority)
			fmt.Printf("  Base URL: %s\n", provider.BaseURL)
			
			if len(provider.Models) > 0 {
				fmt.Printf("  Models: %v\n", provider.Models)
			}
			
			if provider.Enabled && provider.APIKey != "" {
				fmt.Printf("  API Key: ***%s\n", provider.APIKey[len(provider.APIKey)-4:])
			}
		}
		
		return nil
	},
}

// providerHealthCmd checks provider health
var providerHealthCmd = &cobra.Command{
	Use:   "health [provider-name]",
	Short: "Check provider health",
	Long: `Check the health status of one or all providers.

If no provider name is given, checks all providers.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(args) > 0 {
			// Check specific provider
			name := args[0]
			provider, exists := cfg.Providers.GetProviders()[name]
			if !exists {
				return fmt.Errorf("provider not found: %s", name)
			}
			
			fmt.Printf("Provider: %s\n", name)
			fmt.Printf("Status: ")
			if provider.Enabled {
				fmt.Println("✓ Enabled")
				if provider.APIKey != "" {
					fmt.Println("API Key: ✓ Configured")
				} else {
					fmt.Println("API Key: ✗ Not configured")
				}
			} else {
				fmt.Println("✗ Disabled")
			}
		} else {
			// Check all providers
			fmt.Println("Provider Health Status:")
			fmt.Println("=======================")
			
			for name, provider := range cfg.Providers.GetProviders() {
				status := "✗ Disabled"
				if provider.Enabled {
					if provider.APIKey != "" {
						status = "✓ Healthy"
					} else {
						status = "⚠ No API Key"
					}
				}
				
				fmt.Printf("%s: %s\n", name, status)
			}
		}
		
		return nil
	},
}

// providerEnableCmd enables a provider
var providerEnableCmd = &cobra.Command{
	Use:   "enable <provider-name>",
	Short: "Enable a provider",
	Long: `Enable a provider for use.

This allows the router to use this provider for requests.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		provider, exists := cfg.Providers.GetProviders()[name]
		if !exists {
			return fmt.Errorf("provider not found: %s", name)
		}

		provider.Enabled = true
		cfg.Providers.SetProvider(name, provider)
		
		// Save config
		configPath := globalOpts.ConfigFile
		if configPath == "" {
			configPath = "./config.yaml"
		}
		
		if err := SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Enabled provider: %s\n", name)
		return nil
	},
}

// providerDisableCmd disables a provider
var providerDisableCmd = &cobra.Command{
	Use:   "disable <provider-name>",
	Short: "Disable a provider",
	Long: `Disable a provider from use.

The router will not use this provider for requests.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		
		cfg, err := GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		provider, exists := cfg.Providers.GetProviders()[name]
		if !exists {
			return fmt.Errorf("provider not found: %s", name)
		}

		provider.Enabled = false
		cfg.Providers.SetProvider(name, provider)
		
		// Save config
		configPath := globalOpts.ConfigFile
		if configPath == "" {
			configPath = "./config.yaml"
		}
		
		if err := SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Disabled provider: %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerHealthCmd)
	providerCmd.AddCommand(providerEnableCmd)
	providerCmd.AddCommand(providerDisableCmd)
}
