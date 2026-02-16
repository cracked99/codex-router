# @plasmadev/codex-translator

TypeScript translator for Codex API Router - transforms between OpenAI Responses API and Chat Completions API formats.

## Overview

This module provides bidirectional transformation between:
- **OpenAI Responses API** (used by Codex CLI v0.99+)
- **OpenAI Chat Completions API** (used by z.ai)

## Installation

```bash
npm install
```

## Building

```bash
npm run build
```

## Testing

```bash
npm test
```

## Usage

### Basic Usage

```typescript
import { transformRequest, transformResponse } from '@plasmadev/codex-translator';

// Transform Responses API request to Chat Completions format
const responsesRequest = {
  model: 'gpt-4',
  input: 'Hello, how are you?',
  instructions: 'You are a helpful assistant.'
};

const chatRequest = transformRequest(responsesRequest);
console.log(chatRequest);
// {
//   model: 'glm-5',
//   messages: [
//     { role: 'system', content: 'You are a helpful assistant.' },
//     { role: 'user', content: 'Hello, how are you?' }
//   ]
// }

// Transform Chat Completions response to Responses API format
const chatResponse = {
  id: 'chatcmpl-123',
  object: 'chat.completion',
  created: 1234567890,
  model: 'glm-5',
  choices: [{
    index: 0,
    message: {
      role: 'assistant',
      content: 'I am doing well, thank you!'
    },
    finish_reason: 'stop'
  }],
  usage: {
    prompt_tokens: 10,
    completion_tokens: 8,
    total_tokens: 18
  }
};

const responsesResponse = transformResponse(chatResponse);
console.log(responsesResponse);
// {
//   id: 'resp_xxx',
//   object: 'response',
//   created_at: 1234567890,
//   status: 'completed',
//   model: 'glm-5',
//   output: [{
//     type: 'message',
//     id: 'msg_xxx',
//     status: 'completed',
//     role: 'assistant',
//     content: [{
//       type: 'output_text',
//       text: 'I am doing well, thank you!',
//       annotations: []
//     }]
//   }],
//   usage: {
//     input_tokens: 10,
//     output_tokens: 8,
//     total_tokens: 18
//   }
// }
```

### Streaming

```typescript
import { StreamTransformer, parseSSEChunk } from '@plasmadev/codex-translator';

const transformer = new StreamTransformer();

// Process SSE chunks from Chat Completions API
for await (const line of sseStream) {
  const chunk = parseSSEChunk(line);
  if (chunk) {
    const events = transformer.transformChunk(chunk);
    for (const event of events) {
      // Send transformed event to client
      sendSSEEvent(formatSSEEvent(event));
    }
  }
}
```

### Advanced Options

```typescript
import { createTranslator } from '@plasmadev/codex-translator';

const translator = createTranslator({
  strict: true,              // Enable strict validation
  includeMetadata: true,     // Include metadata in responses
  modelMapping: {            // Custom model mapping
    'custom-model': 'mapped-model'
  }
});

// Use the translator instance
const chatRequest = translator.transformRequest(responsesRequest);
const responsesResponse = translator.transformResponse(chatResponse);
```

## API Mappings

### Request Fields

| Responses API | Chat Completions API |
|---------------|---------------------|
| `input` (string) | `messages` (user message) |
| `instructions` | `messages[0]` (system message) |
| `max_output_tokens` | `max_completion_tokens` |
| `tools` | `tools` (with nested `function`) |

### Response Fields

| Chat Completions | Responses API |
|------------------|---------------|
| `id` | `id` (with `resp_` prefix) |
| `created` | `created_at` |
| `choices[0].message.content` | `output[0].content[0].text` |
| `usage.prompt_tokens` | `usage.input_tokens` |
| `usage.completion_tokens` | `usage.output_tokens` |
| `finish_reason` | `status` (mapped) |

## Project Structure

```
src/
├── index.ts              # Main entry point
├── types/
│   ├── responses.ts      # Responses API types
│   ├── chat.ts           # Chat Completions types
│   └── common.ts         # Shared types and utilities
├── transformers/
│   ├── request.ts        # Request transformation
│   ├── response.ts       # Response transformation
│   └── stream.ts         # SSE stream transformation
└── utils/                # Utility functions

tests/
└── transformers.test.ts  # Unit tests
```

## License

MIT
