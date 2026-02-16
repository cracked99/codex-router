package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/plasmadev/codex-api-router/internal/config"
	"github.com/plasmadev/codex-api-router/internal/server/handlers"
	"github.com/plasmadev/codex-api-router/internal/server/middleware"
)

// Config holds the server configuration for the simple Start function
type Config struct {
	Host   string
	Port   int
	ZaiKey string
	ZaiURL string
}

// Start starts the HTTP server with the given configuration
func Start(cfg *Config) error {
	// Create logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting codex-api-router",
		"host", cfg.Host,
		"port", cfg.Port,
		"zai_url", cfg.ZaiURL,
	)

	// Create full config for handlers
	fullCfg := &config.Config{
		Zai: config.ZaiConfig{
			BaseURL: cfg.ZaiURL,
			APIKey:  cfg.ZaiKey,
			Timeout: 120 * time.Second,
		},
	}

	// Create handler
	proxyHandler := handlers.NewProxyHandler(fullCfg, logger)

	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/responses", proxyHandler.ServeHTTP)
	mux.HandleFunc("/v1/responses/", proxyHandler.ServeHTTP)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply middleware
	var h http.Handler = mux
	h = middleware.Recovery(h, logger)
	h = middleware.RequestLogging(h, logger)
	h = middleware.CORS(h)

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
		return err
	}

	logger.Info("server stopped")
	return nil
}
