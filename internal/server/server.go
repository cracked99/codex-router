package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/plasmadev/codex-api-router/internal/config"
	"github.com/plasmadev/codex-api-router/internal/server/handlers"
	"github.com/plasmadev/codex-api-router/internal/server/middleware"
)

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	httpServer *http.Server
	listener   net.Listener
	logger     *slog.Logger
	shutdown   atomic.Bool
	wg         sync.WaitGroup
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	logger := newLogger(cfg.Logging)
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting codex-api-router",
		"version", "0.1.0",
		"host", s.cfg.Server.Host,
		"port", s.cfg.Server.Port,
		"backend", s.cfg.Zai.BaseURL,
		"translator_mode", s.cfg.Translator.Mode,
	)

	handler := s.createHandler()

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:           handler,
		ReadTimeout:       s.cfg.Zai.Timeout,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      s.cfg.Zai.Timeout,
		IdleTimeout:       120 * time.Second,
	}

	if s.cfg.Server.TLS.Enabled {
		s.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	var err error
	s.listener, err = net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.logger.Info("server listening", "addr", s.listener.Addr().String())

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if s.cfg.Server.TLS.Enabled {
			if err := s.httpServer.ServeTLS(s.listener, s.cfg.Server.TLS.CertFile, s.cfg.Server.TLS.KeyFile); err != nil && !s.shutdown.Load() {
				s.logger.Error("server error", "error", err)
			}
		} else {
			if err := s.httpServer.Serve(s.listener); err != nil && !s.shutdown.Load() {
				s.logger.Error("server error", "error", err)
			}
		}
	}()

	return s.waitForShutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	s.shutdown.Store(true)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shutdown http server", "error", err)
		return err
	}

	if s.listener != nil {
		s.listener.Close()
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("server shutdown complete")
		return nil
	case <-ctx.Done():
		s.logger.Error("shutdown timeout exceeded")
		return ctx.Err()
	}
}

func (s *Server) createHandler() http.Handler {
	mux := http.NewServeMux()

	proxyHandler := handlers.NewProxyHandler(s.cfg, s.logger)

	mux.HandleFunc("/v1/responses", proxyHandler.ServeHTTP)
	mux.HandleFunc("/v1/responses/", proxyHandler.ServeHTTP)
	mux.HandleFunc("/responses", proxyHandler.ServeHTTP)
	mux.HandleFunc("/responses/", proxyHandler.ServeHTTP)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	if s.cfg.Metrics.Enabled {
		mux.HandleFunc("/metrics", handlers.MetricsHandler(s.logger))
	}

	var handler http.Handler = mux
	handler = middleware.Recovery(handler, s.logger)
	handler = middleware.RequestLogging(handler, s.logger)
	handler = middleware.CORS(handler)

	return handler
}

func (s *Server) waitForShutdown() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	s.logger.Info("received signal, initiating shutdown", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}

func newLogger(cfg config.LoggingConfig) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Level),
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			if cfg.Format == "json" {
				handler = slog.NewJSONHandler(file, opts)
			} else {
				handler = slog.NewTextHandler(file, opts)
			}
		}
	}

	return slog.New(handler)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
