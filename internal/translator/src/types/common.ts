/**
 * Common types and utilities shared between transformers
 */

import type { ResponsesRequest, ResponsesResponse } from './responses.js';
import type { ChatCompletionRequest, ChatMessage } from './chat.js';

// ============================================================================
// Public API - Main transform functions
// ============================================================================

export interface TransformResult<T> {
  success: boolean;
  data?: T;
  errors?: TransformError[];
}

export interface TransformError {
  field: string;
  message: string;
  value?: unknown;
}

export interface TransformOptions {
  strict?: boolean;
  includeMetadata?: boolean;
  modelMapping?: Record<string, string>;
}

export interface TranslatorState {
  conversationHistory: ChatMessage[];
  responseCache: Map<string, ResponsesResponse>;
}

// ============================================================================
// Model Mapping
// ============================================================================

export const DEFAULT_MODEL_MAPPING: Record<string, string> = {
  'gpt-4.1': 'glm-5',
  'gpt-4o': 'glm-5',
  'gpt-4o-2024-08-06': 'glm-5',
  'gpt-5': 'glm-5',
  'gpt-4-turbo': 'glm-5',
  'gpt-4': 'glm-5',
  'gpt-3.5-turbo': 'glm-3-turbo',
};

// ============================================================================
// Finish Reason Mapping
// ============================================================================

export const FINISH_REASON_MAPPING: Record<string, string> = {
  stop: 'completed',
  length: 'incomplete',
  tool_calls: 'completed',
  content_filter: 'failed',
};

export const STATUS_TO_FINISH_REASON: Record<string, string> = {
  completed: 'stop',
  incomplete: 'length',
  failed: 'content_filter',
  cancelled: 'stop',
  in_progress: 'stop',
};

// ============================================================================
// Validation Utilities
// ============================================================================

export function validateResponsesRequest(
  request: ResponsesRequest
): TransformError[] {
  const errors: TransformError[] = [];

  if (!request.model) {
    errors.push({ field: 'model', message: 'Model is required' });
  }

  if (!request.input) {
    errors.push({ field: 'input', message: 'Input is required' });
  }

  if (request.temperature !== undefined) {
    if (typeof request.temperature !== 'number' || request.temperature < 0 || request.temperature > 2) {
      errors.push({
        field: 'temperature',
        message: 'Temperature must be between 0 and 2',
        value: request.temperature,
      });
    }
  }

  if (request.max_output_tokens !== undefined) {
    if (typeof request.max_output_tokens !== 'number' || request.max_output_tokens < 1) {
      errors.push({
        field: 'max_output_tokens',
        message: 'max_output_tokens must be a positive number',
        value: request.max_output_tokens,
      });
    }
  }

  return errors;
}

export function validateChatCompletionRequest(
  request: ChatCompletionRequest
): TransformError[] {
  const errors: TransformError[] = [];

  if (!request.model) {
    errors.push({ field: 'model', message: 'Model is required' });
  }

  if (!request.messages || request.messages.length === 0) {
    errors.push({ field: 'messages', message: 'At least one message is required' });
  }

  if (request.temperature !== undefined) {
    if (typeof request.temperature !== 'number' || request.temperature < 0 || request.temperature > 2) {
      errors.push({
        field: 'temperature',
        message: 'Temperature must be between 0 and 2',
        value: request.temperature,
      });
    }
  }

  return errors;
}

// ============================================================================
// ID Generation
// ============================================================================

export function generateResponseId(originalId?: string): string {
  if (originalId) {
    return originalId.startsWith('resp_') ? originalId : `resp_${originalId}`;
  }
  return `resp_${generateId()}`;
}

export function generateMessageId(originalId?: string): string {
  if (originalId) {
    return originalId.startsWith('msg_') ? originalId : `msg_${originalId}`;
  }
  return `msg_${generateId()}`;
}

export function generateCallId(originalId?: string): string {
  if (originalId) {
    return originalId.startsWith('call_') ? originalId.replace('call_', 'fc_') : `fc_${originalId}`;
  }
  return `fc_${generateId()}`;
}

function generateId(): string {
  return Math.random().toString(36).substring(2, 15) +
         Math.random().toString(36).substring(2, 15);
}
