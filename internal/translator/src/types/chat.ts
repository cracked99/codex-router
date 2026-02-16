/**
 * Chat Completions API Types
 * Based on OpenAI Chat Completions API specification
 */

// ============================================================================
// Request Types
// ============================================================================

export interface ChatCompletionRequest {
  model: string;
  messages: ChatMessage[];
  temperature?: number;
  max_completion_tokens?: number;
  max_tokens?: number;
  top_p?: number;
  stream?: boolean;
  store?: boolean;
  metadata?: Record<string, unknown>;
  tools?: ChatTool[];
  tool_choice?: ChatToolChoice;
  parallel_tool_calls?: boolean;
  response_format?: ResponseFormat;
  reasoning_effort?: 'low' | 'medium' | 'high';
  user?: string;
}

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'tool';
  content: string | ChatMessageContent[];
  name?: string;
  tool_calls?: ToolCall[];
  tool_call_id?: string;
}

export type ChatMessageContent = ChatTextContent | ImageUrlContent;

export interface ChatTextContent {
  type: 'text';
  text: string;
}

export interface ImageUrlContent {
  type: 'image_url';
  image_url: string | ImageUrlDetail;
}

export interface ImageUrlDetail {
  url: string;
  detail?: 'low' | 'high' | 'auto';
}

export interface ChatTool {
  type: 'function';
  function: ChatToolFunction;
}

export interface ChatToolFunction {
  name: string;
  description?: string;
  parameters: ChatToolParameters;
}

export interface ChatToolParameters {
  type: 'object';
  properties: Record<string, ChatToolProperty>;
  required?: string[];
  additionalProperties?: boolean;
  [key: string]: unknown;
}

export interface ChatToolProperty {
  type: string;
  description?: string;
  enum?: string[];
  [key: string]: unknown;
}

export type ChatToolChoice =
  | 'auto'
  | 'none'
  | 'required'
  | ChatToolChoiceOption;

export interface ChatToolChoiceOption {
  type: 'function';
  name: string;
}

export interface ResponseFormat {
  type: 'text' | 'json_object' | 'json_schema';
  json_schema?: JsonSchema;
}

export interface JsonSchema {
  name: string;
  strict?: boolean;
  schema: Record<string, unknown>;
}

// ============================================================================
// Response Types
// ============================================================================

export interface ChatCompletionResponse {
  id: string;
  object: 'chat.completion' | 'chat.completion.chunk';
  created: number;
  model: string;
  choices: ChatChoice[];
  usage?: ChatCompletionUsage;
}

export interface ChatChoice {
  index: number;
  message: ChatMessage;
  finish_reason: FinishReason | null;
}

export type FinishReason = 'stop' | 'length' | 'tool_calls' | 'content_filter';

export interface ChatCompletionUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

// ============================================================================
// Streaming Types
// ============================================================================

export interface ChatCompletionChunk {
  id: string;
  object: 'chat.completion.chunk';
  created: number;
  model: string;
  choices: ChatChoiceDelta[];
  usage?: ChatCompletionUsage;
}

export interface ChatChoiceDelta {
  index: number;
  delta: Delta;
  finish_reason: FinishReason | null;
}

export interface Delta {
  role?: 'assistant';
  content?: string;
  tool_calls?: ToolCallDelta[];
}

export interface ToolCallDelta {
  index: number;
  id?: string;
  type?: 'function';
  function?: {
    name?: string;
    arguments?: string;
  };
}

export interface ToolCall {
  id: string;
  type: 'function';
  function: ToolCallFunction;
}

export interface ToolCallFunction {
  name: string;
  arguments: string;
}
