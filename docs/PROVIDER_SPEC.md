# Provider Agnostic Architecture Specification

## Overview

Implement a provider-agnostic architecture that allows codex-router to work with multiple LLM backends (OpenAI, z.ai, Anthropic, etc.) through a unified interface.

## Goals

1. **Provider Neutrality**: Support multiple LLM providers transparently
2. **Easy Extension**: Add new providers with minimal code changes
3. **Runtime Selection**: Choose provider based on model/config
4. **Health Monitoring**: Per-provider health checks
5. **Failover Support**: Automatic failover between providers

## Architecture

### Hexagonal Architecture Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                    Codex API Router                         │
├─────────────────────────────────────────────────────────────┤
│  Domain Layer (Core)                                        │
│  ├─ Provider Interface (Port)                              │
│  ├─ Request/Response Models                                │
│  └─ Transformation Logic                                   │
├─────────────────────────────────────────────────────────────┤
│  Application Layer                                          │
│  ├─ Provider Registry                                       │
│  ├─ Provider Factory                                        │
│  ├─ Provider Router (selects based on model)               │
│  └─ Health Monitor                                          │
├─────────────────────────────────────────────────────────────┤
│  Infrastructure Layer (Adapters)                            │
│  ├─ OpenAI Provider Adapter                                │
│  ├─ z.ai Provider Adapter                                  │
│  ├─ Anthropic Provider Adapter                             │
│  └─ Custom Provider Adapter                                │
└─────────────────────────────────────────────────────────────┘
```

## Provider Interface (Port)

```go
// Provider defines the interface for LLM backends
type Provider interface {
    // Identity
    Name() string
    Type() ProviderType
    
    // Lifecycle
    Initialize(config ProviderConfig) error
    Shutdown() error
    
    // Core Operations
    TransformRequest(req *ResponsesRequest) (*ProviderRequest, error)
    TransformResponse(resp *ProviderResponse) (*ResponsesResponse, error)
    TransformStreamEvent(event *ProviderStreamEvent) (*ResponsesStreamEvent, error)
    
    // Execution
    Execute(ctx context.Context, req *ProviderRequest) (*ProviderResponse, error)
    ExecuteStream(ctx context.Context, req *ProviderRequest) (<-chan ProviderStreamEvent, error)
    
    // Capabilities
    SupportsModel(model string) bool
    SupportsStreaming() bool
    SupportsTools() bool
    GetModels() []string
    
    // Health
    HealthCheck(ctx context.Context) error
    GetMetrics() ProviderMetrics
}

type ProviderType string

const (
    ProviderTypeOpenAI    ProviderType = "openai"
    ProviderTypeZai       ProviderType = "zai"
    ProviderTypeAnthropic ProviderType = "anthropic"
    ProviderTypeCustom    ProviderType = "custom"
)
```

## Configuration Schema

```yaml
providers:
  # Multiple provider configurations
  zai:
    enabled: true
    type: "zai"
    priority: 1
    base_url: "https://api.z.ai/api/paas/v4"
    api_key: "${ZAI_API_KEY}"
    models:
      - "glm-*"  # Wildcard support
      - "chatglm-*"
    health_check:
      enabled: true
      interval: 30s
      timeout: 5s
      
  openai:
    enabled: true
    type: "openai"
    priority: 2
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"
    models:
      - "gpt-*"
      - "o1-*"
      
  anthropic:
    enabled: false
    type: "anthropic"
    priority: 3
    base_url: "https://api.anthropic.com/v1"
    api_key: "${ANTHROPIC_API_KEY}"
    models:
      - "claude-*"

# Default provider selection strategy
provider_strategy: "priority"  # priority | round_robin | model_match

# Fallback configuration
fallback:
  enabled: true
  timeout: 30s
  retry_count: 3
```

## Model Mapping

```yaml
model_mapping:
  # Map Responses API models to provider models
  "gpt-4.1": "glm-5"
  "gpt-4.1-mini": "glm-5-flash"
  "gpt-4": "gpt-4-turbo"
  "claude-3-5-sonnet": "claude-3-5-sonnet-20241022"
```

## Implementation Phases

### Phase 1: Core Interface
- Define Provider interface
- Create provider registry
- Implement provider factory

### Phase 2: Providers
- OpenAI provider adapter
- z.ai provider adapter  
- Anthropic provider adapter

### Phase 3: Features
- Provider health monitoring
- Model-based routing
- Failover support
- Metrics collection

### Phase 4: CLI
- Provider management commands
- Health check commands
- Model mapping commands

## Provider Adapter Structure

```go
// internal/providers/provider.go
package providers

type BaseProvider struct {
    name    string
    config  ProviderConfig
    client  *http.Client
    metrics ProviderMetrics
}

type ProviderConfig struct {
    Type        ProviderType
    BaseURL     string
    APIKey      string
    Timeout     time.Duration
    MaxRetries  int
    Models      []string
    HealthCheck HealthCheckConfig
}

type ProviderMetrics struct {
    RequestsTotal     int64
    RequestsSuccess   int64
    RequestsFailed    int64
    AverageLatency    time.Duration
    LastHealthCheck   time.Time
    HealthStatus      HealthStatus
}
```

## Request Flow

1. **Incoming Request** → Responses API format
2. **Provider Selection** → Based on model/config
3. **Request Transformation** → Provider-specific format
4. **Execution** → Provider API call
5. **Response Transformation** → Responses API format
6. **Return** → Unified response

## Error Handling

```go
type ProviderError struct {
    Provider    string
    Code        string
    Message     string
    HTTPStatus  int
    Retryable   bool
    Fallback    bool  // Can fallback to another provider
}
```

## Health Monitoring

```go
type HealthStatus struct {
    Status          HealthState  // healthy | degraded | unhealthy
    LastCheck       time.Time
    Latency         time.Duration
    ErrorRate       float64
    ConsecutiveFail int
}
```

## CLI Commands

```bash
# List providers
codex-router provider list

# Check provider health
codex-router provider health zai

# Test provider
codex-router provider test zai --model glm-5

# Enable/disable provider
codex-router provider enable openai
codex-router provider disable anthropic

# Show provider metrics
codex-router provider metrics zai
```

## Benefits

1. **Flexibility**: Switch providers without code changes
2. **Reliability**: Automatic failover
3. **Cost Optimization**: Route to cheapest provider
4. **Performance**: Route to fastest provider
5. **Testing**: Mock providers for testing
6. **Future-Proof**: Easy to add new providers

## Migration Path

1. **Phase 1**: Refactor existing z.ai code into provider adapter
2. **Phase 2**: Add OpenAI and Anthropic adapters
3. **Phase 3**: Update configuration schema
4. **Phase 4**: Add CLI commands
5. **Phase 5**: Update documentation

## Testing Strategy

- Unit tests for each provider adapter
- Integration tests with provider APIs
- Mock providers for end-to-end testing
- Health check validation
- Failover testing
