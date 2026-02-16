/**
 * Response Transformer
 * Transforms Chat Completions API responses to Responses API format
 */

import type {
  ChatCompletionResponse,
  ChatChoice,
} from '../types/chat.js';
import type {
  ResponsesResponse,
  OutputItem,
  MessageOutputItem,
  OutputTextContent,
  OutputCallContent,
  ResponseUsage,
  ResponseStatus,
} from '../types/responses.js';
import {
  generateResponseId,
  generateMessageId,
  generateCallId,
  FINISH_REASON_MAPPING,
  type TransformOptions,
} from '../types/common.js';

// ============================================================================
// Main Transform Function
// ============================================================================

export function transformResponse(
  response: ChatCompletionResponse,
  options: TransformOptions = {}
): ResponsesResponse {
  // Get the first choice (Chat Completions typically returns one)
  const choice = response.choices[0];
  if (!choice) {
    throw new Error('No choices in Chat Completions response');
  }

  // Generate IDs
  const responseId = generateResponseId(response.id);
  const messageId = generateMessageId();

  // Transform the choice to output items
  const outputItems = transformChoiceToOutputItems(choice, messageId);

  // Transform usage
  const usage = response.usage
    ? transformUsage(response.usage)
    : undefined;

  // Determine status from finish_reason
  const status = (choice.finish_reason
    ? (FINISH_REASON_MAPPING[choice.finish_reason] || 'completed')
    : 'completed') as ResponseStatus;

  // Build the Responses API response
  const responsesResponse: ResponsesResponse = {
    id: responseId,
    object: 'response',
    created_at: response.created,
    status,
    model: response.model,
    output: outputItems,
    ...(usage && { usage }),
    ...(options.includeMetadata && response.id && {
      metadata: { original_id: response.id },
    }),
  };

  return responsesResponse;
}

// ============================================================================
// Choice to Output Items Transformation
// ============================================================================

function transformChoiceToOutputItems(
  choice: ChatChoice,
  messageId: string
): OutputItem[] {
  const items: OutputItem[] = [];

  // Create message output item
  const messageItem: MessageOutputItem = {
    type: 'message',
    id: messageId,
    status: 'completed',
    role: 'assistant',
    content: [],
  };

  // Add text content if present
  if (choice.message.content) {
    const textContent: OutputTextContent = {
      type: 'output_text',
      text: extractContentText(choice.message.content),
      annotations: [],
    };
    messageItem.content.push(textContent);
  }

  // Add tool calls if present
  if (choice.message.tool_calls && choice.message.tool_calls.length > 0) {
    for (const toolCall of choice.message.tool_calls) {
      const callContent: OutputCallContent = {
        type: 'function_call',
        id: generateCallId(toolCall.id),
        name: toolCall.function.name,
        arguments: toolCall.function.arguments,
      };
      messageItem.content.push(callContent);
    }
  }

  items.push(messageItem);

  return items;
}

// ============================================================================
// Content Extraction
// ============================================================================

function extractContentText(content: string | unknown[]): string {
  if (typeof content === 'string') {
    return content;
  }

  if (Array.isArray(content)) {
    const textParts: string[] = [];
    for (const block of content) {
      if (typeof block === 'object' && block !== null) {
        const type = (block as { type?: string }).type;
        if (type === 'text') {
          const textBlock = block as { text?: string };
          if (textBlock.text) {
            textParts.push(textBlock.text);
          }
        } else if ('text' in block && typeof block.text === 'string') {
          textParts.push(block.text);
        }
      }
    }
    return textParts.join('');
  }

  return String(content);
}

// ============================================================================
// Usage Transformation
// ============================================================================

function transformUsage(usage: {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}): ResponseUsage {
  return {
    input_tokens: usage.prompt_tokens,
    output_tokens: usage.completion_tokens,
    total_tokens: usage.total_tokens,
  };
}

// ============================================================================
// Error Transformation
// ============================================================================

export interface ErrorResponse {
  type: string;
  message: string;
  param?: string;
  code?: string;
}

export function transformErrorResponse(
  error: {
    type?: string;
    message: string;
    param?: string;
    code?: string;
  }
): ErrorResponse {
  return {
    type: error.type || 'api_error',
    message: error.message,
    ...(error.param && { param: error.param }),
    ...(error.code && { code: error.code }),
  };
}

// ============================================================================
// Streaming Response Helpers
// ============================================================================

export function extractTextFromDelta(delta: { content?: string }): string {
  return delta.content || '';
}

export function extractToolCallsFromDelta(
  delta: { tool_calls?: Array<{ index: number; id?: string; function?: { name?: string; arguments?: string } }> }
): Array<{ index: number; id?: string; name?: string; arguments?: string }> {
  if (!delta.tool_calls) return [];

  return delta.tool_calls.map(tc => ({
    index: tc.index,
    id: tc.id,
    name: tc.function?.name,
    arguments: tc.function?.arguments,
  }));
}
