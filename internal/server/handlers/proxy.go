package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/plasmadev/codex-api-router/internal/config"
)

// ProxyHandler handles proxying requests to the backend
type ProxyHandler struct {
	cfg    *config.Config
	logger *slog.Logger
	client *http.Client
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(cfg *config.Config, logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{
			Timeout: cfg.Zai.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ServeHTTP handles the proxy request
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = time.Now() // Record start time for metrics (not used yet)

	// Log incoming request
	h.logger.Info("incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	// Handle GET requests for retrieving responses
	if r.Method == http.MethodGet {
		h.handleGetResponse(w, r)
		return
	}

	// Handle DELETE requests
	if r.Method == http.MethodDelete {
		h.handleDeleteResponse(w, r)
		return
	}

	// Handle POST requests for creating responses
	if r.Method == http.MethodPost {
		h.handleCreateResponse(w, r)
		return
	}

	// Method not allowed
	h.logger.Warn("method not allowed", "method", r.Method)
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    "invalid_request_error",
			"message": fmt.Sprintf("Method %s not allowed", r.Method),
		},
	})
}

func (h *ProxyHandler) handleCreateResponse(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the Responses API request
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "invalid_request_error",
				"message": "Invalid JSON in request body",
			},
		})
		return
	}

	// Log request details
	h.logger.Debug("request parsed",
		"model", req["model"],
		"stream", req["stream"],
		"has_instructions", req["instructions"] != nil,
	)

	// Transform Responses API request to Chat Completions format
	chatReq := h.transformRequest(req)

	// Marshal chat completions request
	chatBody, err := json.Marshal(chatReq)
	if err != nil {
		h.logger.Error("failed to marshal chat request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Debug log the transformed request
	h.logger.Info("sending to backend", "body", string(chatBody))

	// Create backend request
	backendURL := h.cfg.Zai.BaseURL + "/chat/completions"
	backendReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, backendURL, bytes.NewReader(chatBody))
	if err != nil {
		h.logger.Error("failed to create backend request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set headers
	backendReq.Header.Set("Content-Type", "application/json")
	backendReq.Header.Set("Authorization", "Bearer "+h.cfg.Zai.APIKey)

	// Check if streaming is requested
	streaming := false
	if s, ok := req["stream"].(bool); ok {
		streaming = s
	}

	if streaming {
		h.handleStreamingResponse(w, r, backendReq)
	} else {
		h.handleNonStreamingResponse(w, r, backendReq)
	}
}

func (h *ProxyHandler) handleNonStreamingResponse(w http.ResponseWriter, r *http.Request, backendReq *http.Request) {
	// Execute backend request
	resp, err := h.client.Do(backendReq)
	if err != nil {
		h.logger.Error("backend request failed", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "api_error",
				"message": "Failed to reach backend server",
			},
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("failed to read backend response", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// Check for non-OK status
	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("backend returned non-OK status",
			"status", resp.StatusCode,
			"body", string(body),
		)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	// Parse Chat Completions response
	var chatResp map[string]interface{}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		h.logger.Error("failed to parse backend response", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// Transform to Responses API format
	responsesResp := h.transformResponse(chatResp)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responsesResp)
}

func (h *ProxyHandler) handleStreamingResponse(w http.ResponseWriter, r *http.Request, backendReq *http.Request) {
	// Execute backend request
	resp, err := h.client.Do(backendReq)
	if err != nil {
		h.logger.Error("backend request failed", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "api_error",
				"message": "Failed to reach backend server",
			},
		})
		return
	}
	defer resp.Body.Close()

	// Check for non-OK status
	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("backend returned non-OK status for stream",
			"status", resp.StatusCode,
		)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	// Set up SSE headers
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.Error("streaming not supported")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Transform and stream events
	h.transformStream(resp.Body, w, flusher)
}

func (h *ProxyHandler) handleGetResponse(w http.ResponseWriter, r *http.Request) {
	// Extract response ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "invalid_request_error",
				"message": "Invalid response ID",
			},
		})
		return
	}
	responseID := parts[3]

	// For now, return not implemented since we don't have response storage
	h.logger.Debug("get response not implemented", "response_id", responseID)
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    "invalid_request_error",
			"message": "Response retrieval not implemented in proxy mode",
		},
	})
}

func (h *ProxyHandler) handleDeleteResponse(w http.ResponseWriter, r *http.Request) {
	// Extract response ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "invalid_request_error",
				"message": "Invalid response ID",
			},
		})
		return
	}
	responseID := parts[3]

	// For now, return not implemented
	h.logger.Debug("delete response not implemented", "response_id", responseID)
	w.WriteHeader(http.StatusNotImplemented)
}

// transformRequest transforms Responses API request to Chat Completions format
func (h *ProxyHandler) transformRequest(req map[string]interface{}) map[string]interface{} {
	chatReq := make(map[string]interface{})

	// Copy model
	if model, ok := req["model"]; ok {
		chatReq["model"] = model
	}

	// Transform input to messages
	messages := []map[string]interface{}{}

	// Add instructions if present
	if instructions, ok := req["instructions"].(string); ok && instructions != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": instructions,
		})
	}

	// Add input/messages
	if input, ok := req["input"]; ok {
		switch v := input.(type) {
		case string:
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": v,
			})
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					msg := h.transformInputItem(itemMap)
					if msg != nil {
						messages = append(messages, msg)
					}
				}
			}
		}
	}

	chatReq["messages"] = messages

	// Only copy parameters that z.ai supports
	if temp, ok := req["temperature"]; ok && temp != nil {
		chatReq["temperature"] = temp
	}
	if maxTokens, ok := req["max_output_tokens"]; ok && maxTokens != nil {
		chatReq["max_tokens"] = maxTokens
	}
	if topP, ok := req["top_p"]; ok && topP != nil {
		chatReq["top_p"] = topP
	}
	if stream, ok := req["stream"]; ok {
		chatReq["stream"] = stream
	}

	// Transform tools (only if present and non-empty)
	if tools, ok := req["tools"].([]interface{}); ok && len(tools) > 0 {
		chatReq["tools"] = h.transformTools(tools)
	}

	// Note: Don't copy tool_choice as z.ai may not support all formats
	// Only copy if it's a simple string value
	if tc, ok := req["tool_choice"].(string); ok && tc != "" {
		chatReq["tool_choice"] = tc
	}

	return chatReq
}

func (h *ProxyHandler) transformInputItem(item map[string]interface{}) map[string]interface{} {
	role, _ := item["role"].(string)

	switch role {
	case "user", "assistant", "system":
		content := ""
		if contentArr, ok := item["content"].([]interface{}); ok && len(contentArr) > 0 {
			if contentItem, ok := contentArr[0].(map[string]interface{}); ok {
				if text, ok := contentItem["text"].(string); ok {
					content = text
				}
			}
		}
		return map[string]interface{}{
			"role":    role,
			"content": content,
		}
	}

	return nil
}

func (h *ProxyHandler) transformTools(tools []interface{}) []map[string]interface{} {
	transformed := make([]map[string]interface{}, 0, len(tools))

	for _, tool := range tools {
		if toolMap, ok := tool.(map[string]interface{}); ok {
			// Skip tools with null or empty names
			name, _ := toolMap["name"].(string)
			if name == "" {
				continue
			}

			functionDef := map[string]interface{}{
				"name": name,
			}
			if desc, ok := toolMap["description"].(string); ok && desc != "" {
				functionDef["description"] = desc
			}
			if params, ok := toolMap["parameters"]; ok && params != nil {
				functionDef["parameters"] = params
			}

			transformedTool := map[string]interface{}{
				"type":     "function",
				"function": functionDef,
			}
			transformed = append(transformed, transformedTool)
		}
	}

	return transformed
}

// transformResponse transforms Chat Completions response to Responses API format
func (h *ProxyHandler) transformResponse(resp map[string]interface{}) map[string]interface{} {
	responsesResp := map[string]interface{}{
		"id":         "resp_" + generateID(),
		"object":     "response",
		"created_at": time.Now().Unix(),
		"status":     "completed",
	}

	// Transform choices to output
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			output := []map[string]interface{}{}

			if message, ok := choice["message"].(map[string]interface{}); ok {
				msg := map[string]interface{}{
					"type":   "message",
					"id":     "msg_" + generateID(),
					"status": "completed",
					"role":   "assistant",
					"content": []map[string]interface{}{},
				}

				if content, ok := message["content"].(string); ok {
					msg["content"] = []map[string]interface{}{
						{
							"type": "output_text",
							"text": content,
						},
					}
				}

				if _, ok := message["tool_calls"].([]interface{}); ok {
					// Handle tool calls if present (TODO: implement tool call transformation)
				}

				output = append(output, msg)
			}

			responsesResp["output"] = output

			// Map finish reason
			if finishReason, ok := choice["finish_reason"].(string); ok {
				responsesResp["status"] = mapFinishReason(finishReason)
			}
		}
	}

	// Copy usage
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		responsesResp["usage"] = map[string]interface{}{
			"input_tokens":  usage["prompt_tokens"],
			"output_tokens": usage["completion_tokens"],
			"total_tokens":  usage["total_tokens"],
		}
	}

	// Copy model
	if model, ok := resp["model"].(string); ok {
		responsesResp["model"] = model
	}

	return responsesResp
}

func (h *ProxyHandler) transformStream(body io.ReadCloser, w io.Writer, flusher http.Flusher) {
	reader := bufio.NewReader(body)
	responseID := fmt.Sprintf("resp_%d", time.Now().UnixNano())
	itemID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	sentCreated := false
	sentOutputItemAdded := false
	sentContentPartAdded := false
	sequenceNumber := 0
	fullText := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				h.logger.Error("error reading stream", "error", err)
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			if data == "[DONE]" {
				// Send output_text.done first if we have content
				if sentContentPartAdded && fullText != "" {
					outputTextDone := map[string]interface{}{
						"type":            "response.output_text.done",
						"item_id":         itemID,
						"output_index":    0,
						"content_index":   0,
						"sequence_number": sequenceNumber,
						"text":            fullText,
					}
					eventData, _ := json.Marshal(outputTextDone)
					fmt.Fprintf(w, "event: response.output_text.done\n")
					fmt.Fprintf(w, "data: %s\n\n", string(eventData))
					flusher.Flush()
					sequenceNumber++
				}

				// Send content_part.done if we added content
				if sentContentPartAdded {
					contentPartDone := map[string]interface{}{
						"type":            "response.content_part.done",
						"item_id":         itemID,
						"output_index":    0,
						"content_index":   0,
						"sequence_number": sequenceNumber,
						"part": map[string]interface{}{
							"type":        "output_text",
							"text":        fullText,
							"annotations": []interface{}{},
						},
					}
					eventData, _ := json.Marshal(contentPartDone)
					fmt.Fprintf(w, "event: response.content_part.done\n")
					fmt.Fprintf(w, "data: %s\n\n", string(eventData))
					flusher.Flush()
					sequenceNumber++
				}

				// Send output_item.done
				if sentOutputItemAdded {
					outputItemDone := map[string]interface{}{
						"type":            "response.output_item.done",
						"output_index":    0,
						"sequence_number": sequenceNumber,
						"item": map[string]interface{}{
							"id":      itemID,
							"type":    "message",
							"role":    "assistant",
							"status":  "completed",
							"content": []interface{}{
								map[string]interface{}{
									"type":        "output_text",
									"text":        fullText,
									"annotations": []interface{}{},
								},
							},
						},
					}
					eventData, _ := json.Marshal(outputItemDone)
					fmt.Fprintf(w, "event: response.output_item.done\n")
					fmt.Fprintf(w, "data: %s\n\n", string(eventData))
					flusher.Flush()
					sequenceNumber++
				}

				// Send response.completed
				completedEvent := map[string]interface{}{
					"type":            "response.completed",
					"sequence_number": sequenceNumber,
					"response": map[string]interface{}{
						"id":         responseID,
						"object":     "response",
						"status":     "completed",
						"output": []map[string]interface{}{
							{
								"id":      itemID,
								"type":    "message",
								"role":    "assistant",
								"status":  "completed",
								"content": []interface{}{},
							},
						},
					},
				}
				eventData, _ := json.Marshal(completedEvent)
				fmt.Fprintf(w, "event: response.completed\n")
				fmt.Fprintf(w, "data: %s\n\n", string(eventData))
				flusher.Flush()
				sequenceNumber++

				// Send response.done event
				doneEvent := map[string]interface{}{
					"type":            "response.done",
					"sequence_number": sequenceNumber,
				}
				eventData, _ = json.Marshal(doneEvent)
				fmt.Fprintf(w, "event: response.done\n")
				fmt.Fprintf(w, "data: %s\n\n", string(eventData))
				flusher.Flush()
				break
			}

			// Parse the Chat Completions chunk
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				h.logger.Debug("failed to parse chunk", "error", err)
				continue
			}

			// Send response.created event first
			if !sentCreated {
				created := int64(0)
				if c, ok := chunk["created"].(float64); ok {
					created = int64(c)
				}
				model, _ := chunk["model"].(string)

				// Send response.created
				createdEvent := map[string]interface{}{
					"type":            "response.created",
					"sequence_number": sequenceNumber,
					"response": map[string]interface{}{
						"id":         responseID,
						"object":     "response",
						"created_at": created,
						"model":      model,
						"status":     "in_progress",
						"output":     []interface{}{},
					},
				}
				eventData, _ := json.Marshal(createdEvent)
				fmt.Fprintf(w, "event: response.created\n")
				fmt.Fprintf(w, "data: %s\n\n", string(eventData))
				flusher.Flush()
				sequenceNumber++

				// Send response.in_progress
				inProgressEvent := map[string]interface{}{
					"type":            "response.in_progress",
					"sequence_number": sequenceNumber,
					"response": map[string]interface{}{
						"id":         responseID,
						"object":     "response",
						"created_at": created,
						"model":      model,
						"status":     "in_progress",
						"output":     []interface{}{},
					},
				}
				eventData, _ = json.Marshal(inProgressEvent)
				fmt.Fprintf(w, "event: response.in_progress\n")
				fmt.Fprintf(w, "data: %s\n\n", string(eventData))
				flusher.Flush()
				sentCreated = true
				sequenceNumber++
			}

			// Transform choices to output_text deltas
			if choices, ok := chunk["choices"].([]interface{}); ok {
				for _, choice := range choices {
					if choiceMap, ok := choice.(map[string]interface{}); ok {
						if delta, ok := choiceMap["delta"].(map[string]interface{}); ok {
							// Handle content
							if content, ok := delta["content"].(string); ok && content != "" {
								// Send output_item.added first if not sent
								if !sentOutputItemAdded {
									outputItemAdded := map[string]interface{}{
										"type":            "response.output_item.added",
										"output_index":    0,
										"sequence_number": sequenceNumber,
										"item": map[string]interface{}{
											"id":      itemID,
											"type":    "message",
											"role":    "assistant",
											"status":  "in_progress",
											"content": []interface{}{},
										},
									}
									eventData, _ := json.Marshal(outputItemAdded)
									fmt.Fprintf(w, "event: response.output_item.added\n")
									fmt.Fprintf(w, "data: %s\n\n", string(eventData))
									flusher.Flush()
									sentOutputItemAdded = true
									sequenceNumber++
								}

								// Send content_part.added if not sent
								if !sentContentPartAdded {
									contentPartAdded := map[string]interface{}{
										"type":            "response.content_part.added",
										"item_id":         itemID,
										"output_index":    0,
										"content_index":   0,
										"sequence_number": sequenceNumber,
										"part": map[string]interface{}{
											"type":        "output_text",
											"text":        "",
											"annotations": []interface{}{},
										},
									}
									eventData, _ := json.Marshal(contentPartAdded)
									fmt.Fprintf(w, "event: response.content_part.added\n")
									fmt.Fprintf(w, "data: %s\n\n", string(eventData))
									flusher.Flush()
									sentContentPartAdded = true
									sequenceNumber++
								}

								// Append to full text
								fullText += content

								// Send delta event with correct format
								deltaEvent := map[string]interface{}{
									"type":            "response.output_text.delta",
									"item_id":         itemID,
									"output_index":    0,
									"content_index":   0,
									"sequence_number": sequenceNumber,
									"delta":           content,
								}
								eventData, _ := json.Marshal(deltaEvent)
								fmt.Fprintf(w, "event: response.output_text.delta\n")
								fmt.Fprintf(w, "data: %s\n\n", string(eventData))
								flusher.Flush()
								sequenceNumber++
							}
						}
					}
				}
			}
		}
	}
}

func mapFinishReason(reason string) string {
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

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
