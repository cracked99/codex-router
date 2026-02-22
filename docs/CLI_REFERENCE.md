# Codex Router CLI Reference

Complete command-line interface reference for codex-router.

## Installation

```bash
# Build from source
go build -o codex-router ./cmd/codex-router

# Or install to $GOPATH/bin
go install ./cmd/codex-router
```

## Global Flags

These flags are available for all commands:

```
  -c, --config string      Config file (default is $HOME/.codex-router/config.yaml)
  -v, --verbose            Verbose output
      --debug              Debug mode (very verbose)
  -o, --output string      Output format (text, json, yaml) (default "text")
      --no-color           Disable colored output
```

## Commands

### Root Command

```bash
codex-router [flags]
```

Display help and version information.

### serve - Start the API Router

```bash
codex-router serve [flags]
```

Start the HTTP server that listens for Responses API requests and proxies them to z.ai's Chat Completions API.

**Flags:**
```
  -H, --host string              Host to bind to (overrides config)
  -p, --port int                 Port to listen on (overrides config)
  -k, --api-key string           z.ai API key (overrides config)
  -b, --backend-url string       Backend URL for z.ai API (overrides config)
      --timeout duration         Request timeout (e.g., 120s)
      --translator-mode string   Translator mode (wasm or sidecar)
  -D, --dev                      Enable development mode (sidecar translator, debug logging)
      --tls                      Enable TLS
      --tls-cert string          TLS certificate file
      --tls-key string           TLS private key file
  -n, --dry-run                  Validate configuration without starting server
```

**Examples:**
```bash
# Start with default configuration
codex-router serve

# Start with custom port
codex-router serve --port 9090

# Start with API key
codex-router serve --api-key sk-xxx

# Start in development mode
codex-router serve --dev

# Start with custom backend
codex-router serve --backend-url https://custom.api.com

# Dry run to validate config
codex-router serve --dry-run
```

### config - Configuration Management

```bash
codex-router config [command]
```

Manage configuration files and settings.

#### config init - Initialize Configuration

```bash
codex-router config init [path] [flags]
```

Create a new configuration file with default values.

**Flags:**
```
      --force          Overwrite existing config file
  -i, --interactive    Interactive configuration
```

**Examples:**
```bash
# Create default config
codex-router config init

# Create config at specific location
codex-router config init ./my-config.yaml

# Interactive configuration
codex-router config init --interactive

# Overwrite existing config
codex-router config init --force
```

#### config show - Display Configuration

```bash
codex-router config show [flags]
```

Display the effective configuration from all sources.

**Flags:**
```
  -f, --format string   Output format (yaml, json) (default "yaml")
```

**Examples:**
```bash
# Show configuration
codex-router config show

# Show as JSON
codex-router config show --format json
```

#### config validate - Validate Configuration

```bash
codex-router config validate [path] [flags]
```

Validate a configuration file for correctness.

**Flags:**
```
      --strict   Enable strict security validation
```

**Examples:**
```bash
# Validate default config
codex-router config validate

# Validate specific config file
codex-router config validate ./my-config.yaml

# Strict validation with security checks
codex-router config validate --strict
```

#### config edit - Edit Configuration

```bash
codex-router config edit
```

Open the configuration file in your default editor ($EDITOR or $VISUAL).

**Examples:**
```bash
codex-router config edit
```

#### config set - Set Configuration Value

```bash
codex-router config set <key> <value>
```

Set an individual configuration value in the config file.

**Examples:**
```bash
codex-router config set server.port 9090
codex-router config set zai.api_key sk-xxx
codex-router config set logging.level debug
```

#### config get - Get Configuration Value

```bash
codex-router config get <key>
```

Get an individual configuration value.

**Examples:**
```bash
codex-router config get server.port
codex-router config get zai.base_url
```

### health - Health Checks

```bash
codex-router health [flags]
```

Check the health status of a running codex-router instance.

**Flags:**
```
      --url string          Router URL (default: http://localhost:8080)
      --host string         Router host (default: localhost)
      --port int            Router port (default: 8080)
      --wait                Wait for healthy state
      --timeout duration    Timeout for wait mode (default 30s)
```

**Examples:**
```bash
# Check health of local router
codex-router health

# Check health of remote router
codex-router health --url http://router.example.com:8080

# Wait for healthy state
codex-router health --wait --timeout 30s

# JSON output
codex-router health --output json
```

### status - Detailed Status

```bash
codex-router status [flags]
```

Show detailed status information about a running router.

**Flags:**
```
      --url string       Router URL (default: http://localhost:8080)
      --host string      Router host (default: localhost)
      --port int         Router port (default: 8080)
```

**Examples:**
```bash
codex-router status
codex-router status --url http://router.example.com:8080
```

### version - Version Information

```bash
codex-router version [flags]
```

Display detailed version and build information.

**Examples:**
```bash
# Text output
codex-router version

# JSON output
codex-router version --output json
```

### proxy - Proxy Management

```bash
codex-router proxy [command]
```

Commands for testing and managing the proxy functionality.

#### proxy test - Test Transformation

```bash
codex-router proxy test [request-file] [flags]
```

Test how a request would be transformed without actually sending it.

**Examples:**
```bash
# Test with file
codex-router proxy test request.json

# Test with stdin
echo '{"model":"gpt-4","input":"hello"}' | codex-router proxy test
```

#### proxy validate - Validate Request

```bash
codex-router proxy validate [request-file]
```

Validate a Responses API request for correctness.

**Examples:**
```bash
# Validate file
codex-router proxy validate request.json

# Validate stdin
echo '{"model":"gpt-4","input":"test"}' | codex-router proxy validate
```

#### proxy call - Make Request

```bash
codex-router proxy call [request-file] [flags]
```

Send a request through the router to the backend.

**Flags:**
```
      --url string   Router URL (default: http://localhost:8080)
```

**Examples:**
```bash
# Make request with file
codex-router proxy call request.json

# Make request with stdin
echo '{"model":"gpt-4","input":"hello"}' | codex-router proxy call
```

## Configuration Priority

Configuration is loaded in the following priority order (highest to lowest):

1. **Command-line flags** - Override everything
2. **Environment variables** - `CODEX_ROUTER_*` prefix
3. **Config file** - YAML configuration
4. **Default values** - Built-in defaults

### Environment Variables

All configuration values can be set via environment variables with the `CODEX_ROUTER_` prefix:

```bash
# Server configuration
export CODEX_ROUTER_SERVER_HOST=localhost
export CODEX_ROUTER_SERVER_PORT=8080

# Z.ai configuration
export CODEX_ROUTER_ZAI_API_KEY=sk-xxx
export CODEX_ROUTER_ZAI_BASE_URL=https://api.z.ai/api/paas/v4

# Or use the shorthand
export ZAI_API_KEY=sk-xxx
```

## Configuration File

Default location: `~/.codex-router/config.yaml`

```yaml
# Server configuration
server:
  host: "localhost"
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

# Z.ai backend configuration
zai:
  base_url: "https://api.z.ai/api/paas/v4"
  api_key: "${ZAI_API_KEY}"  # Environment variable expansion
  timeout: 120s
  max_retries: 3
  retry_delay: 1s

# Codex configuration
codex:
  base_url: ""
  api_key_header: "Authorization"

# Translator configuration
translator:
  mode: "wasm"  # wasm | sidecar
  wasm_path: "./translator.wasm"
  sidecar_command: "node ./internal/translator/dist/index.js"

# Session management
session:
  enabled: true
  ttl: 3600s
  max_conversations: 1000

# Logging configuration
logging:
  level: "info"  # debug | info | warn | error
  format: "json"  # json | text
  file: ""  # Optional: log to file

# Metrics configuration
metrics:
  enabled: true
  path: "/metrics"
  format: "prometheus"
```

## Examples

### Quick Start

```bash
# 1. Initialize configuration
codex-router config init

# 2. Set API key
codex-router config set zai.api_key sk-xxx

# 3. Validate configuration
codex-router config validate

# 4. Start server
codex-router serve
```

### Development Workflow

```bash
# Start in development mode
codex-router serve --dev

# In another terminal, test a request
echo '{"model":"gpt-4","input":"hello"}' | codex-router proxy test

# Check health
codex-router health

# View logs
tail -f /var/log/codex-router.log
```

### Production Deployment

```bash
# Validate configuration with strict checks
codex-router config validate --strict

# Start with production settings
codex-router serve \
  --host 0.0.0.0 \
  --port 8080 \
  --tls \
  --tls-cert /etc/ssl/cert.pem \
  --tls-key /etc/ssl/key.pem

# Monitor health
codex-router health --url https://router.example.com
```

### Docker Deployment

```bash
# Build image
docker build -t codex-router .

# Run with environment variables
docker run -d \
  -p 8080:8080 \
  -e ZAI_API_KEY=sk-xxx \
  -e CODEX_ROUTER_LOGGING_LEVEL=info \
  codex-router

# Health check
docker exec <container> codex-router health
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: codex-router
spec:
  template:
    spec:
      containers:
      - name: router
        image: codex-router:latest
        ports:
        - containerPort: 8080
        env:
        - name: ZAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: codex-secrets
              key: api-key
        livenessProbe:
          exec:
            command: ["codex-router", "health"]
          initialDelaySeconds: 5
          periodSeconds: 10
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Validation error
- `130` - Interrupted (Ctrl+C)

## Troubleshooting

### Configuration Issues

```bash
# Show effective configuration
codex-router config show

# Validate configuration
codex-router config validate --strict

# Test configuration without starting server
codex-router serve --dry-run
```

### Connection Issues

```bash
# Check if router is running
codex-router health

# Check detailed status
codex-router status

# Test a request transformation
echo '{"model":"gpt-4","input":"test"}' | codex-router proxy test
```

### Debug Mode

```bash
# Enable debug logging
codex-router serve --debug

# Or set in config
codex-router config set logging.level debug
```

## See Also

- [Configuration Guide](./CONFIGURATION.md)
- [API Reference](./API_REFERENCE.md)
- [Deployment Guide](./DEPLOYMENT.md)
