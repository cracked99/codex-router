package providers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	name    string
	config  ProviderConfig
	client  *http.Client
	metrics ProviderMetrics
	mu      sync.RWMutex
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{
		name: name,
		metrics: ProviderMetrics{
			HealthStatus: HealthStateHealthy,
		},
	}
}

// Initialize initializes the base provider
func (p *BaseProvider) Initialize(config ProviderConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

	// Create HTTP client
	p.client = &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return nil
}

// Shutdown cleans up resources
func (p *BaseProvider) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		p.client.CloseIdleConnections()
	}

	return nil
}

// Name returns the provider name
func (p *BaseProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *BaseProvider) Type() ProviderType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config.Type
}

// SupportsModel checks if this provider supports a model
func (p *BaseProvider) SupportsModel(model string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, pattern := range p.config.Models {
		// Support wildcard patterns
		matched, err := filepath.Match(pattern, model)
		if err == nil && matched {
			return true
		}

		// Support prefix matching
		if strings.HasSuffix(pattern, "-*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(model, prefix) {
				return true
			}
		}

		// Exact match
		if pattern == model {
			return true
		}
	}

	return false
}

// GetModels returns list of supported models
func (p *BaseProvider) GetModels() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	models := make([]string, len(p.config.Models))
	copy(models, p.config.Models)
	return models
}

// SupportsStreaming returns whether streaming is supported
func (p *BaseProvider) SupportsStreaming() bool {
	return true // Most providers support streaming
}

// SupportsTools returns whether tool calling is supported
func (p *BaseProvider) SupportsTools() bool {
	return true // Most providers support tools
}

// HealthCheck performs a health check
func (p *BaseProvider) HealthCheck(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update last check time
	p.metrics.LastHealthCheck = time.Now()

	// Simple health check - try to connect to base URL
	if p.client == nil {
		p.metrics.HealthStatus = HealthStateUnhealthy
		return fmt.Errorf("client not initialized")
	}

	// Create a simple GET request to health endpoint
	healthURL := p.config.BaseURL
	if p.config.HealthCheck.Endpoint != "" {
		healthURL = p.config.HealthCheck.Endpoint
	}

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		p.metrics.HealthStatus = HealthStateUnhealthy
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	start := time.Now()
	resp, err := p.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		p.metrics.HealthStatus = HealthStateUnhealthy
		p.metrics.ConsecutiveFail++
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Update metrics
	p.metrics.ConsecutiveFail = 0
	p.metrics.HealthStatus = HealthStateHealthy
	p.metrics.LastHealthCheck = time.Now()
	p.metrics.AverageLatency = latency

	return nil
}

// GetMetrics returns provider metrics
func (p *BaseProvider) GetMetrics() ProviderMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.metrics
}

// RecordRequest records a request in metrics
func (p *BaseProvider) RecordRequest(success bool, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	atomic.AddInt64(&p.metrics.RequestsTotal, 1)
	if success {
		atomic.AddInt64(&p.metrics.RequestsSuccess, 1)
	} else {
		atomic.AddInt64(&p.metrics.RequestsFailed, 1)
	}

	p.metrics.LastRequestTime = time.Now()

	// Update average latency (simple moving average)
	if p.metrics.RequestsTotal == 1 {
		p.metrics.AverageLatency = latency
	} else {
		// Simple exponential moving average
		p.metrics.AverageLatency = time.Duration(
			float64(p.metrics.AverageLatency)*0.9 + float64(latency)*0.1,
		)
	}

	// Update error rate
	if p.metrics.RequestsTotal > 0 {
		p.metrics.ErrorRate = float64(p.metrics.RequestsFailed) / float64(p.metrics.RequestsTotal)
	}

	// Update health status based on error rate
	if p.metrics.ErrorRate > 0.5 {
		p.metrics.HealthStatus = HealthStateUnhealthy
	} else if p.metrics.ErrorRate > 0.1 {
		p.metrics.HealthStatus = HealthStateDegraded
	} else {
		p.metrics.HealthStatus = HealthStateHealthy
	}
}

// GetClient returns the HTTP client
func (p *BaseProvider) GetClient() *http.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.client
}

// GetConfig returns the provider configuration
func (p *BaseProvider) GetConfig() ProviderConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}
