package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ZaiProvider implements Provider for z.ai backend
type ZaiProvider struct {
	*BaseProvider
}

// NewZaiProvider creates a new z.ai provider
func NewZaiProvider() *ZaiProvider {
	return &ZaiProvider{
		BaseProvider: NewBaseProvider("zai"),
	}
}

// Initialize initializes the z.ai provider
func (p *ZaiProvider) Initialize(config ProviderConfig) error {
	// Set defaults
	if config.BaseURL == "" {
		config.BaseURL = "https://api.z.ai/api/paas/v4"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if len(config.Models) == 0 {
		config.Models = []string{"glm-*", "chatglm-*"}
	}

	return p.BaseProvider.Initialize(config)
}

// TransformRequest transforms Responses API request to Chat Completions
func (p *ZaiProvider) TransformRequest(req *ResponsesRequest) (interface{}, error) {
	// Transform Responses API to Chat Completions API
	chatReq := map[string]interface{}{
		"model": p.mapModel(req.Model),
	}

	// Transform input to messages
	messages := []map[string]interface{}{}

	// Add system message if instructions present
	if req.Instructions != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": req.Instructions,
		})
	}

	// Transform input
	switch v := req.Input.(type) {
	case string:
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": v,
		})
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				msg := p.transformInputItem(itemMap)
				if msg != nil {
					messages = append(messages, msg)
				}
			}
		}
	}

	chatReq["messages"] = messages

	// Copy optional parameters
	if req.Temperature != nil {
		chatReq["temperature"] = *req.Temperature
	}
	if req.MaxOutputTokens != nil {
		chatReq["max_completion_tokens"] = *req.MaxOutputTokens
	}
	if req.TopP != nil {
		chatReq["top_p"] = *req.TopP
	}
	if req.Stream {
		chatReq["stream"] = true
	}

	// Transform tools if present
	if len(req.Tools) > 0 {
		chatReq["tools"] = p.transformTools(req.Tools)
	}

	return chatReq, nil
}

// TransformResponse transforms Chat Completions to Responses API
func (p *ZaiProvider) TransformResponse(resp interface{}) (*ResponsesResponse, error) {
	// Cast to Chat Completions response
	chatResp, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	// Transform to Responses API format
	responsesResp := &ResponsesResponse{
		ID:        fmt.Sprintf("resp_%d", time.Now().UnixNano()),
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Status:    "completed",
		Output:    []OutputItem{},
	}

	// Extract choices
	if choices, ok := chatResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			output := OutputItem{
				Type:    "message",
				Role:    "assistant",
				Content: []Content{},
			}

			// Extract message content
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					output.Content = append(output.Content, Content{
						Type: "output_text",
						Text: content,
					})
				}

				// Handle tool calls
				if toolCalls, ok := message["tool_calls"].([]interface{}); ok {
					for _, tc := range toolCalls {
						if tcMap, ok := tc.(map[string]interface{}); ok {
							toolCall := ToolCall{
								ID:   fmt.Sprintf("%v", tcMap["id"]),
								Type: "function",
								Function: &FunctionCall{
									Name:      fmt.Sprintf("%v", tcMap["function"].(map[string]interface{})["name"]),
									Arguments: tcMap["function"].(map[string]interface{})["arguments"].(map[string]interface{}),
								},
							}
							output.ToolCalls = append(output.ToolCalls, toolCall)
						}
					}
				}
			}

			responsesResp.Output = append(responsesResp.Output, output)
		}
	}

	// Extract usage
	if usage, ok := chatResp["usage"].(map[string]interface{}); ok {
		responsesResp.Usage = &ResponseUsage{
			InputTokens:  int(usage["prompt_tokens"].(float64)),
			OutputTokens: int(usage["completion_tokens"].(float64)),
			TotalTokens:  int(usage["total_tokens"].(float64)),
		}
	}

	return responsesResp, nil
}

// TransformStreamEvent transforms a streaming event
func (p *ZaiProvider) TransformStreamEvent(event interface{}) (*ResponsesStreamEvent, error) {
	// Handle Chat Completions streaming events
	chunk, ok := event.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	// Extract delta
	if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if content, ok := delta["content"].(string); ok && content != "" {
					return &ResponsesStreamEvent{
						Type: "response.output_text.delta",
						Data: map[string]interface{}{
							"delta": content,
						},
						Timestamp: time.Now().Unix(),
					}, nil
				}
			}

			// Check for finish reason
			if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
				return &ResponsesStreamEvent{
					Type: "response.completed",
					Data: map[string]interface{}{
						"status": p.mapFinishReason(finishReason),
					},
					Timestamp: time.Now().Unix(),
				}, nil
			}
		}
	}

	return nil, nil
}

// Execute executes a request to z.ai
func (p *ZaiProvider) Execute(ctx context.Context, req interface{}) (interface{}, error) {
	start := time.Now()

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		p.RecordRequest(false, 0)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	config := p.GetConfig()
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		config.BaseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		p.RecordRequest(false, 0)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Execute request
	client := p.GetClient()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		p.RecordRequest(false, time.Since(start))
		return nil, &ProviderError{
			Provider:   p.name,
			Code:       "api_error",
			Message:    string(respBody),
			HTTPStatus: httpResp.StatusCode,
			Retryable:  httpResp.StatusCode >= 500,
			Fallback:   httpResp.StatusCode >= 500,
		}
	}

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	p.RecordRequest(true, time.Since(start))
	return resp, nil
}


// Helper methods

func (p *ZaiProvider) mapModel(model string) string {
	// Map Responses API models to z.ai models
	modelMap := map[string]string{
		"gpt-4.1":       "glm-5",
		"gpt-4.1-mini":  "glm-5-flash",
		"gpt-4":         "glm-4",
		"gpt-4-turbo":   "glm-4-turbo",
		"gpt-3.5-turbo": "glm-3-turbo",
	}

	if mapped, ok := modelMap[model]; ok {
		return mapped
	}

	// Return as-is if no mapping
	return model
}

func (p *ZaiProvider) transformInputItem(item map[string]interface{}) map[string]interface{} {
	itemType, _ := item["type"].(string)

	switch itemType {
	case "message":
		role, _ := item["role"].(string)
		content := p.transformContent(item["content"])
		return map[string]interface{}{
			"role":    role,
			"content": content,
		}

	case "input_text":
		text, _ := item["text"].(string)
		return map[string]interface{}{
			"role":    "user",
			"content": text,
		}

	case "function_call_output":
		// Handle function call results
		callID, _ := item["call_id"].(string)
		output, _ := item["output"].(string)
		return map[string]interface{}{
			"role":         "tool",
			"tool_call_id": callID,
			"content":      output,
		}

	default:
		return nil
	}
}

func (p *ZaiProvider) transformContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func (p *ZaiProvider) transformTools(tools []Tool) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, tool := range tools {
		if tool.Type == "function" && tool.Function != nil {
			result = append(result, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Function.Name,
					"description": tool.Function.Description,
					"parameters":  tool.Function.Parameters,
				},
			})
		}
	}
	return result
}

func (p *ZaiProvider) mapFinishReason(reason string) string {
	switch reason {
	case "stop":
		return "completed"
	case "length":
		return "incomplete"
	case "tool_calls":
		return "completed"
	default:
		return "failed"
	}
}

// ExecuteStream is implemented in zai_streaming.go
