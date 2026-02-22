# Quick Start Guide

Get up and running with codex-router in 5 minutes.

## Prerequisites

- Go 1.23 or later
- z.ai API key

## Installation

### Option 1: Build from Source

```bash
git clone https://github.com/plasmadev/codex-api-router.git
cd codex-api-router
make build
sudo make install
```

### Option 2: Go Install

```bash
go install github.com/plasmadev/codex-api-router/cmd/codex-router@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
codex-router config init
```

This creates `~/.codex-router/config.yaml` with default settings.

### 2. Set Your API Key

Option A: Environment variable (recommended)
```bash
export ZAI_API_KEY=your-api-key-here
```

Option B: Configuration file
```bash
codex-router config set zai.api_key your-api-key-here
```

### 3. Validate Configuration

```bash
codex-router config validate
```

Expected output:
```
✓ Configuration is valid
  Server: localhost:8080
  Backend: https://api.z.ai/api/paas/v4
  Translator: wasm
```

### 4. Start the Server

```bash
codex-router serve
```

Expected output:
```
╔═══════════════════════════════════════════════════════════╗
║          Codex API Router - Production Ready              ║
╚═══════════════════════════════════════════════════════════╝

  Version:     0.1.0 (commit: abc123, built: 2025-02-16)
  Server:      http://localhost:8080
  Backend:     https://api.z.ai/api/paas/v4
  Translator:  wasm
  Log Level:   info

  Endpoints:
    Proxy:    POST http://localhost:8080/v1/responses
    Health:   GET  http://localhost:8080/health
    Metrics:  GET  http://localhost:8080/metrics

  Press Ctrl+C to shutdown gracefully
```

### 5. Test the Router

In another terminal, make a test request:

```bash
curl -X POST http://localhost:8080/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "input": "Hello, how are you?"
  }'
```

Or use the built-in test command:

```bash
echo '{"model":"gpt-4","input":"hello"}' | codex-router proxy call
```

### 6. Check Health

```bash
codex-router health
```

Expected output:
```
✓ Router is healthy
  Status: ok
  Version: 0.1.0
```

## Common Use Cases

### Development Mode

Start with debug logging and sidecar translator:

```bash
codex-router serve --dev
```

### Custom Port

```bash
codex-router serve --port 9090
```

### Custom Backend

```bash
codex-router serve --backend-url https://custom.api.com/v4
```

### TLS Enabled

```bash
codex-router serve \
  --tls \
  --tls-cert /etc/ssl/cert.pem \
  --tls-key /etc/ssl/key.pem
```

## Next Steps

- Read the [CLI Reference](./CLI_REFERENCE.md) for all available commands
- Configure [advanced settings](./CONFIGURATION.md)
- Learn about [deployment options](./DEPLOYMENT.md)
- Explore the [API Reference](./API_REFERENCE.md)

## Troubleshooting

### Port Already in Use

```bash
# Use a different port
codex-router serve --port 9090
```

### Missing API Key

```bash
# Set API key
export ZAI_API_KEY=your-key-here
# Or in config
codex-router config set zai.api_key your-key-here
```

### Connection Refused

```bash
# Check if router is running
codex-router health

# Check logs
codex-router serve --debug
```

### Configuration Errors

```bash
# Validate configuration
codex-router config validate --strict

# Show effective configuration
codex-router config show
```

## Configuration Examples

### Minimal Configuration

```yaml
# ~/.codex-router/config.yaml
zai:
  api_key: "${ZAI_API_KEY}"
```

### Full Configuration

```yaml
# ~/.codex-router/config.yaml
server:
  host: "localhost"
  port: 8080
  tls:
    enabled: false

zai:
  base_url: "https://api.z.ai/api/paas/v4"
  api_key: "${ZAI_API_KEY}"
  timeout: 120s
  max_retries: 3

translator:
  mode: "wasm"

logging:
  level: "info"
  format: "json"

metrics:
  enabled: true
  path: "/metrics"
```

## Environment Variables

All settings can be configured via environment variables:

```bash
export CODEX_ROUTER_SERVER_HOST=localhost
export CODEX_ROUTER_SERVER_PORT=8080
export CODEX_ROUTER_ZAI_API_KEY=your-key-here
export CODEX_ROUTER_LOGGING_LEVEL=info
```

## Docker Quick Start

```bash
# Build
docker build -t codex-router .

# Run
docker run -d \
  -p 8080:8080 \
  -e ZAI_API_KEY=your-key-here \
  codex-router

# Test
curl http://localhost:8080/health
```

## Getting Help

- CLI help: `codex-router --help`
- Command help: `codex-router serve --help`
- Issues: https://github.com/plasmadev/codex-api-router/issues
- Documentation: https://github.com/plasmadev/codex-api-router/tree/main/docs
