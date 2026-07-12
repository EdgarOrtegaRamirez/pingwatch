# PingWatch

Lightweight HTTP endpoint monitoring CLI. Check endpoints, validate responses, and monitor uptime — all from the terminal or CI/CD pipelines.

## Features

- **Multi-format output** — text, JSON, CSV
- **Response validation** — status codes, body content, response time thresholds
- **Configurable retries** — automatic retry with configurable delays
- **HTTP methods** — GET, POST, PUT, DELETE with custom headers and body
- **Config file** — YAML or JSON configuration for batch monitoring
- **CI-friendly** — exit codes for automated pipelines
- **Fast & standalone** — no dependencies beyond the CLI binary

## Installation

### Go Install

```bash
go install github.com/EdgarOrtegaRamirez/pingwatch@latest
```

### From Source

```bash
git clone https://github.com/EdgarOrtegaRamirez/pingwatch.git
cd pingwatch
go build -o pingwatch
sudo mv pingwatch /usr/local/bin/
```

### Homebrew (future)

```bash
# tap coming soon
brew install EdgarOrtegaRamirez/tap/pingwatch
```

## Quick Start

### Check a single URL

```bash
pingwatch check https://example.com
```

### Check multiple URLs

```bash
pingwatch check https://example.com https://httpbin.org/get
```

### Check with custom method and body

```bash
pingwatch check \
  --url https://httpbin.org/post \
  --method POST \
  --body '{"test": true}' \
  --expected-status 201
```

### Validate response content

```bash
pingwatch check \
  --url https://httpbin.org/get \
  --contains '"url": "https://httpbin.org/get"'
```

### Use a config file

```bash
pingwatch check --config config.yaml
```

### Output as JSON (for CI)

```bash
pingwatch check --format json https://example.com > results.json
```

## Configuration

### Initialize a config file

```bash
pingwatch init
# Creates config.yaml with sample endpoints
```

### Config file format (YAML)

```yaml
defaults:
  interval: 60s
  timeout: 10s
  expected_status: 200
  method: GET
  retries: 1
  retry_delay: 1s

endpoints:
  - url: https://example.com
    name: Example
    expected_status: 200
    max_response_time_ms: 5000

  - url: https://httpbin.org/post
    name: HTTPBin POST
    method: POST
    body: '{"key": "value"}'
    expected_status: 200
    response_body_contains: '"key": "value"'
    headers:
      X-Api-Key: secret123
```

### Config file format (JSON)

```json
{
  "defaults": {
    "timeout": "10s",
    "expected_status": 200
  },
  "endpoints": [
    {
      "url": "https://example.com",
      "name": "Example"
    }
  ]
}
```

## CLI Reference

```
pingwatch [command]

Commands:
  check     Check HTTP endpoints
  init      Initialize a new config file
  show      Show parsed config from a file
  help      Help about any command

check flags:
  -c, --config string       Config file path
  -f, --format string       Output format: text, json, csv
  -u, --url string          Single URL to check
  -m, --method string       HTTP method (GET, POST, PUT, DELETE)
  -t, --timeout int         Timeout in milliseconds
  -s, --expected-status int Expected HTTP status code
  -b, --body string         Request body (for POST/PUT)
      --contains string     Response body must contain this string
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Check endpoints
  run: |
    go install github.com/EdgarOrtegaRamirez/pingwatch@latest
    pingwatch check --config pingwatch.yaml --format json > results.json
```

### Makefile

```makefile
check:
	pingwatch check --format json --config config.yaml
```

### Docker

```bash
docker run --rm \
  -v $(pwd):/config \
  ghcr.io/edgarortegaramirez/pingwatch \
  check --config /config/pingwatch.yaml
```

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   cmd/      │────▶│  monitor/   │────▶│  output/    │
│  (Cobra CLI)│     │ (HTTP client)│     │(text/json/csv)│
└─────────────┘     └─────────────┘     └─────────────┘
         │                  │
         ▼                  ▼
   ┌─────────────┐     ┌─────────────┐
   │ config/     │     │   config/   │
   │ (YAML/JSON) │     │  (structs)  │
   └─────────────┘     └─────────────┘
```

## Security

- No hardcoded credentials or secrets
- Headers and body content are passed via config, never logged
- Follows safe HTTP practices (no redirect following by default)
- Response body truncation (1KB) for large responses

## License

MIT License — see [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request
