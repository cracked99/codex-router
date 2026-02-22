package providers

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages provider instances
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	order     []string // Provider order by priority
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		order:     []string{},
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider Provider, config ProviderConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Initialize provider
	if err := provider.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize provider %s: %w", config.Name, err)
	}

	// Store provider
	r.providers[config.Name] = provider

	// Update order based on priority
	r.updateOrder(config.Name, config.Priority, config.Enabled)

	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	return provider, exists
}

// GetByModel finds a provider that supports the given model
func (r *Registry) GetByModel(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check providers in priority order
	for _, name := range r.order {
		provider := r.providers[name]
		if provider.SupportsModel(model) {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no provider supports model: %s", model)
}

// GetDefault returns the highest priority enabled provider
func (r *Registry) GetDefault() (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.order) == 0 {
		return nil, fmt.Errorf("no providers registered")
	}

	// Return first provider in order
	return r.providers[r.order[0]], nil
}

// GetAll returns all registered providers
func (r *Registry) GetAll() map[string]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Provider)
	for k, v := range r.providers {
		result[k] = v
	}
	return result
}

// List returns provider names in priority order
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, len(r.order))
	copy(result, r.order)
	return result
}

// Unregister removes a provider from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider, exists := r.providers[name]
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	// Shutdown provider
	if err := provider.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown provider %s: %w", name, err)
	}

	// Remove from registry
	delete(r.providers, name)

	// Update order
	newOrder := []string{}
	for _, n := range r.order {
		if n != name {
			newOrder = append(newOrder, n)
		}
	}
	r.order = newOrder

	return nil
}

// HealthCheck runs health checks on all providers
func (r *Registry) HealthCheck() map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, provider := range r.providers {
		results[name] = provider.HealthCheck(context.Background())
	}
	return results
}

// updateOrder maintains provider order by priority
func (r *Registry) updateOrder(name string, priority int, enabled bool) {
	// Remove if already in order
	newOrder := []string{}
	for _, n := range r.order {
		if n != name {
			newOrder = append(newOrder, n)
		}
	}

	// Only add if enabled
	if !enabled {
		r.order = newOrder
		return
	}

	// Insert at correct position based on priority
	// Lower priority number = higher priority
	inserted := false
	for i, n := range newOrder {
		// Get priority of existing provider (would need to store this)
		// For now, just append
		if i >= priority-1 {
			newOrder = append(newOrder[:i], append([]string{name}, newOrder[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		newOrder = append(newOrder, name)
	}

	r.order = newOrder
}
