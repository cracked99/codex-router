package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider implements Provider for OpenAI backend
type OpenAIProvider struct {
	*BaseProvider
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		BaseProvider: NewBaseProvider("openai"),
	}
}

// Initialize initializes the OpenAI provider
func (p *OpenAIProvider) Initialize(config ProviderConfig) error {
	// Set defaults
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}
	if len(config.Models) == 0 {
		config.Models = []string{"gpt-*", "o1-*", "o3-*"}
	}

	return p.BaseProvider.Initialize(config)
}

// TransformRequest transforms Responses API request to OpenAI format
func (p *OpenAIProvider) TransformRequest(req *ResponsesRequest) (interface{}, error) {
	// OpenAI's Responses API is similar, minimal transformation needed
	chatReq := map[string]interface{}{
		"model": req.Model,
	}

	// Transform input to messages (similar to z.ai)
	messages := []map[string]interface{}{}

	if req.Instructions != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": req.Instructions,
		})
	}

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

	// Copy parameters
	if req.Temperature != nil {
		chatReq["temperature"] = *req.Temperature
	}
	if req.MaxOutputTokens != nil {
		chatReq["max_tokens"] = *req.MaxOutputTokens
	}
	if req.TopP != nil {
		chatReq["top_p"] = *req.TopP
	}
	if req.Stream {
		chatReq["stream"] = true
	}

	return chatReq, nil
}

// TransformResponse transforms OpenAI response to Responses API
func (p *OpenAIProvider) TransformResponse(resp interface{}) (*ResponsesResponse, error) {
	chatResp, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	responsesResp := &ResponsesResponse{
		ID:        fmt.Sprintf("resp_%d", time.Now().UnixNano()),
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Status:    "completed",
		Output:    []OutputItem{},
	}

	// Extract from OpenAI format (similar structure)
	if choices, ok := chatResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			output := OutputItem{
				Type:    "message",
				Role:    "assistant",
				Content: []Content{},
			}

			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					output.Content = append(output.Content, Content{
						Type: "output_text",
						Text: content,
					})
				}
			}

			responsesResp.Output = append(responsesResp.Output, output)
		}
	}

	return responsesResp, nil
}

// TransformStreamEvent transforms streaming event
func (p *OpenAIProvider) TransformStreamEvent(event interface{}) (*ResponsesStreamEvent, error) {
	chunk, ok := event.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if content, ok := delta["content"].(string); ok && content != "" {
					return &ResponsesStreamEvent{
						Type: "response.output_text.delta",
						Data: map[string]interface{}{"delta": content},
						Timestamp: time.Now().Unix(),
					}, nil
				}
			}
		}
	}

	return nil, nil
}

// Execute executes a request to OpenAI
func (p *OpenAIProvider) Execute(ctx context.Context, req interface{}) (interface{}, error) {
	start := time.Now()

	body, err := json.Marshal(req)
	if err != nil {
		p.RecordRequest(false, 0)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

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

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := p.GetClient()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

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

	var resp map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	p.RecordRequest(true, time.Since(start))
	return resp, nil
}


func (p *OpenAIProvider) transformInputItem(item map[string]interface{}) map[string]interface{} {
	itemType, _ := item["type"].(string)

	switch itemType {
	case "message":
		role, _ := item["role"].(string)
		content, _ := item["content"].(string)
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
	default:
		return nil
	}
}

// ExecuteStream is implemented in openai_streaming.go
