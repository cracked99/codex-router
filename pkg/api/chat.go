package api

// Chat Completions API types as per OpenAI specification

// ChatCompletionRequest represents a request to the Chat Completions API
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []ChatMessage          `json:"messages"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Tools            []ChatTool             `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"` // string or ChatToolChoice
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string          `json:"role"` // system, user, assistant
	Content interface{}     `json:"content"` // string or []ChatContentPart
}

// ChatContentPart represents a part of message content (for multimodal)
type ChatContentPart struct {
	Type     string     `json:"type"` // text, image_url
	Text     string     `json:"text,omitempty"`
	ImageURL *ImageURL  `json:"image_url,omitempty"`
}

// ChatTool represents a tool in the chat completion API
type ChatTool struct {
	Type     string                 `json:"type"` // function
	Function *ChatFunction          `json:"function"`
}

// ChatFunction defines a function for chat completion
type ChatFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ChatToolChoice specifies how tools should be used in chat
type ChatToolChoice struct {
	Type     string                 `json:"type"` // auto, none, required, function
	Function *ChatFunctionCall      `json:"function,omitempty"`
}

// ChatFunctionCall represents a function call in chat
type ChatFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponse represents a response from the Chat Completions API
type ChatCompletionResponse struct {
	ID      string                    `json:"id"`
	Object  string                    `json:"object"` // "chat.completion"
	Created int64                     `json:"created"`
	Model   string                    `json:"model"`
	Choices []ChatCompletionChoice    `json:"choices"`
	Usage   ChatUsage                 `json:"usage"`
}

// ChatCompletionChoice represents a choice in the completion
type ChatCompletionChoice struct {
	Index        int                      `json:"index"`
	Message      ChatCompletionMessage    `json:"message"`
	FinishReason string                   `json:"finish_reason"` // stop, length, tool_calls, content_filter
}

// ChatCompletionMessage represents a message in the completion
type ChatCompletionMessage struct {
	Role       string                `json:"role"`
	Content    string                `json:"content,omitempty"`
	ToolCalls  []ChatToolCallItem    `json:"tool_calls,omitempty"`
}

// ChatToolCallItem represents a tool call in the completion
type ChatToolCallItem struct {
	Index    int                  `json:"index"`
	ID       string               `json:"id"`
	Type     string               `json:"type"` // function
	Function ChatFunctionCall     `json:"function"`
}

// ChatUsage represents token usage for chat completion
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionStreamChunk represents a chunk in a streaming response
type ChatCompletionStreamChunk struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"` // "chat.completion.chunk"
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}

// ChatCompletionStreamChoice represents a choice in a streaming chunk
type ChatCompletionStreamChoice struct {
	Index        int                      `json:"index"`
	Delta        ChatCompletionDelta      `json:"delta"`
	FinishReason *string                  `json:"finish_reason,omitempty"`
}

// ChatCompletionDelta represents the delta in a streaming chunk
type ChatCompletionDelta struct {
	Role      string                `json:"role,omitempty"`
	Content   string                `json:"content,omitempty"`
	ToolCalls []ChatToolCallDelta   `json:"tool_calls,omitempty"`
}

// ChatToolCallDelta represents a tool call delta in streaming
type ChatToolCallDelta struct {
	Index    int     `json:"index"`
	ID       *string `json:"id,omitempty"`
	Type     *string `json:"type,omitempty"`
	Function *struct {
		Name      *string `json:"name,omitempty"`
		Arguments *string `json:"arguments,omitempty"`
	} `json:"function,omitempty"`
}

// ErrorResponse represents an error from the API
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}
