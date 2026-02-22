package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ExecuteStream executes streaming request to OpenAI with SSE
func (p *OpenAIProvider) ExecuteStream(ctx context.Context, req interface{}) (<-chan interface{}, error) {
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
	httpReq.Header.Set("Accept", "text/event-stream")

	// Execute request
	client := p.GetClient()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		p.RecordRequest(false, time.Since(start))
		return nil, fmt.Errorf("streaming request failed: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		respBody, _ := io.ReadAll(httpResp.Body)
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

	// Create channel for events
	eventChan := make(chan interface{}, 100)

	// Start goroutine to read SSE stream
	go func() {
		defer close(eventChan)
		defer httpResp.Body.Close()

		scanner := bufio.NewScanner(httpResp.Body)

		for scanner.Scan() {
			line := scanner.Text()

			// Handle SSE format
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				
				// Check for end of stream
				if data == "[DONE]" {
					eventChan <- map[string]interface{}{
						"type": "done",
						"data": nil,
					}
					break
				}

				// Parse JSON data
				var chunk map[string]interface{}
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					continue
				}

				// Send chunk to channel
				eventChan <- chunk
			}
		}

		if err := scanner.Err(); err != nil {
			eventChan <- map[string]interface{}{
				"type": "error",
				"error": err.Error(),
			}
		}

		p.RecordRequest(true, time.Since(start))
	}()

	return eventChan, nil
}
