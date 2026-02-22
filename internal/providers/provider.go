package providers

import (
	"context"
	"time"
)

// ProviderType defines the type of LLM provider
type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeZai       ProviderType = "zai"
	ProviderTypeAnthropic ProviderType = "anthropic"
	ProviderTypeCustom    ProviderType = "custom"
)

// HealthState represents the health status of a provider
type HealthState string

const (
	HealthStateHealthy   HealthState = "healthy"
	HealthStateDegraded  HealthState = "degraded"
	HealthStateUnhealthy HealthState = "unhealthy"
)

// Provider defines the interface for LLM backends
type Provider interface {
	// Identity
	Name() string
	Type() ProviderType

	// Lifecycle
	Initialize(config ProviderConfig) error
	Shutdown() error

	// Core Operations
	TransformRequest(req *ResponsesRequest) (interface{}, error)
	TransformResponse(resp interface{}) (*ResponsesResponse, error)
	TransformStreamEvent(event interface{}) (*ResponsesStreamEvent, error)

	// Execution
	Execute(ctx context.Context, req interface{}) (interface{}, error)
	ExecuteStream(ctx context.Context, req interface{}) (<-chan interface{}, error)

	// Capabilities
	SupportsModel(model string) bool
	SupportsStreaming() bool
	SupportsTools() bool
	GetModels() []string

	// Health
	HealthCheck(ctx context.Context) error
	GetMetrics() ProviderMetrics
}

// ProviderConfig contains provider configuration
type ProviderConfig struct {
	Name        string
	Type        ProviderType
	Enabled     bool
	Priority    int
	BaseURL     string
	APIKey      string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	Models      []string
	HealthCheck HealthCheckConfig
}

// HealthCheckConfig contains health check configuration
type HealthCheckConfig struct {
	Enabled  bool
	Interval time.Duration
	Timeout  time.Duration
	Endpoint string
}

// ProviderMetrics contains provider performance metrics
type ProviderMetrics struct {
	RequestsTotal      int64
	RequestsSuccess    int64
	RequestsFailed     int64
	AverageLatency     time.Duration
	LastHealthCheck    time.Time
	HealthStatus       HealthState
	ConsecutiveFail    int
	ErrorRate          float64
	LastRequestTime    time.Time
}

// ResponsesRequest represents a Responses API request
type ResponsesRequest struct {
	Model              string                   `json:"model"`
	Input              interface{}              `json:"input"`
	Instructions       string                   `json:"instructions,omitempty"`
	Temperature        *float64                 `json:"temperature,omitempty"`
	MaxOutputTokens    *int                     `json:"max_output_tokens,omitempty"`
	TopP               *float64                 `json:"top_p,omitempty"`
	Tools              []Tool                   `json:"tools,omitempty"`
	ToolChoice         interface{}              `json:"tool_choice,omitempty"`
	ParallelToolCalls  bool                     `json:"parallel_tool_calls,omitempty"`
	Stream             bool                     `json:"stream,omitempty"`
	PreviousResponseID string                   `json:"previous_response_id,omitempty"`
	Conversation       *Conversation            `json:"conversation,omitempty"`
	Include            []string                 `json:"include,omitempty"`
	Metadata           map[string]interface{}   `json:"metadata,omitempty"`
}

// ResponsesResponse represents a Responses API response
type ResponsesResponse struct {
	ID        string        `json:"id"`
	Object    string        `json:"object"`
	CreatedAt int64         `json:"created_at"`
	Status    string        `json:"status"`
	Output    []OutputItem  `json:"output"`
	Error     *ResponseError `json:"error,omitempty"`
	Usage     *ResponseUsage `json:"usage,omitempty"`
}

// ResponsesStreamEvent represents a streaming event
type ResponsesStreamEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

// Tool represents a tool definition
type Tool struct {
	Type       string                 `json:"type"`
	Function   *FunctionDefinition    `json:"function,omitempty"`
}

// FunctionDefinition defines a function
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Conversation represents conversation context
type Conversation struct {
	ID       string `json:"id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OutputItem represents an item in the output
type OutputItem struct {
	Type      string       `json:"type"`
	Role      string       `json:"role,omitempty"`
	Content   []Content    `json:"content,omitempty"`
	ToolCalls []ToolCall   `json:"tool_calls,omitempty"`
}

// Content represents content in an output item
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolCall represents a tool call
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function *FunctionCall          `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ResponseError represents an error
type ResponseError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ResponseUsage represents token usage
type ResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ProviderError represents an error from a provider
type ProviderError struct {
	Provider   string `json:"provider"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"http_status"`
	Retryable  bool   `json:"retryable"`
	Fallback   bool   `json:"fallback"`
}

func (e *ProviderError) Error() string {
	return e.Message
}
