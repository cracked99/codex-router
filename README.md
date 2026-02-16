# Codex API Router

A proxy router that translates between Codex CLI's Responses API and z.ai's Chat Completions API.

## Features

- Transparent proxy translation between Responses API and Chat Completions API
- Single binary distribution
- Configuration file support with environment variable overrides
- Request/response logging for debugging
- Graceful shutdown handling
- Health check and metrics endpoints
- CORS support

## Installation

### From Source

```bash
go build -o codex-router ./cmd/codex-router
```

### Using Go Install

```bash
go install github.com/plasmadev/codex-api-router/cmd/codex-router@latest
```

## Configuration

### Initialize Configuration

```bash
codex-router config init
```

This creates a default configuration file at `~/.codex-router/config.yaml`.

### Configuration File

```yaml
server:
  host: "localhost"
  port: 8080

zai:
  base_url: "https://api.z.ai/api/paas/v4"
  api_key: "${ZAI_API_KEY}"  # Set via environment variable
  timeout: 120s
  max_retries: 3

translator:
  mode: "wasm"  # or "sidecar" for development

logging:
  level: "info"
  format: "json"
```

### Environment Variables

- `ZAI_API_KEY`: Your z.ai API key
- `CODEX_ROUTER_API_KEY`: Alternative API key variable

## Usage

### Start the Server

```bash
# Using default configuration
codex-router serve

# With custom port
codex-router serve --port 3000

# With API key override
codex-router serve --api-key your-api-key

# With custom backend URL
codex-router serve --backend-url https://custom-backend.com
```

### Command-Line Options

```
Global Flags:
      --config string       config file (default is $HOME/.codex-router/config.yaml)
  -l, --log-level string    log level (debug, info, warn, error) (default "info")
  -v, --verbose            verbose output

Serve Flags:
  -p, --port int            port to listen on (overrides config)
  -H, --host string         host to bind to (overrides config)
  -k, --api-key string      z.ai API key (overrides config)
  -b, --backend-url string  backend URL for z.ai API (overrides config)
      --dev-mode            enable development mode (sidecar translator)
```

### Configuration Commands

```bash
# Validate configuration
codex-router config validate

# Show version
codex-router version
```

## API Endpoints

### Proxy Endpoints

- `POST /v1/responses` - Create a response (proxy to z.ai)
- `GET /v1/responses/{id}` - Retrieve a response
- `DELETE /v1/responses/{id}` - Delete a response

### Monitoring Endpoints

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

## Development

### Project Structure

```
codex-api-router/
├── cmd/
│   └── codex-router/       # CLI application
├── internal/
│   ├── config/             # Configuration management
│   └── server/             # HTTP server
│       ├── handlers/       # Request handlers
│       └── middleware/     # HTTP middleware
├── pkg/                    # Public packages
└── translator/             # TypeScript translator
```

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Linter

```bash
make lint
```

## License

MIT License - see LICENSE file for details
