package providers

import (
	"fmt"
)

// Factory creates provider instances
type Factory struct {
	registry *Registry
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		registry: NewRegistry(),
	}
}

// CreateProvider creates a provider based on type
func (f *Factory) CreateProvider(providerType string) (Provider, error) {
	switch providerType {
	case "zai":
		return NewZaiProvider(), nil
	case "openai":
		return NewOpenAIProvider(), nil
	case "anthropic":
		return nil, fmt.Errorf("anthropic provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// RegisterProvider creates and registers a provider
func (f *Factory) RegisterProvider(config ProviderConfig) error {
	// Create provider
	provider, err := f.CreateProvider(config.Type)
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", config.Name, err)
	}

	// Register with registry
	if err := f.registry.Register(provider, config); err != nil {
		return fmt.Errorf("failed to register provider %s: %w", config.Name, err)
	}

	return nil
}

// GetRegistry returns the provider registry
func (f *Factory) GetRegistry() *Registry {
	return f.registry
}

// InitializeProviders initializes all providers from config
func (f *Factory) InitializeProviders(configs map[string]ProviderConfig) error {
	for name, config := range configs {
		if !config.Enabled {
			continue
		}

		config.Name = name
		if err := f.RegisterProvider(config); err != nil {
			return fmt.Errorf("failed to initialize provider %s: %w", name, err)
		}
	}

	return nil
}

// GetProvider retrieves a provider by name
func (f *Factory) GetProvider(name string) (Provider, error) {
	provider, exists := f.registry.Get(name)
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// GetProviderForModel finds a provider that supports the model
func (f *Factory) GetProviderForModel(model string) (Provider, error) {
	return f.registry.GetByModel(model)
}

// GetDefaultProvider returns the default (highest priority) provider
func (f *Factory) GetDefaultProvider() (Provider, error) {
	return f.registry.GetDefault()
}

// ListProviders returns list of registered providers
func (f *Factory) ListProviders() []string {
	return f.registry.List()
}

// HealthCheckAll runs health checks on all providers
func (f *Factory) HealthCheckAll() map[string]error {
	return f.registry.HealthCheck()
}
