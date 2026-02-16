/**
 * Unit tests for API transformers
 */

import { describe, test, expect } from '@jest/globals';
import { transformRequest } from '../src/transformers/request.js';
import { transformResponse } from '../src/transformers/response.js';
import { StreamTransformer } from '../src/transformers/stream.js';
import type { ResponsesRequest } from '../src/types/responses.js';
import type { ChatCompletionResponse, ChatCompletionChunk } from '../src/types/chat.js';

describe('Request Transformer', () => {
  describe('transformRequest', () => {
    test('should transform simple string input', () => {
      const request: ResponsesRequest = {
        model: 'gpt-4',
        input: 'Hello, how are you?',
      };

      const result = transformRequest(request);

      expect(result.model).toBe('glm-5'); // Model mapping
      expect(result.messages).toHaveLength(1);
      expect(result.messages[0].role).toBe('user');
      expect(result.messages[0].content).toBe('Hello, how are you?');
    });

    test('should add system message from instructions', () => {
      const request: ResponsesRequest = {
        model: 'gpt-4',
        input: 'Tell me a joke',
        instructions: 'You are a funny comedian.',
      };

      const result = transformRequest(request);

      expect(result.messages).toHaveLength(2);
      expect(result.messages[0].role).toBe('system');
      expect(result.messages[0].content).toBe('You are a funny comedian.');
      expect(result.messages[1].role).toBe('user');
      expect(result.messages[1].content).toBe('Tell me a joke');
    });

    test('should transform input array with message items', () => {
      const request: ResponsesRequest = {
        model: 'gpt-4',
        input: [
          {
            type: 'message',
            role: 'user',
            content: [
              { type: 'input_text', text: 'What is the capital of France?' },
            ],
          },
        ],
      };

      const result = transformRequest(request);

      expect(result.messages).toHaveLength(1);
      expect(result.messages[0].role).toBe('user');
      expect(result.messages[0].content).toBe('What is the capital of France?');
    });

    test('should transform tools from internal to external tagging', () => {
      const request: ResponsesRequest = {
        model: 'gpt-4',
        input: 'Get weather for Paris',
        tools: [
          {
            type: 'function',
            name: 'get_weather',
            description: 'Get weather for a location',
            parameters: {
              type: 'object',
              properties: {
                location: { type: 'string' },
              },
              required: ['location'],
            },
          },
        ],
      };

      const result = transformRequest(request);

      expect(result.tools).toBeDefined();
      expect(result.tools).toHaveLength(1);
      expect(result.tools![0].type).toBe('function');
      expect(result.tools![0].function.name).toBe('get_weather');
      expect(result.tools![0].function.description).toBe('Get weather for a location');
    });

    test('should map temperature parameter', () => {
      const request: ResponsesRequest = {
        model: 'gpt-4',
        input: 'Test',
        temperature: 0.7,
      };

      const result = transformRequest(request);

      expect(result.temperature).toBe(0.7);
    });

    test('should map max_output_tokens to max_completion_tokens', () => {
      const request: ResponsesRequest = {
        model: 'gpt-4',
        input: 'Test',
        max_output_tokens: 1000,
      };

      const result = transformRequest(request);

      expect(result.max_completion_tokens).toBe(1000);
    });

    test('should use custom model mapping when provided', () => {
      const request: ResponsesRequest = {
        model: 'custom-model',
        input: 'Test',
      };

      const result = transformRequest(request, {
        modelMapping: { 'custom-model': 'mapped-model' },
      });

      expect(result.model).toBe('mapped-model');
    });
  });
});

describe('Response Transformer', () => {
  describe('transformResponse', () => {
    test('should transform simple text response', () => {
      const response: ChatCompletionResponse = {
        id: 'chatcmpl-123',
        object: 'chat.completion',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            message: {
              role: 'assistant',
              content: 'Hello! I am doing well, thank you.',
            },
            finish_reason: 'stop',
          },
        ],
        usage: {
          prompt_tokens: 10,
          completion_tokens: 8,
          total_tokens: 18,
        },
      };

      const result = transformResponse(response);

      expect(result.id).toMatch(/^resp_/);
      expect(result.object).toBe('response');
      expect(result.status).toBe('completed');
      expect(result.output).toHaveLength(1);
      expect(result.output[0].type).toBe('message');
      expect(result.output[0].content[0].type).toBe('output_text');
      expect(result.output[0].content[0].text).toBe('Hello! I am doing well, thank you.');
      expect(result.usage?.input_tokens).toBe(10);
      expect(result.usage?.output_tokens).toBe(8);
      expect(result.usage?.total_tokens).toBe(18);
    });

    test('should map usage tokens correctly', () => {
      const response: ChatCompletionResponse = {
        id: 'chatcmpl-123',
        object: 'chat.completion',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            message: {
              role: 'assistant',
              content: 'Test',
            },
            finish_reason: 'stop',
          },
        ],
        usage: {
          prompt_tokens: 100,
          completion_tokens: 50,
          total_tokens: 150,
        },
      };

      const result = transformResponse(response);

      expect(result.usage?.input_tokens).toBe(100);
      expect(result.usage?.output_tokens).toBe(50);
      expect(result.usage?.total_tokens).toBe(150);
    });

    test('should handle tool calls in response', () => {
      const response: ChatCompletionResponse = {
        id: 'chatcmpl-123',
        object: 'chat.completion',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            message: {
              role: 'assistant',
              content: 'Let me check the weather.',
              tool_calls: [
                {
                  id: 'call_abc123',
                  type: 'function',
                  function: {
                    name: 'get_weather',
                    arguments: '{"location":"Paris"}',
                  },
                },
              ],
            },
            finish_reason: 'tool_calls',
          },
        ],
      };

      const result = transformResponse(response);

      expect(result.output[0].content).toHaveLength(2);
      expect(result.output[0].content[0].type).toBe('output_text');
      expect(result.output[0].content[1].type).toBe('function_call');
      expect(result.output[0].content[1].name).toBe('get_weather');
      expect(result.output[0].content[1].arguments).toBe('{"location":"Paris"}');
    });

    test('should map finish_reason to status', () => {
      const testCases = [
        { finish_reason: 'stop', expected_status: 'completed' },
        { finish_reason: 'length', expected_status: 'incomplete' },
        { finish_reason: 'tool_calls', expected_status: 'completed' },
        { finish_reason: 'content_filter', expected_status: 'failed' },
      ];

      for (const testCase of testCases) {
        const response: ChatCompletionResponse = {
          id: 'chatcmpl-123',
          object: 'chat.completion',
          created: 1234567890,
          model: 'gpt-4',
          choices: [
            {
              index: 0,
              message: {
                role: 'assistant',
                content: 'Test',
              },
              finish_reason: testCase.finish_reason as any,
            },
          ],
        };

        const result = transformResponse(response);
        expect(result.status).toBe(testCase.expected_status);
      }
    });
  });
});

describe('Stream Transformer', () => {
  describe('StreamTransformer', () => {
    test('should create response.created event on first chunk', () => {
      const transformer = new StreamTransformer();

      const chunk: ChatCompletionChunk = {
        id: 'chatcmpl-123',
        object: 'chat.completion.chunk',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: { role: 'assistant' },
            finish_reason: null,
          },
        ],
      };

      const events = transformer.transformChunk(chunk);

      expect(events).toHaveLength(1);
      expect(events[0].type).toBe('response.created');
      expect((events[0] as any).response_id).toMatch(/^resp_/);
    });

    test('should transform text delta events', () => {
      const transformer = new StreamTransformer();

      // First chunk to initialize
      const initChunk: ChatCompletionChunk = {
        id: 'chatcmpl-123',
        object: 'chat.completion.chunk',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: { role: 'assistant' },
            finish_reason: null,
          },
        ],
      };
      transformer.transformChunk(initChunk);

      // Text delta chunk
      const textChunk: ChatCompletionChunk = {
        id: 'chatcmpl-123',
        object: 'chat.completion.chunk',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: { content: 'Hello' },
            finish_reason: null,
          },
        ],
      };

      const events = transformer.transformChunk(textChunk);

      expect(events.length).toBeGreaterThan(0);
      const deltaEvent = events.find(e => e.type === 'response.delta');
      expect(deltaEvent).toBeDefined();
      expect((deltaEvent as any).delta.type).toBe('output_text');
      expect((deltaEvent as any).delta.text).toBe('Hello');
    });

    test('should create response.completed event on finish', () => {
      const transformer = new StreamTransformer();

      // Initialize
      const initChunk: ChatCompletionChunk = {
        id: 'chatcmpl-123',
        object: 'chat.completion.chunk',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: { role: 'assistant' },
            finish_reason: null,
          },
        ],
      };
      transformer.transformChunk(initChunk);

      // Final chunk
      const finalChunk: ChatCompletionChunk = {
        id: 'chatcmpl-123',
        object: 'chat.completion.chunk',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: {},
            finish_reason: 'stop',
          },
        ],
      };

      const events = transformer.transformChunk(finalChunk);

      expect(events.length).toBeGreaterThan(0);
      const completedEvent = events.find(e => e.type === 'response.completed');
      expect(completedEvent).toBeDefined();
    });

    test('should reset state between streams', () => {
      const transformer = new StreamTransformer();

      // First stream
      const chunk1: ChatCompletionChunk = {
        id: 'chatcmpl-123',
        object: 'chat.completion.chunk',
        created: 1234567890,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: { content: 'Hello' },
            finish_reason: null,
          },
        ],
      };
      transformer.transformChunk(chunk1);

      transformer.reset();

      // Second stream should start fresh
      const chunk2: ChatCompletionChunk = {
        id: 'chatcmpl-456',
        object: 'chat.completion.chunk',
        created: 1234567891,
        model: 'gpt-4',
        choices: [
          {
            index: 0,
            delta: { role: 'assistant' },
            finish_reason: null,
          },
        ],
      };
      const events = transformer.transformChunk(chunk2);

      expect(events[0].type).toBe('response.created');
    });
  });
});
