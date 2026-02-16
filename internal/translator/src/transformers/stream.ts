/**
 * Stream Transformer
 * Handles SSE stream transformation between Chat Completions and Responses API formats
 */

import type {
  ChatCompletionChunk,
  Delta,
} from '../types/chat.js';
import type {
  ResponsesSSEEvent,
  ResponseCreatedEvent,
  ResponseDeltaEvent,
  ResponseCompletedEvent,
  OutputTextDelta,
} from '../types/responses.js';
import {
  generateResponseId,
  generateMessageId,
  generateCallId,
  FINISH_REASON_MAPPING,
} from '../types/common.js';

// Local type for function call deltas
interface FunctionCallDelta {
  type: 'function_call';
  id: string;
  name: string;
  arguments: string;
}

// ============================================================================
// Stream Transformer State
// ============================================================================

export interface StreamState {
  responseId: string;
  messageId: string;
  createdAt: number;
  model: string;
  outputIndex: number;
  currentText: string;
  currentToolCalls: Map<number, Partial<FunctionCallDelta>>;
  isComplete: boolean;
}

export class StreamTransformer {
  private state: StreamState | null = null;

  /**
   * Transform a Chat Completions SSE chunk to Responses API format
   */
  transformChunk(chunk: ChatCompletionChunk): ResponsesSSEEvent[] {
    // Initialize state on first chunk
    if (!this.state) {
      this.state = this.initializeState(chunk);
      return [this.createCreatedEvent()];
    }

    const events: ResponsesSSEEvent[] = [];

    // Check if this is a completion chunk
    const choice = chunk.choices[0];
    if (!choice) {
      return events;
    }

    // Handle finish reason (stream end)
    if (choice.finish_reason) {
      events.push(...this.finalizeStream(choice.finish_reason));
      return events;
    }

    // Handle delta content
    if (choice.delta) {
      const deltaEvents = this.transformDelta(choice.delta, choice.index);
      events.push(...deltaEvents);
    }

    return events;
  }

  /**
   * Get the final accumulated state as a complete response
   */
  getFinalState(): Partial<StreamState> | null {
    return this.state;
  }

  /**
   * Reset the transformer state
   */
  reset(): void {
    this.state = null;
  }

  // ============================================================================
  // Private Methods
  // ============================================================================

  private initializeState(chunk: ChatCompletionChunk): StreamState {
    return {
      responseId: generateResponseId(chunk.id),
      messageId: generateMessageId(),
      createdAt: chunk.created,
      model: chunk.model,
      outputIndex: 0,
      currentText: '',
      currentToolCalls: new Map(),
      isComplete: false,
    };
  }

  private createCreatedEvent(): ResponseCreatedEvent {
    if (!this.state) throw new Error('State not initialized');

    return {
      type: 'response.created',
      response_id: this.state.responseId,
      created_at: this.state.createdAt,
      model: this.state.model,
    };
  }

  private transformDelta(delta: Delta, _choiceIndex: number): ResponsesSSEEvent[] {
    if (!this.state) throw new Error('State not initialized');

    const events: ResponsesSSEEvent[] = [];

    // Handle text content delta
    if (delta.content) {
      events.push(this.createTextDeltaEvent(delta.content));
      this.state.currentText += delta.content;
    }

    // Handle tool calls delta
    if (delta.tool_calls && delta.tool_calls.length > 0) {
      for (const toolCall of delta.tool_calls) {
        const toolCallEvents = this.transformToolCallDelta(toolCall);
        events.push(...toolCallEvents);
      }
    }

    return events;
  }

  private createTextDeltaEvent(text: string): ResponseDeltaEvent {
    if (!this.state) throw new Error('State not initialized');

    const textDelta: OutputTextDelta = {
      type: 'output_text',
      text,
    };

    return {
      type: 'response.delta',
      delta: textDelta,
      item_id: this.state.messageId,
      output_index: this.state.outputIndex,
      response_id: this.state.responseId,
    };
  }

  private transformToolCallDelta(
    toolCallDelta: { index: number; id?: string; function?: { name?: string; arguments?: string } }
  ): ResponseDeltaEvent[] {
    if (!this.state) throw new Error('State not initialized');

    const events: ResponseDeltaEvent[] = [];
    const { index, id, function: fn } = toolCallDelta;

    // Get or create tool call state
    let toolCallState = this.state.currentToolCalls.get(index);
    if (!toolCallState) {
      toolCallState = {
        type: 'function_call',
        id: '',
        name: '',
        arguments: '',
      };
      this.state.currentToolCalls.set(index, toolCallState);
    }

    // Update tool call state
    if (id) {
      toolCallState.id = generateCallId(id);
    }
    if (fn?.name) {
      toolCallState.name = fn.name;
    }
    if (fn?.arguments) {
      toolCallState.arguments += fn.arguments;
    }

    // Create delta event if we have meaningful data
    if (fn?.name || fn?.arguments) {
      const functionDelta = {
        type: 'function_call' as const,
        id: toolCallState.id || '',
        name: toolCallState.name || '',
        arguments: fn?.arguments || '',
      };

      events.push({
        type: 'response.delta',
        delta: functionDelta,
        item_id: this.state.messageId,
        output_index: this.state.outputIndex,
        response_id: this.state.responseId,
      });
    }

    return events;
  }

  private finalizeStream(finishReason: string): ResponsesSSEEvent[] {
    if (!this.state) throw new Error('State not initialized');

    const events: ResponsesSSEEvent[] = [];

    // Mark as complete
    this.state.isComplete = true;

    // Map finish reason to status
    const status = FINISH_REASON_MAPPING[finishReason] || 'completed';

    // Create completed event
    const completedEvent: ResponseCompletedEvent = {
      type: 'response.completed',
      response: {
        id: this.state.responseId,
        object: 'response',
        created_at: this.state.createdAt,
        status: status as any,
        model: this.state.model,
        output: this.buildOutputItems(),
      },
    };

    events.push(completedEvent);

    return events;
  }

  private buildOutputItems() {
    if (!this.state) throw new Error('State not initialized');

    const items: any[] = [];

    // Build message item
    const messageItem = {
      type: 'message',
      id: this.state.messageId,
      status: 'completed',
      role: 'assistant',
      content: [] as any[],
    };

    // Add text content
    if (this.state.currentText) {
      messageItem.content.push({
        type: 'output_text',
        text: this.state.currentText,
        annotations: [],
      });
    }

    // Add tool calls
    for (const [_, toolCall] of this.state.currentToolCalls) {
      if (toolCall.id && toolCall.name) {
        messageItem.content.push({
          type: 'function_call',
          id: toolCall.id,
          name: toolCall.name,
          arguments: toolCall.arguments,
        });
      }
    }

    items.push(messageItem);

    return items;
  }
}

// ============================================================================
// SSE Parsing Utilities
// ============================================================================

export function parseSSELine(line: string): { event?: string; data?: string } | null {
  const trimmedLine = line.trim();

  if (!trimmedLine || trimmedLine.startsWith(':')) {
    return null; // Skip empty lines and comments
  }

  if (trimmedLine === 'data: [DONE]') {
    return { data: '[DONE]' };
  }

  const eventMatch = trimmedLine.match(/^event:\s*(.+)$/);
  const dataMatch = trimmedLine.match(/^data:\s*(.+)$/);

  if (eventMatch) {
    return { event: eventMatch[1].trim() };
  }

  if (dataMatch) {
    return { data: dataMatch[1].trim() };
  }

  return null;
}

export function parseSSEChunk(data: string): ChatCompletionChunk | null {
  try {
    if (data === '[DONE]') {
      return null;
    }
    return JSON.parse(data) as ChatCompletionChunk;
  } catch {
    return null;
  }
}

export function formatSSEEvent(event: ResponsesSSEEvent): string[] {
  const lines: string[] = [];

  // Extract event type
  let eventType = 'message';
  if ('type' in event) {
    eventType = (event as any).type;
  }

  // Format data
  const data = JSON.stringify(event);

  lines.push(`event: ${eventType}`);
  lines.push(`data: ${data}`);
  lines.push(''); // Empty line to end the event

  return lines;
}
