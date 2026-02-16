package api

import "time"

// Responses API types as per OpenAI specification

// ResponseRequest represents a request to the Responses API
type ResponseRequest struct {
	Model              string          `json:"model"`
	Input              interface{}     `json:"input"` // string or []InputItem
	Instructions       string          `json:"instructions,omitempty"`
	Temperature        *float64        `json:"temperature,omitempty"`
	MaxOutputTokens    *int            `json:"max_output_tokens,omitempty"`
	TopP              *float64        `json:"top_p,omitempty"`
	Tools             []Tool          `json:"tools,omitempty"`
	ToolChoice        interface{}     `json:"tool_choice,omitempty"` // string or ToolChoice
	Stream            bool            `json:"stream,omitempty"`
	PreviousResponseID string         `json:"previous_response_id,omitempty"`
	Conversation      *Conversation   `json:"conversation,omitempty"`
	Include           []string        `json:"include,omitempty"`
	Metadata          map[string]any  `json:"metadata,omitempty"`
}

// InputItem represents an item in the input array
type InputItem struct {
	Type      string                 `json:"type"` // message, input_text, input_image
	Role      string                 `json:"role,omitempty"`
	Content   []ContentBlock         `json:"content,omitempty"`
	Text      string                 `json:"text,omitempty"`
	ImageURL  string                 `json:"image_url,omitempty"`
}

// ContentBlock represents a block of content
type ContentBlock struct {
	Type      string                 `json:"type"` // output_text, input_text, image_url, etc.
	Text      string                 `json:"text,omitempty"`
	ImageURL  *ImageURL              `json:"image_url,omitempty"`
}

// ImageURL represents an image URL with optional detail level
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// Tool represents a function/tool definition
type Tool struct {
	Type       string                 `json:"type"` // function
	Function   *FunctionDefinition    `json:"function,omitempty"`
}

// FunctionDefinition defines a function that can be called
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolChoice specifies how tools should be used
type ToolChoice struct {
	Type     string                  `json:"type"` // auto, none, required, function
	Function *FunctionCall           `json:"function,omitempty"`
}

// FunctionCall represents a function to be called
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Conversation represents conversation state
type Conversation struct {
	ID        string       `json:"id"`
	Metadata  Metadata      `json:"metadata,omitempty"`
}

// Metadata contains additional metadata
type Metadata struct {
	UserID    string `json:"user_id,omitempty"`
}

// Response represents a Responses API response
type Response struct {
	ID           string        `json:"id"`
	Object       string        `json:"object"` // "response"
	CreatedAt    int64         `json:"created_at"`
	Status       string        `json:"status"` // "in_progress", "completed", "failed"
	Error        *ResponseError `json:"error,omitempty"`
	Result       *ResponseResult `json:"result,omitempty"`
}

// ResponseError represents an error in the response
type ResponseError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ResponseResult contains the actual response data
type ResponseResult struct {
	ID         string       `json:"id"`
	Status     string       `json:"status"`
	Output     []OutputItem `json:"output,omitempty"`
	Metadata   Metadata     `json:"metadata,omitempty"`
	Usage      Usage        `json:"usage,omitempty"`
}

// OutputItem represents an item in the output array
type OutputItem struct {
	Type      string          `json:"type"` // message, tool_call
	Role      string          `json:"role,omitempty"`
	Content   []ContentBlock  `json:"content,omitempty"`
	ToolCalls []ToolCallItem  `json:"tool_calls,omitempty"`
}

// ToolCallItem represents a tool call in the output
type ToolCallItem struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // function
	Function *FunctionCall          `json:"function"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ResponseStreamEvent represents an SSE event for streaming
type ResponseStreamEvent struct {
	Event string      `json:"event"` // response.delta, response.done, etc.
	Data  interface{} `json:"data"`
}

// ResponseDelta represents a delta in a streaming response
type ResponseDelta struct {
	Type string `json:"type"` // output_text, tool_call, etc.
	Text string `json:"text,omitempty"`
	// Tool call fields would be added here
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// Helper function to get current timestamp
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}
