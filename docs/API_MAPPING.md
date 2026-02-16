# API Format Mapping: Codex CLI (z.ai) to OpenAI Chat Completions

## Overview

This document describes the mapping between the z.ai API format (used by Codex CLI) and the OpenAI Chat Completions API format. This mapping is essential for the TypeScript developer implementing the API translator.

**Codex CLI (z.ai) uses:** OpenAI Responses API format
**Target format:** OpenAI Chat Completions API format

## API Endpoint Mapping

| Aspect | Codex CLI (z.ai) | OpenAI Chat Completions |
|--------|-------------------|----------------------|
| **Endpoint** | `https://api.z.ai/v1/responses` | `https://api.openai.com/v1/chat/completions` |
| **Method** | POST | POST |
| **Authentication** | Bearer token in header | Bearer token in header |

---

## Request Format Mapping

### Simple Request Examples

#### Codex CLI (z.ai) - Responses API Format
```json
{
  "model": "gpt-5",
  "input": "Tell me a joke.",
  "instructions": "You are a helpful assistant."
}
```

#### OpenAI Chat Completions Format
```json
{
  "model": "gpt-5",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Tell me a joke."
    }
  ]
}
```

---

## Field Mapping Table

### Request Fields

| Codex CLI (z.ai) Field | OpenAI Chat Completions Field | Notes |
|-------------------------|-------------------------------|-------|
| `input` | `messages` | `input` can be string or array of items; `messages` is always array |
| `instructions` | `messages[0].content` where `role: "system"` | Maps to first message with system role |
| `model` | `model` | Direct mapping |
| `temperature` | `temperature` | Direct mapping |
| `max_output_tokens` | `max_completion_tokens` | Name difference, same function |
| `top_p` | `top_p` | Direct mapping |
| `stream` | `stream` | Direct mapping |
| `store` | `store` | Direct mapping |
| `metadata` | `metadata` | Direct mapping |
| `tools` | `tools` | Different structure (see Tools section) |
| `tool_choice` | `tool_choice` | Different structure (see Tools section) |
| `parallel_tool_calls` | `parallel_tool_calls` | Direct mapping |
| `previous_response_id` | N/A | Responses API only; not supported |
| `truncation` | N/A | Responses API only |
| `text.format` | `response_format` | Different structure (see Structured Outputs) |
| `effort` | `reasoning_effort` | Different name, same values |
| `verbosity` | `verbosity` | Direct mapping |
| `include` | N/A | Responses API only |
| `conversation` | N/A | Responses API only |

---

## Message Format Mapping

### Input Messages

Codex CLI accepts input in multiple ways:

#### 1. Simple string input
```json
{
  "input": "What is the capital of France?"
}
```
**Maps to:**
```json
{
  "messages": [
    {
      "role": "user",
      "content": "What is the capital of France?"
    }
  ]
}
```

#### 2. Array of message objects
```json
{
  "input": [
    {
      "role": "user",
      "content": "What is the capital of France?"
    }
  ]
}
```
**Maps to:**
```json
{
  "messages": [
    {
      "role": "user",
      "content": "What is the capital of France?"
    }
  ]
}
```

#### 3. Instructions + input
```json
{
  "instructions": "You are a helpful assistant.",
  "input": "What is the capital of France?"
}
```
**Maps to:**
```json
{
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "What is the capital of France?"
    }
  ]
}
```

---

## Response Format Mapping

### Response Structure

#### Codex CLI (z.ai) - Responses API Response
```json
{
  "id": "resp_123",
  "object": "response",
  "created_at": 1741476777,
  "status": "completed",
  "model": "gpt-4o-2024-08-06",
  "output": [
    {
      "type": "message",
      "id": "msg_123",
      "status": "completed",
      "role": "assistant",
      "content": [
        {
          "type": "output_text",
          "text": "Hello!",
          "annotations": []
        }
      ]
    }
  ],
  "usage": {
    "input_tokens": 10,
    "output_tokens": 5,
    "total_tokens": 15
  }
}
```

#### OpenAI Chat Completions Response
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1741476777,
  "model": "gpt-4o-2024-08-06",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello!"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 5,
    "total_tokens": 15
  }
}
```

### Response Field Mapping

| Codex CLI (z.ai) Field | OpenAI Chat Completions Field | Transformation Required |
|-------------------------|-------------------------------|----------------------|
| `id` | `id` | Different ID format |
| `object` | `object` | Different value: "response" vs "chat.completion" |
| `created_at` | `created` | Same value, different name |
| `status` | N/A | Responses API only; use `finish_reason` instead |
| `output[]` | `choices[]` | Array of Items vs array of Choices |
| `output[].type` | N/A | Filter items where type="message" |
| `output[].id` | N/A | Internal ID, not in Chat Completions |
| `output[].content[]` | `choices[].message.content` | Array of content blocks vs single string |
| `output[].content[].text` | `choices[].message.content` | Extract from content array |
| `usage.input_tokens` | `usage.prompt_tokens` | Direct mapping |
| `usage.output_tokens` | `usage.completion_tokens` | Direct mapping |
| `usage.total_tokens` | `usage.total_tokens` | Direct mapping |

---

## Tools Format Mapping

### Tool Definitions

#### Codex CLI (z.ai) - Responses API Format
```json
{
  "tools": [
    {
      "type": "function",
      "name": "get_weather",
      "description": "Get weather for a location",
      "parameters": {
        "type": "object",
        "properties": {
          "location": {
            "type": "string"
          }
        },
        "required": ["location"],
        "additionalProperties": false
      }
    }
  ]
}
```

#### OpenAI Chat Completions Format
```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get weather for a location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string"
            }
          },
          "required": ["location"],
          "additionalProperties": false
        }
      }
    }
  ]
}
```

**Key Difference:** Responses API uses internal tagging (fields at same level), Chat Completions uses external tagging (nested `function` object).

### Tool Call Response Format

#### Codex CLI (z.ai) - Responses API Tool Call
```json
{
  "type": "function_call",
  "id": "fc_123",
  "name": "get_weather",
  "arguments": "{\"location\":\"Paris\"}"
}
```

#### OpenAI Chat Completions Tool Call
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1741476777,
  "model": "gpt-4o-2024-08-06",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "tool_calls": [
          {
            "id": "call_123",
            "type": "function",
            "function": {
              "name": "get_weather",
              "arguments": "{\"location\":\"Paris\"}"
            }
          }
        ]
      }
    }
  ]
}
```

### Tool Output Format

#### Codex CLI (z.ai) - Responses API
```json
{
  "type": "function_call_output",
  "call_id": "fc_123",
  "output": "{\"temp\": 20, \"condition\": \"sunny\"}"
}
```

#### OpenAI Chat Completions
```json
{
  "role": "tool",
  "content": null,
  "tool_call_id": "call_123",
  "name": "get_weather"
}
```
With the output as a separate message or in the next request's messages array.

---

## Structured Outputs Mapping

### Codex CLI (z.ai) - Responses API Format
```json
{
  "text": {
    "format": {
      "type": "json_schema",
      "name": "person",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "age": {"type": "number"}
        },
        "required": ["name", "age"],
        "additionalProperties": false
      }
    }
  }
}
```

### OpenAI Chat Completions Format
```json
{
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "person",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "age": {"type": "number"}
        },
        "required": ["name", "age"],
        "additionalProperties": false
      }
    }
  }
}
```

---

## Streaming (SSE) Mapping

### Codex CLI (z.ai) - Responses API Streaming

The Responses API uses Server-Sent Events (SSE) with the following event types:

| Event Type | Description |
|------------|-------------|
| `response.created` | Emitted when a new response is created |
| `response.output_item.added` | Emitted when a new output item is added |
| `response.output_item.done` | Emitted when an output item is completed |
| `response.completed` | Emitted when the response is complete |
| `response.failed` | Emitted if the response fails |

**Example SSE Event:**
```
event: response.output_item.added
data: {"id":"msg_123","type":"message","status":"in_progress",...}
```

### OpenAI Chat Completions Streaming

The Chat Completions API uses SSE with these event types:

| Event Type | Description |
|------------|-------------|
| `chat.completion.chunk` | Emitted for each delta of the response |
| `chat.completion.completed` | Emitted when the response is complete |

**Example SSE Event:**
```
data: {"id":"chatcmpl-123","object":"chat.completion","created":1741476777,...}
```

---

## Conversation/Context Management

### Multi-turn Conversations

#### Codex CLI (z.ai) Approach

The Responses API can manage conversation state using `previous_response_id`:

```json
{
  "model": "gpt-5",
  "input": "And its population?",
  "previous_response_id": "resp_123"
}
```

#### OpenAI Chat Completions Approach

Chat Completions is stateless - you must manually manage context:

```json
{
  "model": "gpt-5",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "What is the capital of France?"
    },
    {
      "role": "assistant",
      "content": "The capital of France is Paris."
    },
    {
      "role": "user",
      "content": "And its population?"
    }
  ]
}
```

---

## Special Considerations

### 1. Content Array vs. Single String

**Codex CLI (z.ai):** Content is always an array of content blocks
```json
{
  "content": [
    {
      "type": "output_text",
      "text": "Hello!"
    }
  ]
}
```

**OpenAI Chat Completions:** Content is a string
```json
{
  "content": "Hello!"
}
```

**Transformation:** Concatenate all text content blocks into a single string.

### 2. Reasoning Items

**Codex CLI (z.ai):** Can include reasoning items in output
```json
{
  "output": [
    {
      "type": "reasoning",
      "summary": [
        {
          "type": "summary_text",
          "text": "Let me think about this..."
        }
      ]
    },
    {
      "type": "message",
      "role": "assistant",
      "content": [...]
    }
  ]
}
```

**OpenAI Chat Completions:** No direct equivalent; reasoning is opaque

**Transformation:** Filter out reasoning items, only process message items.

### 3. Built-in Tools

**Codex CLI (z.ai) - Responses API:** Supports built-in tools
```json
{
  "tools": [
    {"type": "web_search"},
    {"type": "file_search"},
    {"type": "code_interpreter"}
  ]
}
```

**OpenAI Chat Completions:** No direct equivalent for built-in tools

**Transformation:** These tools cannot be directly translated; they require custom implementation.

### 4. Call ID Mapping

**Codex CLI (z.ai):** Uses `call_id` field in function_call and function_call_output items
**OpenAI Chat Completions:** Uses `id` field in tool_calls and `tool_call_id` in tool responses

**Transformation:** Map `call_id` → `id` for tool calls, and `call_id` → `tool_call_id` for outputs.

---

## Summary Table: Quick Reference

| Category | Codex CLI (z.ai) | OpenAI Chat Completions |
|----------|-------------------|----------------------|
| **Basic Text** | `input: "text"` | `messages: [{role: "user", content: "text"}]` |
| **System Prompt** | `instructions: "text"` | `messages: [{role: "system", content: "text"}]` |
| **Temperature** | `temperature: 0.7` | `temperature: 0.7` |
| **Max Tokens** | `max_output_tokens: 1000` | `max_completion_tokens: 1000` |
| **Streaming** | `stream: true` | `stream: true` |
| **Tool Definition** | `{type: "function", name: "..."}` | `{type: "function", function: {name: "..."}}` |
| **Tool Call in Response** | `{type: "function_call", ...}` | `message.tool_calls: [...]` |
| **Response Text** | `output_text` (SDK helper) | `choices[0].message.content` |
| **Usage** | `usage.{input_tokens, output_tokens}` | `usage.{prompt_tokens, completion_tokens}` |

---

## Implementation Notes for TypeScript Developer

### Critical Transformation Steps

1. **Request Translation:**
   - Convert `instructions` → `messages[0]` with `role: "system"`
   - Convert `input` (string) → `messages[n]` with `role: "user"`
   - Convert `input` (array) → `messages` array directly
   - Flatten tool definitions from internal to external tagging
   - Map `text.format` → `response_format`

2. **Response Translation:**
   - Extract `output_text` from content blocks in message items
   - Filter out non-message items (reasoning, tool calls without content)
   - Map tool calls from Items array to message.tool_calls
   - Convert `usage.output_tokens` → `usage.completion_tokens`

3. **Streaming Handling:**
   - Transform SSE event types between formats
   - Accumulate delta chunks appropriately for each format

4. **Error Handling:**
   - Map error responses between formats
   - Handle missing fields gracefully

---

## Additional Resources

- [OpenAI Responses API Documentation](https://platform.openai.com/docs/api-reference/responses/create)
- [OpenAI Chat Completions API Documentation](https://platform.openai.com/docs/api-reference/chat/create)
- [Migration Guide: Chat Completions to Responses](https://platform.openai.com/docs/guides/migrate-to-responses/)

---

**Document Version:** 1.0
**Last Updated:** February 2026
**Codex CLI Version:** Latest (uses z.ai API)
**OpenAI API Version:** 2025 updates
