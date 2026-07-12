# PingWatch

Lightweight HTTP endpoint monitoring CLI. Check endpoints, validate responses, monitor SSL certificates, track uptime history, and analyze performance — all from the terminal or CI/CD pipelines.

## Features

- **HTTP endpoint checking** — GET, POST, PUT, DELETE with custom headers and body
- **Response validation** — status codes, body content (contains/not-contains), response time thresholds
- **SSL/TLS certificate monitoring** — check certificate validity, expiry, issuer, and days remaining
- **Configurable retries** — automatic retry with configurable delays
- **Historical data storage** — SQLite-backed history of all checks
- **Uptime reports** — uptime percentages over configurable periods (1h, 6h, 24h, 7d, 30d, 90d)
- **Performance percentiles** — P50/P95/P99 response time analysis
- **Multi-format output** — text, JSON, CSV (for both checks and reports)
- **Config file** — YAML or JSON configuration for batch monitoring
- **CI-friendly** — exit codes for automated pipelines
- **Fast & standalone** — single binary, no runtime dependencies

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

## Quick Start

### Check a single URL

```bash
pingwatch check https://example.com
```

Output shows response status, time, and SSL certificate info:
```
✓ https://example.com [200] 143ms | SSL: example.com (48 days)
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
# Initialize a sample config
pingwatch init

# Check endpoints from config
pingwatch check --config config.yaml
```

### Store results in history database

```bash
pingwatch check --config config.yaml --db history.db
```

## Uptime & Performance Reports

Generate comprehensive reports from historical check data, including uptime percentages and response time percentiles (P50/P95/P99).

### Basic report (last 24 hours)

```bash
pingwatch report --db history.db
```

### Report over different periods

```bash
# Last 7 days
pingwatch report --db history.db --period 7d

# Last 30 days
pingwatch report --db history.db --period 30d

# Custom date range
pingwatch report --db history.db --since "2026-07-01T00:00:00Z"
```

### Report output formats

```bash
# JSON (for machine processing)
pingwatch report --db history.db --format json

# CSV (for spreadsheets)
pingwatch report --db history.db --format csv
```

### Sample report output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Example
  URL: https://example.com
  Status: ● UP
────────────────────────────────────────────────────────────
  Period:             2026-07-11 → 2026-07-12
  Total checks:       1440
  Successful:         1438
  Failed:             2
  Uptime:             99.86%

  Response Times:
    Average:          152.3 ms
    Min:              98.0 ms
    Max:              2341.0 ms
    P50 (median):     148.0 ms
    P95:              195.0 ms
    P99:              312.0 ms
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
  report    Generate uptime and performance reports
  help      Help about any command

check flags:
  -c, --config string       Config file path (YAML or JSON)
  -f, --format string       Output format: text, json, csv
  -u, --url string          Single URL to check
  -m, --method string       HTTP method (GET, POST, PUT, DELETE)
  -t, --timeout int         Timeout in milliseconds
  -s, --expected-status int Expected HTTP status code
  -b, --body string         Request body (for POST/PUT)
      --contains string     Response body must contain this string
  -d, --db string           Save results to SQLite history database

report flags:
  -d, --db string           Path to history database (required)
  -f, --format string       Output format: text, json, csv (default "text")
  -p, --period string       Report period: 1h, 6h, 12h, 24h, 7d, 30d, 90d (default "24h")
  -s, --since string        Explicit start date (RFC3339 format, overrides --period)
```

## CI/CD Integration

### GitHub Actions with history

```yaml
- name: Check endpoints
  run: |
    go install github.com/EdgarOrtegaRamirez/pingwatch@latest
    pingwatch check --config pingwatch.yaml --db history.db --format json

- name: Generate report
  run: |
    pingwatch report --db history.db --format json > report.json
```

### Makefile

```makefile
check:
	pingwatch check --format json --config config.yaml --db history.db

report:
	pingwatch report --db history.db --period 7d
```

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   cmd/      │────▶│  monitor/   │────▶│  output/    │
│  (Cobra CLI)│     │ (HTTP client)│     │(text/json/csv)│
└──────┬──────┘     └──────┬──────┘     └─────────────┘
       │                   │
       ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ config/     │     │   ssl/     │     │  history/   │
│ (YAML/JSON) │     │(TLS certs) │     │  (SQLite)   │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Security

- No hardcoded credentials or secrets
- Headers and body content are passed via config, never logged
- TLS certificate validation is enabled by default (no `InsecureSkipVerify`)
- Follows safe HTTP practices (no redirect following by default)
- Response body truncation (1KB) for large responses
- SQLite database uses WAL mode with safe defaults

## License

MIT License — see [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request