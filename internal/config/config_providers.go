package config

import "time"

// Config represents the application configuration with provider support
type Config struct {
	Server          ServerConfig          `yaml:"server" mapstructure:"server"`
	Zai             ZaiConfig             `yaml:"zai" mapstructure:"zai"` // Legacy, will be deprecated
	Providers       ProvidersConfig       `yaml:"providers" mapstructure:"providers"`
	Codex           CodexConfig           `yaml:"codex" mapstructure:"codex"`
	Translator      TranslatorConfig      `yaml:"translator" mapstructure:"translator"`
	Session         SessionConfig         `yaml:"session" mapstructure:"session"`
	Logging         LoggingConfig         `yaml:"logging" mapstructure:"logging"`
	Metrics         MetricsConfig         `yaml:"metrics" mapstructure:"metrics"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host string    `yaml:"host" mapstructure:"host"`
	Port int       `yaml:"port" mapstructure:"port"`
	TLS  TLSConfig `yaml:"tls" mapstructure:"tls"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	CertFile string `yaml:"cert_file" mapstructure:"cert_file"`
	KeyFile  string `yaml:"key_file" mapstructure:"key_file"`
}

// ZaiConfig contains z.ai API configuration (legacy)
type ZaiConfig struct {
	BaseURL    string        `yaml:"base_url" mapstructure:"base_url"`
	APIKey     string        `yaml:"api_key" mapstructure:"api_key"`
	Timeout    time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxRetries int           `yaml:"max_retries" mapstructure:"max_retries"`
	RetryDelay time.Duration `yaml:"retry_delay" mapstructure:"retry_delay"`
}

// CodexConfig contains Codex CLI configuration
type CodexConfig struct {
	BaseURL      string `yaml:"base_url" mapstructure:"base_url"`
	APIKeyHeader string `yaml:"api_key_header" mapstructure:"api_key_header"`
}

// TranslatorConfig contains translator configuration
type TranslatorConfig struct {
	Mode           string `yaml:"mode" mapstructure:"mode"` // wasm | sidecar | native
	WasmPath       string `yaml:"wasm_path" mapstructure:"wasm_path"`
	SidecarCommand string `yaml:"sidecar_command" mapstructure:"sidecar_command"`
}

// SessionConfig contains session management configuration
type SessionConfig struct {
	Enabled          bool          `yaml:"enabled" mapstructure:"enabled"`
	TTL              time.Duration `yaml:"ttl" mapstructure:"ttl"`
	MaxConversations int           `yaml:"max_conversations" mapstructure:"max_conversations"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" mapstructure:"level"`   // debug | info | warn | error
	Format string `yaml:"format" mapstructure:"format"` // json | text
	File   string `yaml:"file" mapstructure:"file"`     // Optional file output
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
	Format  string `yaml:"format" mapstructure:"format"` // prometheus
}
