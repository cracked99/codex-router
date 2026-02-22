package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check router health status",
	Long: `Check the health status of a running codex-router instance.

This command queries the /health endpoint and displays the response.
Useful for monitoring and health checks in container orchestration.

Examples:
  # Check health of local router
  codex-router health

  # Check health of remote router
  codex-router health --url http://router.example.com:8080

  # JSON output
  codex-router health --output json

  # Wait for healthy state
  codex-router health --wait --timeout 30s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get URL
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			if host == "" {
				host = "localhost"
			}
			if port == 0 {
				port = 8080
			}
			url = fmt.Sprintf("http://%s:%d", host, port)
		}

		// Check if we should wait
		wait, _ := cmd.Flags().GetBool("wait")
		if wait {
			return waitForHealth(url, cmd)
		}

		// Single health check
		return checkHealth(url)
	},
}

// statusCmd shows detailed router status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show detailed router status",
	Long: `Show detailed status information about a running router.

Includes:
  • Server status
  • Backend connectivity
  • Configuration summary
  • Performance metrics

Examples:
  codex-router status
  codex-router status --url http://router.example.com:8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			if host == "" {
				host = "localhost"
			}
			if port == 0 {
				port = 8080
			}
			url = fmt.Sprintf("http://%s:%d", host, port)
		}

		return checkStatus(url)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(statusCmd)

	// Health flags
	healthCmd.Flags().String("url", "", "router URL (default: http://localhost:8080)")
	healthCmd.Flags().String("host", "", "router host (default: localhost)")
	healthCmd.Flags().Int("port", 0, "router port (default: 8080)")
	healthCmd.Flags().Bool("wait", false, "wait for healthy state")
	healthCmd.Flags().Duration("timeout", 30*time.Second, "timeout for wait mode")

	// Status flags
	statusCmd.Flags().String("url", "", "router URL (default: http://localhost:8080)")
	statusCmd.Flags().String("host", "", "router host (default: localhost)")
	statusCmd.Flags().Int("port", 0, "router port (default: 8080)")
}

func checkHealth(url string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(url + "/health")
	if err != nil {
		fmt.Printf("✗ Router not reachable: %v\n", err)
		return fmt.Errorf("health check failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("✗ Health check failed (status %d)\n", resp.StatusCode)
		return fmt.Errorf("unhealthy")
	}

	// Parse and display health response
	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		fmt.Printf("✓ Router is healthy (status %d)\n", resp.StatusCode)
		return nil
	}

	// Format output
	if globalOpts.Output == "json" {
		fmt.Println(string(body))
	} else {
		fmt.Println("✓ Router is healthy")
		if status, ok := health["status"]; ok {
			fmt.Printf("  Status: %v\n", status)
		}
		if version, ok := health["version"]; ok {
			fmt.Printf("  Version: %v\n", version)
		}
	}

	return nil
}

func checkStatus(url string) error {
	// Check health first
	fmt.Println("Checking router status...")
	fmt.Printf("URL: %s\n\n", url)

	// Health check
	if err := checkHealth(url); err != nil {
		return err
	}

	// Try to get metrics
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(url + "/metrics")
	if err != nil {
		fmt.Println("\n⚠ Metrics endpoint not available")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read metrics: %w", err)
	}

	fmt.Println("\nMetrics:")
	fmt.Println(string(body))

	return nil
}

func waitForHealth(url string, cmd *cobra.Command) error {
	timeout, _ := cmd.Flags().GetDuration("timeout")
	
	fmt.Printf("Waiting for router to become healthy (timeout: %v)...\n", timeout)

	ctx := time.After(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx:
			fmt.Println("\n✗ Timeout waiting for healthy state")
			return fmt.Errorf("timeout")
		case <-ticker.C:
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(url + "/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				fmt.Println("\n✓ Router is healthy")
				return nil
			}
			fmt.Print(".")
		}
	}
}
