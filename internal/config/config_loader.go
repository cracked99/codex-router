package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from file or environment variables
func Load(configPath string) (*Config, error) {
	cfg := Default()

	// Try to load from file
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	} else {
		// Try default location
		homeDir, err := os.UserHomeDir()
		if err == nil {
			defaultPath := filepath.Join(homeDir, ".codex-router", "config.yaml")
			if data, err := os.ReadFile(defaultPath); err == nil {
				if err := yaml.Unmarshal(data, cfg); err != nil {
					return nil, fmt.Errorf("failed to parse config file: %w", err)
				}
			}
		}
	}

	// Migrate legacy Zai config to providers if providers not set
	if len(cfg.Providers.GetProviders()) == 0 && cfg.Zai.APIKey != "" {
		defaults := DefaultProvidersConfig()
		cfg.Providers = defaults
		zaiProvider := defaults.Zai
		zaiProvider.APIKey = cfg.Zai.APIKey
		zaiProvider.BaseURL = cfg.Zai.BaseURL
		zaiProvider.Timeout = cfg.Zai.Timeout
		zaiProvider.MaxRetries = cfg.Zai.MaxRetries
		zaiProvider.RetryDelay = cfg.Zai.RetryDelay
		cfg.Providers.SetProvider("zai", zaiProvider)
	}

	// Override with environment variables (support both ZAI_API_KEY and Z_AI_API_KEY)
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("Z_AI_API_KEY")
	}
	if apiKey != "" {
		if len(cfg.Providers.GetProviders()) == 0 {
			cfg.Providers = DefaultProvidersConfig()
		}
		if _, exists := cfg.Providers.GetProviders()["zai"]; !exists {
			defaults := DefaultProvidersConfig()
			cfg.Providers.Zai = defaults.Zai
		}
		zaiProvider := cfg.Providers.Zai
		zaiProvider.APIKey = apiKey
		cfg.Providers.SetProvider("zai", zaiProvider)
		cfg.Zai.APIKey = apiKey // Legacy
	}

	// Load OpenAI API key from env
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		if _, exists := cfg.Providers.GetProviders()["openai"]; !exists {
			defaults := DefaultProvidersConfig()
			cfg.Providers.OpenAI = defaults.OpenAI
		}
		openaiProvider := cfg.Providers.OpenAI
		openaiProvider.APIKey = apiKey
		openaiProvider.Enabled = true
		cfg.Providers.SetProvider("openai", openaiProvider)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Check if at least one provider is configured
	hasProvider := false
	for _, provider := range c.Providers.GetProviders() {
		if provider.Enabled && provider.APIKey != "" {
			hasProvider = true
			break
		}
	}

	// Fallback to legacy config
	if !hasProvider && c.Zai.APIKey != "" {
		hasProvider = true
	}

	if !hasProvider {
		return fmt.Errorf("at least one provider must be configured with an API key")
	}

	if c.Translator.Mode != "wasm" && c.Translator.Mode != "sidecar" && c.Translator.Mode != "native" {
		return fmt.Errorf("invalid translator mode: %s (must be 'wasm', 'sidecar', or 'native')", c.Translator.Mode)
	}

	return nil
}

// Save saves configuration to a file
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Zai: ZaiConfig{
			BaseURL:    "https://api.z.ai/api/paas/v4",
			APIKey:     "",
			Timeout:    120 * time.Second,
			MaxRetries: 3,
			RetryDelay: 1 * time.Second,
		},
		Providers:    DefaultProvidersConfig(),
		Codex: CodexConfig{
			BaseURL:      "",
			APIKeyHeader: "Authorization",
		},
		Translator: TranslatorConfig{
			Mode:           "native",
			WasmPath:       "./translator.wasm",
			SidecarCommand: "node ./translator/index.js",
		},
		Session: SessionConfig{
			Enabled:          true,
			TTL:              3600 * time.Second,
			MaxConversations: 1000,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			File:   "",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
			Format:  "prometheus",
		},
	}
}
