package translator

import (
	"github.com/plasmadev/codex-api-router/pkg/api"
)

// Translator transforms between Responses API and Chat Completions API formats
// This is a stub interface - will be implemented by the TypeScript agent
type Translator interface {
	// TransformRequest transforms a Responses API request to Chat Completions format
	TransformRequest(req *api.ResponseRequest) (*api.ChatCompletionRequest, error)

	// TransformResponse transforms a Chat Completions response to Responses API format
	TransformResponse(resp *api.ChatCompletionResponse) (*api.Response, error)

	// TransformStreamChunk transforms a streaming SSE chunk
	TransformStreamChunk(event, data string) (string, string, error)
}

// StubTranslator is a stub implementation that does basic transformations
type StubTranslator struct{}

// NewStubTranslator creates a new stub translator
func NewStubTranslator() *StubTranslator {
	return &StubTranslator{}
}

// TransformRequest does basic request transformation
func (t *StubTranslator) TransformRequest(req *api.ResponseRequest) (*api.ChatCompletionRequest, error) {
	chatReq := &api.ChatCompletionRequest{
		Model:     t.mapModel(req.Model),
		Messages:  t.transformInput(req.Input),
		Stream:    req.Stream,
	}

	if req.Temperature != nil {
		chatReq.Temperature = req.Temperature
	}
	if req.TopP != nil {
		chatReq.TopP = req.TopP
	}
	if req.MaxOutputTokens != nil {
		chatReq.MaxTokens = req.MaxOutputTokens
	}

	if len(req.Tools) > 0 {
		chatReq.Tools = t.transformTools(req.Tools)
	}

	return chatReq, nil
}

// TransformResponse does basic response transformation
func (t *StubTranslator) TransformResponse(resp *api.ChatCompletionResponse) (*api.Response, error) {
	output := []api.OutputItem{}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		content := []api.ContentBlock{}

		if choice.Message.Content != "" {
			content = append(content, api.ContentBlock{
				Type: "output_text",
				Text: choice.Message.Content,
			})
		}

		if len(content) > 0 {
			output = append(output, api.OutputItem{
				Type:    "message",
				Role:    "assistant",
				Content: content,
			})
		}
	}

	return &api.Response{
		ID:        "resp_" + resp.ID,
		Object:    "response",
		CreatedAt: resp.Created,
		Status:    "completed",
		Result: &api.ResponseResult{
			ID:     "resp_" + resp.ID,
			Status: "completed",
			Output: output,
			Usage: api.Usage{
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				TotalTokens:  resp.Usage.TotalTokens,
			},
		},
	}, nil
}

// TransformStreamChunk transforms streaming SSE chunks
func (t *StubTranslator) TransformStreamChunk(event, data string) (string, string, error) {
	// For now, pass through
	return event, data, nil
}

// transformInput converts input to messages
func (t *StubTranslator) transformInput(input interface{}) []api.ChatMessage {
	if str, ok := input.(string); ok {
		return []api.ChatMessage{
			{Role: "user", Content: str},
		}
	}
	// Default fallback
	return []api.ChatMessage{
		{Role: "user", Content: ""},
	}
}

// transformTools converts tools
func (t *StubTranslator) transformTools(tools []api.Tool) []api.ChatTool {
	chatTools := make([]api.ChatTool, len(tools))
	for i, tool := range tools {
		chatTools[i] = api.ChatTool{
			Type: tool.Type,
			Function: &api.ChatFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}
	return chatTools
}

// mapModel maps model names
func (t *StubTranslator) mapModel(model string) string {
	modelMap := map[string]string{
		"gpt-4.1":       "glm-5",
		"gpt-4":         "glm-5",
		"gpt-3.5-turbo": "glm-4",
	}
	if mapped, ok := modelMap[model]; ok {
		return mapped
	}
	return model
}
