/**
 * Request Transformer
 * Transforms Responses API requests to Chat Completions API format
 */

import type {
  ResponsesRequest,
  InputItem,
  MessageInputItem,
  InputTextItem,
  InputImageItem,
  Tool,
  ToolChoice,
  TextConfig,
} from '../types/responses.js';
import type {
  ChatCompletionRequest,
  ChatMessage,
  ChatTool,
  ChatToolChoice,
  ResponseFormat,
} from '../types/chat.js';
import {
  DEFAULT_MODEL_MAPPING,
  validateResponsesRequest,
  type TransformOptions,
} from '../types/common.js';

// ============================================================================
// Main Transform Function
// ============================================================================

export function transformRequest(
  request: ResponsesRequest,
  options: TransformOptions = {}
): ChatCompletionRequest {
  const modelMapping = { ...DEFAULT_MODEL_MAPPING, ...(options.modelMapping || {}) };

  // Validate input if in strict mode
  const errors = validateResponsesRequest(request);
  if (options.strict && errors.length > 0) {
    throw new Error(`Request validation failed: ${errors.map(e => e.message).join(', ')}`);
  }

  // Map model name
  const model = modelMapping[request.model] || request.model;

  // Build messages array
  const messages = transformInputToMessages(request);

  // Transform tools if present
  const tools = request.tools ? transformTools(request.tools) : undefined;

  // Transform tool_choice if present
  const toolChoice = request.tool_choice
    ? transformToolChoice(request.tool_choice)
    : undefined;

  // Transform response format if present
  const responseFormat = request.text?.format
    ? transformResponseFormat(request.text.format)
    : undefined;

  // Build the Chat Completions request
  const chatRequest: ChatCompletionRequest = {
    model,
    messages,
    ...(request.temperature !== undefined && { temperature: request.temperature }),
    ...(request.max_output_tokens !== undefined && {
      max_completion_tokens: request.max_output_tokens,
    }),
    ...(request.top_p !== undefined && { top_p: request.top_p }),
    ...(request.stream !== undefined && { stream: request.stream }),
    ...(request.store !== undefined && { store: request.store }),
    ...(request.metadata && { metadata: request.metadata }),
    ...(tools && { tools }),
    ...(toolChoice && { tool_choice: toolChoice }),
    ...(request.parallel_tool_calls !== undefined && {
      parallel_tool_calls: request.parallel_tool_calls,
    }),
    ...(responseFormat && { response_format: responseFormat }),
    ...(request.effort && { reasoning_effort: request.effort }),
  };

  return chatRequest;
}

// ============================================================================
// Message Transformation
// ============================================================================

function transformInputToMessages(request: ResponsesRequest): ChatMessage[] {
  const messages: ChatMessage[] = [];

  // Add system message from instructions if present
  if (request.instructions) {
    messages.push({
      role: 'system',
      content: request.instructions,
    });
  }

  // Transform input to messages
  if (typeof request.input === 'string') {
    // Simple string input becomes a user message
    messages.push({
      role: 'user',
      content: request.input,
    });
  } else if (Array.isArray(request.input)) {
    // Array of input items
    const transformedItems = transformInputItems(request.input);
    messages.push(...transformedItems);
  }

  return messages;
}

function transformInputItems(items: InputItem[]): ChatMessage[] {
  const messages: ChatMessage[] = [];

  for (const item of items) {
    switch (item.type) {
      case 'message': {
        const msg = item as MessageInputItem;
        messages.push({
          role: msg.role,
          content: transformMessageContent(msg.content),
        });
        break;
      }
      case 'input_text': {
        const textItem = item as InputTextItem;
        messages.push({
          role: 'user',
          content: textItem.text,
        });
        break;
      }
      case 'input_image': {
        const imageItem = item as InputImageItem;
        messages.push({
          role: 'user',
          content: [
            {
              type: 'image_url',
              image_url: {
                url: imageItem.image_url,
                ...(imageItem.detail && { detail: imageItem.detail }),
              },
            },
          ],
        });
        break;
      }
      case 'function_call':
      case 'function_call_output':
        // These are handled in conversation state, not in initial request
        break;
    }
  }

  return messages;
}

function transformMessageContent(content: unknown[]): string {
  // Extract text from content blocks
  if (Array.isArray(content)) {
    const textParts: string[] = [];
    for (const block of content) {
      if (typeof block === 'object' && block !== null) {
        const type = (block as { type?: string }).type;
        if (type === 'input_text' || type === 'text') {
          const textBlock = block as { text?: string };
          if (textBlock.text) {
            textParts.push(textBlock.text);
          }
        } else if ('text' in block && typeof block.text === 'string') {
          textParts.push(block.text);
        }
      }
    }
    return textParts.join('\n');
  }
  return String(content);
}

// ============================================================================
// Tool Transformation
// ============================================================================

function transformTools(tools: Tool[]): ChatTool[] {
  return tools
    .filter(tool => tool.type === 'function')
    .map(tool => ({
      type: 'function',
      function: {
        name: tool.name || '',
        description: tool.description,
        parameters: tool.parameters || { type: 'object', properties: {} },
      },
    }));
}

function transformToolChoice(toolChoice: ToolChoice): ChatToolChoice {
  if (typeof toolChoice === 'string') {
    return toolChoice as 'auto' | 'none' | 'required';
  }
  return {
    type: 'function',
    name: toolChoice.name,
  };
}

// ============================================================================
// Response Format Transformation
// ============================================================================

function transformResponseFormat(format: TextConfig['format']): ResponseFormat | undefined {
  if (!format) return undefined;

  if (format.type === 'text') {
    return { type: 'text' };
  }

  if (format.type === 'json_schema') {
    return {
      type: 'json_schema',
      json_schema: {
        name: format.name || 'schema',
        strict: format.strict,
        schema: format.schema || {},
      },
    };
  }

  return undefined;
}
