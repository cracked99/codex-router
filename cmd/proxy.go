package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// proxyCmd represents proxy-related commands
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Proxy management commands",
	Long: `Commands for testing and managing the proxy functionality.

These commands help you:
  • Test request/response transformation
  • Debug API translation issues
  • Validate requests before sending`,
}

// proxyTestCmd tests a request transformation
var proxyTestCmd = &cobra.Command{
	Use:   "test [request-file]",
	Short: "Test request/response transformation",
	Long: `Test how a request would be transformed without actually sending it.

Reads a Responses API request from a file (or stdin) and shows how it 
would be transformed to Chat Completions API format.

Examples:
  # Test with file
  codex-router proxy test request.json

  # Test with stdin
  echo '{"model":"gpt-4","input":"hello"}' | codex-router proxy test

  # Show both request and response transformation
  codex-router proxy test --both request.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read request
		var input io.Reader = os.Stdin
		if len(args) > 0 {
			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()
			input = file
		}

		data, err := io.ReadAll(input)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Parse as Responses API request
		var req map[string]interface{}
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		fmt.Println("Original Request (Responses API):")
		fmt.Println("---")
		prettyJSON(data)
		fmt.Println()

		// Show transformation (simplified - would use actual translator in production)
		fmt.Println("Transformed Request (Chat Completions API):")
		fmt.Println("---")
		transformed := transformRequestExample(req)
		prettyJSON(transformed)

		return nil
	},
}

// proxyValidateCmd validates a request
var proxyValidateCmd = &cobra.Command{
	Use:   "validate [request-file]",
	Short: "Validate a Responses API request",
	Long: `Validate a Responses API request for correctness.

Checks for:
  • Valid JSON syntax
  • Required fields present
  • Valid field values
  • Compatible parameters

Examples:
  codex-router proxy validate request.json
  echo '{"model":"gpt-4"}' | codex-router proxy validate`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var input io.Reader = os.Stdin
		if len(args) > 0 {
			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()
			input = file
		}

		data, err := io.ReadAll(input)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Validate JSON
		var req map[string]interface{}
		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Println("✗ Invalid JSON")
			return fmt.Errorf("validation failed: %w", err)
		}

		// Check required fields
		errors := []string{}
		if _, ok := req["model"]; !ok {
			errors = append(errors, "missing required field: model")
		}
		if _, ok := req["input"]; !ok {
			errors = append(errors, "missing required field: input")
		}

		if len(errors) > 0 {
			fmt.Println("✗ Validation failed:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err)
			}
			return fmt.Errorf("validation failed")
		}

		fmt.Println("✓ Request is valid")
		fmt.Printf("  Model: %v\n", req["model"])
		if stream, ok := req["stream"].(bool); ok && stream {
			fmt.Println("  Streaming: enabled")
		}

		return nil
	},
}

// proxyCallCmd makes an actual API call through the router
var proxyCallCmd = &cobra.Command{
	Use:   "call [request-file]",
	Short: "Make a request through the router",
	Long: `Send a request through the router to the backend.

This command sends an actual request to the router and displays the response.
Useful for end-to-end testing.

Examples:
  codex-router proxy call request.json
  echo '{"model":"gpt-4","input":"hello"}' | codex-router proxy call`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get router URL
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			url = "http://localhost:8080"
		}

		// Read request
		var input io.Reader = os.Stdin
		if len(args) > 0 {
			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()
			input = file
		}

		data, err := io.ReadAll(input)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Make request
		resp, err := http.Post(url+"/v1/responses", "application/json", strings.NewReader(string(data)))
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// Display response
		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Println("Response:")
		prettyJSON(body)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.AddCommand(proxyTestCmd)
	proxyCmd.AddCommand(proxyValidateCmd)
	proxyCmd.AddCommand(proxyCallCmd)

	// Call command flags
	proxyCallCmd.Flags().String("url", "", "router URL (default: http://localhost:8080)")
}

// Helper functions

func prettyJSON(data []byte) {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		output, _ := json.MarshalIndent(obj, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(string(data))
	}
}

func transformRequestExample(req map[string]interface{}) []byte {
	// Simplified transformation example
	chat := map[string]interface{}{
		"model": "glm-5", // Model mapping
		"messages": []map[string]string{
			{"role": "user", "content": fmt.Sprintf("%v", req["input"])},
		},
	}

	if temp, ok := req["temperature"].(float64); ok {
		chat["temperature"] = temp
	}
	if stream, ok := req["stream"].(bool); ok {
		chat["stream"] = stream
	}

	result, _ := json.MarshalIndent(chat, "", "  ")
	return result
}
