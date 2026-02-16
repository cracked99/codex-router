/**
 * Codex API Router - TypeScript Translator
 *
 * Transforms between OpenAI Responses API and Chat Completions API formats.
 * This module provides bidirectional transformation for requests, responses,
 * and streaming events.
 *
 * @module @codex-api-router/translator
 */

// ============================================================================
// Type Exports
// ============================================================================

export * from './types/responses.js';
export * from './types/chat.js';
export * from './types/common.js';

// ============================================================================
// Transformer Exports
// ============================================================================

export { transformRequest } from './transformers/request.js';

export {
  transformResponse,
  transformErrorResponse,
  extractTextFromDelta,
  extractToolCallsFromDelta,
} from './transformers/response.js';

export {
  StreamTransformer,
  parseSSELine,
  parseSSEChunk,
  formatSSEEvent,
} from './transformers/stream.js';
export type { StreamState } from './transformers/stream.js';

// ============================================================================
// Public API
// ============================================================================

import { transformRequest } from './transformers/request.js';
import { transformResponse, transformErrorResponse } from './transformers/response.js';
import { StreamTransformer } from './transformers/stream.js';
import type {
  ResponsesRequest,
  ResponsesResponse,
  ResponsesSSEEvent,
} from './types/responses.js';
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  ChatCompletionChunk,
} from './types/chat.js';
import type { TransformOptions } from './types/common.js';

/**
 * Translator class for API transformations
 */
export class Translator {
  private streamTransformer?: StreamTransformer;
  private options: TransformOptions;

  constructor(options: TransformOptions = {}) {
    this.options = options;
  }

  /**
   * Transform a Responses API request to Chat Completions format
   */
  transformRequest(request: ResponsesRequest): ChatCompletionRequest {
    return transformRequest(request, this.options);
  }

  /**
   * Transform a Chat Completions response to Responses API format
   */
  transformResponse(response: ChatCompletionResponse): ResponsesResponse {
    return transformResponse(response, this.options);
  }

  /**
   * Transform a Chat Completions error to Responses API format
   */
  transformError(error: {
    type?: string;
    message: string;
    param?: string;
    code?: string;
  }): ReturnType<typeof transformErrorResponse> {
    return transformErrorResponse(error);
  }

  /**
   * Get a stream transformer for SSE transformation
   */
  getStreamTransformer(): StreamTransformer {
    if (!this.streamTransformer) {
      this.streamTransformer = new StreamTransformer();
    }
    return this.streamTransformer;
  }

  /**
   * Transform a streaming SSE chunk
   */
  transformStreamChunk(chunk: ChatCompletionChunk): ResponsesSSEEvent[] {
    return this.getStreamTransformer().transformChunk(chunk);
  }

  /**
   * Reset stream transformer state
   */
  resetStream(): void {
    const transformer = this.getStreamTransformer();
    transformer.reset();
  }
}

// ============================================================================
// Convenience Functions
// ============================================================================

/**
 * Create a new translator with the given options
 */
export function createTranslator(options?: TransformOptions): Translator {
  return new Translator(options);
}

/**
 * Quick transformation for Requests API -> Chat Completions
 */
export function toChatCompletions(
  request: ResponsesRequest,
  options?: TransformOptions
): ChatCompletionRequest {
  return transformRequest(request, options);
}

/**
 * Quick transformation for Chat Completions -> Responses API
 */
export function toResponses(
  response: ChatCompletionResponse,
  options?: TransformOptions
): ResponsesResponse {
  return transformResponse(response, options);
}
