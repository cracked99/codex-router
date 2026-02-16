# Codex API Router - Architecture Design

## Overview

The Codex API Router is a proxy service that translates between Codex CLI's Responses API and z.ai's Chat Completions API. This enables Codex CLI v0.99+ (which only supports Responses API) to work with z.ai (which only supports Chat Completions API).

## Problem Statement

- **Codex CLI v0.99+**: Only supports OpenAI's Responses API (`/v1/responses`)
- **z.ai**: Only supports OpenAI-compatible Chat Completions API (`/v4/chat/completions`)
- **Solution**: A transparent proxy router that translates request/response formats bidirectionally

## High-Level Architecture

```
┌─────────────────┐      ┌──────────────────────────────┐      ┌─────────────┐
│   Codex CLI     │─────▶│   Codex API Router           │─────▶│    z.ai     │
│  (Responses API)│      │  - Go CLI & HTTP Server      │      │ (Chat API)  │
└─────────────────┘      │  - TypeScript Translator     │      └─────────────┘
                         │  - SSE Stream Handler        │              │
                         └──────────────────────────────┘              │
                                    ▲                                │
                                    └────────────────────────────────┘
```

### Component Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         codex-api-router                            │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐   ┌──────────────┐   ┌─────────────────────────┐ │
│  │  Go CLI     │──▶│ HTTP Server  │──▶│   TypeScript (WASM)     │ │
│  │  - cmd/     │   │ - handlers/  │   │   - translator/         │ │
│  │  - server/  │   │ - middleware │   │   - types/              │ │
│  └─────────────┘   └──────────────┘   └─────────────────────────┘ │
│         │                   │                      │                │
│         ▼                   ▼                      ▼                │
│  ┌─────────────┐   ┌──────────────┐   ┌─────────────────────────┐ │
│  │ Config     │   │ SSE Stream   │   │  API Mapping Layer       │ │
│  │ Management │   │ Processor    │   │  - Request Transform    │ │
│  └─────────────┘   └──────────────┘   │  - Response Transform   │ │
│                                         │  - Tool Calling         │ │
│                                         └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

## Request/Response Flow

### Non-Streaming Request Flow

```
1. Codex CLI → Router (Responses API format)
   POST /v1/responses
   {
     "model": "gpt-4.1",
     "input": "Hello, how are you?"
   }

2. Router → Translator (TypeScript)
   - Parse Responses API request
   - Map parameters to Chat Completions format

3. Translator → Router (Chat Completions format)
   {
     "model": "glm-5",
     "messages": [
       {"role": "user", "content": "Hello, how are you?"}
     ]
   }

4. Router → z.ai
   POST https://api.z.ai/api/paas/v4/chat/completions

5. z.ai → Router (Chat Completions response)
   {
     "choices": [{
       "message": {
         "role": "assistant",
         "content": "I'm doing well!"
       }
     }]
   }

6. Translator → Router (Responses API format)
   {
     "id": "resp_123",
     "object": "response",
     "output": [{
       "type": "message",
       "role": "assistant",
       "content": [{
         "type": "output_text",
         "text": "I'm doing well!"
       }]
     }]
   }

7. Router → Codex CLI
```

### Streaming Request Flow (SSE)

```
1. Codex CLI → Router (stream: true)
   POST /v1/responses
   { "stream": true, "input": "..." }

2. Router establishes connection to z.ai with stream=true

3. z.ai streams SSE events:
   data: {"choices":[{"delta":{"content":"Hello"}}]}
   data: {"choices":[{"delta":{"content":" world"}}]}
   data: [DONE]

4. Router transforms each event:
   event: response.delta
   data: {"type":"output_text","text":"Hello"}

   event: response.delta
   data: {"type":"output_text","text":" world"}

   event: response.done
   data: {"id":"resp_123",...}

5. Codex CLI receives transformed SSE stream
```

## API Mapping: Responses → Chat Completions

### Request Parameter Mapping

| Responses API | Chat Completions API | Notes |
|---------------|---------------------|-------|
| `model` | `model` | Direct mapping |
| `input` (string) | `messages` | Convert to single user message |
| `input` (array) | `messages` | Transform items to message array |
| `instructions` | `messages[0]` | Prepend as system message |
| `temperature` | `temperature` | Direct mapping |
| `max_output_tokens` | `max_tokens` | Direct mapping |
| `top_p` | `top_p` | Direct mapping |
| `tools` | `tools` | Format transformation needed |
| `tool_choice` | `tool_choice` | Direct mapping |
| `stream` | `stream` | Direct mapping |
| `previous_response_id` | N/A | Handle via conversation state |
| `conversation` | N/A | Track session state internally |
| `include` | N/A | Filter response fields |

### Input Item Types to Message Mapping

| Responses Input Item | Chat Message |
|---------------------|--------------|
| `{"type": "message", "role": "user", "content": [...]}` | `{"role": "user", "content": "..."}` |
| `{"type": "message", "role": "assistant", "content": [...]}` | `{"role": "assistant", "content": "..."}` |
| `{"type": "message", "role": "system", "content": [...]}` | `{"role": "system", "content": "..."}` |
| `{"type": "input_text", "text": "..."}` | `{"role": "user", "content": "..."}` |
| `{"type": "input_image", "image_url": "..."}` | `{"role": "user", "content": [{"type":"image_url", ...}]}` |

### Response Mapping

| Chat Completions Response | Responses API Response |
|---------------------------|------------------------|
| `id` | `id` (with `resp_` prefix) |
| `created` | `created_at` |
| `model` | `model` |
| `choices[0].message` | `output[0]` (message item) |
| `choices[0].message.content` | `output[0].content[0].text` |
| `choices[0].message.tool_calls` | `output[0].content[1]` (tool_call item) |
| `choices[0].finish_reason` | `status` (mapped) |
| `usage.prompt_tokens` | `usage.input_tokens` |
| `usage.completion_tokens` | `usage.output_tokens` |
| `usage.total_tokens` | `usage.total_tokens` |

### Tool Calling Mapping

Responses API tools format:
```json
{
  "type": "function",
  "name": "get_weather",
  "description": "Get weather",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {"type": "string"}
    }
  }
}
```

Chat Completions tools format:
```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Get weather",
    "parameters": {
      "type": "object",
      "properties": {
        "location": {"type": "string"}
      }
    }
  }
}
```

## Streaming (SSE) Handling Strategy

### SSE Event Transformation

| z.ai SSE Event | Router Output Event | Format |
|----------------|---------------------|--------|
| `data: {"choices":[{"delta":{"content":"..."}}]}` | `event: response.delta\ndata: {"type":"output_text","text":"..."}` | Delta |
| `data: {"choices":[{"delta":{"tool_calls":[]}}]}` | `event: response.delta\ndata: {"type":"tool_call","...":...}` | Tool Call |
| `data: [DONE]` | `event: response.done\ndata: {...complete response...}` | Done |

### Stream Buffering

For streaming responses, the router:
1. Buffers incoming SSE chunks from z.ai
2. Transforms each chunk to Responses API format
3. Maintains response state (accumulated text, tool calls)
4. Sends transformed events to client
5. Sends final complete response on stream end

### Error Handling

- Connection drops: Reconnect and resume from last token
- Malformed SSE: Log and send error event
- z.ai errors: Translate to Responses API error format

## Go CLI Structure

```
cmd/
├── root.go              # Root command, Cobra setup
├── serve.go             # Server start command
├── config.go            # Config management commands
└── version.go           # Version info

internal/
├── server/
│   ├── server.go        # HTTP server setup
│   ├── router.go        # Route handlers
│   └── middleware.go    # Auth, logging, CORS
├── config/
│   ├── config.go        # Config struct & loading
│   └── defaults.go      # Default values
├── proxy/
│   ├── proxy.go         # Reverse proxy handler
│   └── stream.go        # SSE stream handler
├── translator/
│   ├── wasm.go          # WASM bridge to TS
│   └── bridge.go        # Go-TS communication
└── session/
    ├── session.go       # Session/conversation state
    └── cache.go         # Response ID caching

pkg/
├── api/
│   ├── responses.go     # Responses API types
│   └── chat.go          # Chat Completions types
└── version/
    └── version.go       # Version info
```

### Key Commands

```bash
# Start proxy server
codex-api-router serve

# With custom config
codex-api-router serve --config ./config.yaml

# Validate configuration
codex-api-router config validate

# Show version
codex-api-router version
```

## TypeScript Translator Modules

```
translator/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts              # Entry point for WASM
│   ├── types/
│   │   ├── responses.ts      # Responses API types
│   │   ├── chat.ts           # Chat Completions types
│   │   └── common.ts         # Shared types
│   ├── transformers/
│   │   ├── request.ts        # Request transformation
│   │   ├── response.ts       # Response transformation
│   │   ├── stream.ts         # SSE stream transformation
│   │   └── tools.ts          # Tool calling transformation
│   ├── validators/
│   │   ├── request.ts        # Request validation
│   │   └── response.ts       # Response validation
│   └── utils/
│       ├── mapper.ts         # Field mapping utilities
│       ├── buffer.ts         # Stream buffering
│       └── errors.ts         # Error formatting
└── tests/
    ├── transformers/
    └── validators/
```

### TypeScript Module Structure

#### Request Transformer
```typescript
export function transformRequest(
  request: ResponsesRequest
): ChatCompletionRequest {
  // 1. Map model parameter
  // 2. Transform input to messages
  // 3. Map tools array
  // 4. Map generation parameters
  // 5. Handle conversation state
}

export function transformInputItems(
  items: InputItem[]
  ): ChatMessage[] {
  // Transform various input item types
}
```

#### Response Transformer
```typescript
export function transformResponse(
  response: ChatCompletionResponse
): ResponsesResponse {
  // 1. Map response structure
  // 2. Transform choices to output items
  // 3. Map usage statistics
  // 4. Map finish reason to status
}
```

#### Stream Transformer
```typescript
export class StreamTransformer {
  private buffer: ResponseBuffer;

  transformChunk(chunk: ChatSSEChunk): ResponsesSSEEvent | null {
    // Transform streaming delta events
  }

  finalize(): ResponsesCompleteEvent {
    // Generate final complete response
  }
}
```

## Integration Approach

### Option A: WASM Embedding (Recommended)

**Pros:**
- Single binary distribution
- No external dependencies
- Cross-platform consistency
- TypeScript type safety

**Cons:**
- WASM compilation complexity
- Slight performance overhead
- Limited Node.js APIs in WASM

**Implementation:**
1. Compile TypeScript to WASM using `AssemblyScript` or `ts2wasm`
2. Embed WASM binary in Go binary
3. Use `wasmer-go` or `wasmtime-go` runtime
4. Export functions for request/response transformation

```go
// Go-WASM bridge
type Translator struct {
    instance *wasmer.Instance
}

func (t *Translator) TransformRequest(req ResponsesRequest) (ChatRequest, error) {
    // Call WASM exported function
}
```

### Option B: Sidecar Process

**Pros:**
- Full Node.js ecosystem access
- Easier development/debugging
- Hot reload for development

**Cons:**
- Multi-process deployment
- Inter-process communication overhead
- More complex distribution

**Implementation:**
1. Run Node.js process alongside Go binary
2. Communicate via stdio, Unix socket, or gRPC
3. Go spawns and manages Node.js lifecycle

```go
// Sidecar communication
type SidecarClient struct {
    cmd *exec.Cmd
    rpc *grpc.ClientConn
}
```

### Recommendation: WASM for Production, Sidecar for Dev

- **Development:** Use sidecar mode for faster iteration
- **Production:** Use WASM for single-binary distribution

## Configuration Schema

```yaml
# config.yaml
server:
  host: "localhost"
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

zai:
  base_url: "https://api.z.ai/api/paas/v4"
  api_key: "${ZAI_API_KEY}"  # Env var expansion
  timeout: 120s
  max_retries: 3
  retry_delay: 1s

codex:
  base_url: ""  # If running behind another proxy
  api_key_header: "Authorization"

translator:
  mode: "wasm"  # wasm | sidecar
  wasm_path: "./translator.wasm"
  sidecar_command: "node ./translator/index.js"

session:
  enabled: true
  ttl: 3600s
  max_conversations: 1000

logging:
  level: "info"  # debug | info | warn | error
  format: "json"  # json | text
  file: ""  # Optional file output

metrics:
  enabled: true
  path: "/metrics"
  format: "prometheus"
```

## Error Handling

### Error Mapping

| z.ai Error | Responses API Error |
|------------|---------------------|
| 401 Unauthorized | 401, `{"error": {"type": "authentication_error"}}` |
| 429 Rate Limit | 429, `{"error": {"type": "rate_limit_error"}}` |
| 500 Internal Error | 500, `{"error": {"type": "api_error"}}` |
| Invalid Request | 400, `{"error": {"type": "invalid_request_error"}}` |

### Error Response Format (Responses API)

```json
{
  "error": {
    "type": "invalid_request_error",
    "message": "Invalid model parameter",
    "param": "model",
    "code": "invalid_model"
  }
}
```

## Security Considerations

1. **API Key Management**: Support env vars, secret files
2. **Request Validation**: Validate all incoming requests
3. **Rate Limiting**: Implement client-side rate limiting
4. **TLS**: Support TLS for server connections
5. **Input Sanitization**: Sanitize all user inputs before forwarding
6. **Audit Logging**: Log all requests (configurable sensitivity)

## Performance Optimization

1. **Connection Pooling**: Reuse HTTP connections to z.ai
2. **Response Caching**: Cache identical requests (optional)
3. **Streaming**: Zero-copy streaming where possible
4. **Compression**: Support gzip compression
5. **Concurrent Requests**: Handle multiple simultaneous requests

## Deployment Options

### Single Binary
```bash
codex-api-router serve --config config.yaml
```

### Docker Container
```dockerfile
FROM alpine:latest
COPY codex-api-router /usr/local/bin/
EXPOSE 8080
CMD ["codex-api-router", "serve"]
```

### Kubernetes
```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: codex-api-router
spec:
  template:
    spec:
      containers:
      - name: router
        image: codex-api-router:latest
        ports:
        - containerPort: 8080
        env:
        - name: ZAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: zai-secrets
              key: api-key
```

## Development Workflow

```
1. Develop TypeScript translator
   ├── npm run dev (sidecar mode with hot reload)
   ├── npm run test
   └── npm run build:wasm

2. Develop Go server
   ├── go run cmd/serve.go (uses sidecar by default)
   └── go test ./...

3. Integration testing
   ├── make test-integration
   └── make test-e2e

4. Build for production
   ├── make build-wasm
   ├── make build-binary
   └── Output: single binary with embedded WASM
```

## Testing Strategy

1. **Unit Tests**
   - TypeScript transformers (Jest)
   - Go handlers (Go test)

2. **Integration Tests**
   - Mock z.ai responses
   - Test transformation round-trips

3. **E2E Tests**
   - Real z.ai API (with test keys)
   - Full request/response cycle

4. **Conformance Tests**
   - Verify Responses API compliance
   - Test edge cases

## Monitoring & Observability

1. **Metrics**
   - Request count, latency, errors
   - Upstream API call metrics
   - Cache hit rates

2. **Logging**
   - Structured JSON logs
   - Request/response tracing
   - Error stack traces

3. **Health Checks**
   - `/health` endpoint
   - Dependency health (z.ai connectivity)

## Future Enhancements

1. **Multi-Backend Support**: Add support for other Chat API providers
2. **Request Batching**: Batch multiple requests for efficiency
3. **Semantic Caching**: Cache by semantic similarity
4. **Request Transformation**: Apply prompt engineering automatically
5. **Cost Tracking**: Track token usage and costs

## References

- [OpenAI Responses API Reference](https://platform.openai.com/docs/api-reference/responses)
- [Z.AI Chat Completion API](https://docs.z.ai/api-reference/llm/chat-completion)
- [OpenAI Chat Completions API](https://platform.openai.com/docs/api-reference/chat)
- [Server-Sent Events Specification](https://html.spec.whatwg.org/multipage/server-sent-events.html)
