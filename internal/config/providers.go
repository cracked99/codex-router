package config

import (
	"time"
)

// ProvidersConfig contains multiple provider configurations
type ProvidersConfig struct {
	Zai             ProviderConfig `yaml:"zai" mapstructure:"zai"`
	OpenAI          ProviderConfig `yaml:"openai" mapstructure:"openai"`
	Anthropic       ProviderConfig `yaml:"anthropic,omitempty" mapstructure:"anthropic,omitempty"`
	ProviderStrategy string        `yaml:"provider_strategy" mapstructure:"provider_strategy"`
	Fallback        FallbackConfig `yaml:"fallback" mapstructure:"fallback"`
	ModelMapping    map[string]string `yaml:"model_mapping" mapstructure:"model_mapping"`
}

// ProviderConfig contains provider-specific configuration
type ProviderConfig struct {
	Enabled     bool              `yaml:"enabled" mapstructure:"enabled"`
	Type        string            `yaml:"type" mapstructure:"type"`
	Priority    int               `yaml:"priority" mapstructure:"priority"`
	BaseURL     string            `yaml:"base_url" mapstructure:"base_url"`
	APIKey      string            `yaml:"api_key" mapstructure:"api_key"`
	Timeout     time.Duration     `yaml:"timeout" mapstructure:"timeout"`
	MaxRetries  int               `yaml:"max_retries" mapstructure:"max_retries"`
	RetryDelay  time.Duration     `yaml:"retry_delay" mapstructure:"retry_delay"`
	Models      []string          `yaml:"models" mapstructure:"models"`
	HealthCheck HealthCheckConfig `yaml:"health_check" mapstructure:"health_check"`
}

// HealthCheckConfig for provider health monitoring
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled" mapstructure:"enabled"`
	Interval time.Duration `yaml:"interval" mapstructure:"interval"`
	Timeout  time.Duration `yaml:"timeout" mapstructure:"timeout"`
	Endpoint string        `yaml:"endpoint" mapstructure:"endpoint"`
}

// FallbackConfig for provider failover
type FallbackConfig struct {
	Enabled    bool          `yaml:"enabled" mapstructure:"enabled"`
	Timeout    time.Duration `yaml:"timeout" mapstructure:"timeout"`
	RetryCount int           `yaml:"retry_count" mapstructure:"retry_count"`
}

// GetProviders returns all providers as a map for compatibility
func (pc *ProvidersConfig) GetProviders() map[string]ProviderConfig {
	providers := make(map[string]ProviderConfig)
	if pc.Zai.Enabled || pc.Zai.APIKey != "" {
		providers["zai"] = pc.Zai
	}
	if pc.OpenAI.Enabled || pc.OpenAI.APIKey != "" {
		providers["openai"] = pc.OpenAI
	}
	if pc.Anthropic.Enabled || pc.Anthropic.APIKey != "" {
		providers["anthropic"] = pc.Anthropic
	}
	return providers
}

// SetProvider sets a provider configuration
func (pc *ProvidersConfig) SetProvider(name string, config ProviderConfig) {
	switch name {
	case "zai":
		pc.Zai = config
	case "openai":
		pc.OpenAI = config
	case "anthropic":
		pc.Anthropic = config
	}
}

// DefaultProvidersConfig returns default provider configurations
func DefaultProvidersConfig() ProvidersConfig {
	return ProvidersConfig{
		Zai: ProviderConfig{
			Enabled:    true,
			Type:       "zai",
			Priority:   1,
			BaseURL:    "https://api.z.ai/api/coding/paas/v4", // Coding Plan endpoint
			Timeout:    120 * time.Second,
			MaxRetries: 3,
			RetryDelay: 1 * time.Second,
			Models:     []string{"glm-5", "glm-4.7", "glm-4.7-flash", "glm-4.5-air"},
			HealthCheck: HealthCheckConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
			},
		},
		OpenAI: ProviderConfig{
			Enabled:    false,
			Type:       "openai",
			Priority:   2,
			BaseURL:    "https://api.openai.com/v1",
			Timeout:    120 * time.Second,
			MaxRetries: 3,
			RetryDelay: 1 * time.Second,
			Models:     []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
			HealthCheck: HealthCheckConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
			},
		},
		Anthropic: ProviderConfig{
			Enabled:    false,
			Type:       "anthropic",
			Priority:   3,
			BaseURL:    "https://api.anthropic.com/v1",
			Timeout:    120 * time.Second,
			MaxRetries: 3,
			RetryDelay: 1 * time.Second,
			Models:     []string{"claude-3-opus", "claude-3-sonnet"},
			HealthCheck: HealthCheckConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
			},
		},
		ProviderStrategy: "priority",
		Fallback: FallbackConfig{
			Enabled:    true,
			Timeout:    30 * time.Second,
			RetryCount: 2,
		},
		// Default model mapping for z.ai Coding Plan
		// Maps Claude/Codex model names to z.ai equivalents
		ModelMapping: map[string]string{
			// Codex CLI models -> glm-5
			"gpt-5.2-codex":       "glm-5",
			"gpt-5.1-codex-max":   "glm-5",
			"gpt-5.2":             "glm-5",
			"gpt-5.1-codex-mini":  "glm-5",
			// Claude models -> glm-5
			"claude-opus-4":       "glm-5",
			"claude-opus-4-20250514": "glm-5",
			"claude-sonnet-4":     "glm-5",
			"claude-sonnet-4-20250514": "glm-5",
			"claude-3-5-sonnet":   "glm-5",
			"claude-3-5-sonnet-20241022": "glm-5",
			"claude-3-5-haiku":    "glm-5",
			"claude-3-5-haiku-20241022": "glm-5",
			"claude-3-haiku":      "glm-5",
			"claude-3-opus":       "glm-5",
			"claude-3-sonnet":     "glm-5",
			// Common aliases
			"opus":  "glm-5",
			"sonnet": "glm-5",
			"haiku": "glm-5",
		},
	}
}
