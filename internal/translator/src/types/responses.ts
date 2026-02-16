/**
 * Responses API Types
 * Based on OpenAI Responses API specification
 */

// ============================================================================
// Request Types
// ============================================================================

export interface ResponsesRequest {
  model: string;
  input: string | InputItem[];
  instructions?: string;
  temperature?: number;
  max_output_tokens?: number;
  top_p?: number;
  tools?: Tool[];
  tool_choice?: ToolChoice;
  parallel_tool_calls?: boolean;
  stream?: boolean;
  store?: boolean;
  metadata?: Record<string, unknown>;
  previous_response_id?: string;
  truncation?: TruncationConfig;
  text?: TextConfig;
  effort?: 'low' | 'medium' | 'high';
  verbosity?: 'low' | 'medium' | 'high';
  include?: string[];
  conversation?: ConversationConfig;
}

export type InputItem =
  | MessageInputItem
  | InputTextItem
  | InputImageItem
  | FunctionCallInputItem
  | FunctionCallOutputItem;

export interface MessageInputItem {
  type: 'message';
  role: 'system' | 'user' | 'assistant';
  content: MessageContent[];
}

export interface InputTextItem {
  type: 'input_text';
  text: string;
}

export interface InputImageItem {
  type: 'input_image';
  image_url: string;
  detail?: 'low' | 'high' | 'auto';
}

export interface FunctionCallInputItem {
  type: 'function_call';
  id: string;
  name: string;
  arguments: string;
}

export interface FunctionCallOutputItem {
  type: 'function_call_output';
  call_id: string;
  output: string;
}

export type MessageContent = InputTextContent | InputImageContent;

export interface InputTextContent {
  type: 'input_text';
  text: string;
}

export interface InputImageContent {
  type: 'input_image';
  image_url: string;
  detail?: 'low' | 'high' | 'auto';
}

export interface Tool {
  type: 'function' | 'web_search' | 'file_search' | 'code_interpreter';
  name?: string;
  description?: string;
  parameters?: ToolParameters;
}

export interface ToolParameters {
  type: 'object';
  properties: Record<string, ToolProperty>;
  required?: string[];
  additionalProperties?: boolean;
  [key: string]: unknown;
}

export interface ToolProperty {
  type: string;
  description?: string;
  enum?: string[];
  [key: string]: unknown;
}

export type ToolChoice = 'auto' | 'none' | 'required' | ToolChoiceOption;

export interface ToolChoiceOption {
  type: 'function';
  name: string;
}

export interface TruncationConfig {
  type?: 'auto' | 'messages' | 'auto_messages';
  max_messages?: number;
  max_tokens?: number;
}

export interface TextConfig {
  format?: TextFormat;
}

export interface TextFormat {
  type: 'text' | 'json_schema';
  name?: string;
  strict?: boolean;
  schema?: Record<string, unknown>;
}

export interface ConversationConfig {
  state?: ConversationState;
}

export type ConversationState = 'connected' | 'disconnected';

// ============================================================================
// Response Types
// ============================================================================

export interface ResponsesResponse {
  id: string;
  object: 'response';
  created_at: number;
  status: ResponseStatus;
  model: string;
  output: OutputItem[];
  error?: ErrorResponse;
  usage?: ResponseUsage;
  metadata?: Record<string, unknown>;
}

export type ResponseStatus =
  | 'in_progress'
  | 'completed'
  | 'failed'
  | 'incomplete'
  | 'cancelled';

export type OutputItem = MessageOutputItem | ReasoningOutputItem;

export interface MessageOutputItem {
  type: 'message';
  id: string;
  status: ResponseStatus;
  role: 'assistant';
  content: OutputContent[];
}

export interface ReasoningOutputItem {
  type: 'reasoning';
  id: string;
  status: ResponseStatus;
  summary: SummaryContent[];
}

export type OutputContent = OutputTextContent | OutputCallContent;

export interface OutputTextContent {
  type: 'output_text';
  text: string;
  annotations?: unknown[];
}

export interface OutputCallContent {
  type: 'function_call';
  id: string;
  name: string;
  arguments: string;
}

export type SummaryContent = SummaryTextContent;

export interface SummaryTextContent {
  type: 'summary_text';
  text: string;
}

export interface ResponseUsage {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
}

export interface ErrorResponse {
  type: ErrorType;
  message: string;
  param?: string;
  code?: string;
}

export type ErrorType =
  | 'invalid_request_error'
  | 'authentication_error'
  | 'rate_limit_error'
  | 'api_error'
  | 'content_policy_error';

// ============================================================================
// Streaming Types
// ============================================================================

export type ResponsesSSEEvent =
  | ResponseCreatedEvent
  | ResponseOutputItemAddedEvent
  | ResponseOutputItemDoneEvent
  | ResponseDeltaEvent
  | ResponseCompletedEvent
  | ResponseFailedEvent;

export interface ResponseCreatedEvent {
  type: 'response.created';
  response_id: string;
  created_at: number;
  model: string;
}

export interface ResponseOutputItemAddedEvent {
  type: 'response.output_item.added';
  output_item: OutputItem;
  response_id: string;
}

export interface ResponseOutputItemDoneEvent {
  type: 'response.output_item.done';
  output_item: OutputItem;
  response_id: string;
}

export interface ResponseDeltaEvent {
  type: 'response.delta';
  delta: DeltaContent;
  item_id: string;
  output_index: number;
  response_id: string;
}

export type DeltaContent = OutputTextDelta | FunctionCallDelta;

export interface OutputTextDelta {
  type: 'output_text';
  text: string;
}

export interface FunctionCallDelta {
  type: 'function_call';
  id: string;
  name: string;
  arguments: string;
}

export interface ResponseCompletedEvent {
  type: 'response.completed';
  response: ResponsesResponse;
}

export interface ResponseFailedEvent {
  type: 'response.failed';
  response_id: string;
  error: ErrorResponse;
}
