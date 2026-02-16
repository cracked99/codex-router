package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Zai        ZaiConfig        `yaml:"zai"`
	Codex      CodexConfig      `yaml:"codex"`
	Translator TranslatorConfig `yaml:"translator"`
	Session    SessionConfig    `yaml:"session"`
	Logging    LoggingConfig    `yaml:"logging"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// ZaiConfig contains z.ai API configuration
type ZaiConfig struct {
	BaseURL   string        `yaml:"base_url"`
	APIKey    string        `yaml:"api_key"`
	Timeout   time.Duration `yaml:"timeout"`
	MaxRetries int          `yaml:"max_retries"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// CodexConfig contains Codex CLI configuration
type CodexConfig struct {
	BaseURL       string `yaml:"base_url"`
	APIKeyHeader  string `yaml:"api_key_header"`
}

// TranslatorConfig contains translator configuration
type TranslatorConfig struct {
	Mode           string `yaml:"mode"`           // wasm | sidecar
	WasmPath       string `yaml:"wasm_path"`
	SidecarCommand string `yaml:"sidecar_command"`
}

// SessionConfig contains session management configuration
type SessionConfig struct {
	Enabled           bool          `yaml:"enabled"`
	TTL               time.Duration `yaml:"ttl"`
	MaxConversations  int           `yaml:"max_conversations"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug | info | warn | error
	Format string `yaml:"format"` // json | text
	File   string `yaml:"file"`   // Optional file output
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
	Format  string `yaml:"format"` // prometheus
}

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

	// Override with environment variables
	if apiKey := os.Getenv("ZAI_API_KEY"); apiKey != "" {
		cfg.Zai.APIKey = apiKey
	}
	if apiKey := os.Getenv("CODEX_ROUTER_API_KEY"); apiKey != "" {
		cfg.Zai.APIKey = apiKey
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

	if c.Zai.BaseURL == "" {
		return fmt.Errorf("z.ai base URL is required")
	}

	if c.Zai.APIKey == "" {
		return fmt.Errorf("z.ai API key is required (set via ZAI_API_KEY environment variable or config file)")
	}

	if c.Translator.Mode != "wasm" && c.Translator.Mode != "sidecar" {
		return fmt.Errorf("invalid translator mode: %s (must be 'wasm' or 'sidecar')", c.Translator.Mode)
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
		Codex: CodexConfig{
			BaseURL:      "",
			APIKeyHeader: "Authorization",
		},
		Translator: TranslatorConfig{
			Mode:           "wasm",
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
