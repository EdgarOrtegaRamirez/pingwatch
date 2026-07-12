# AGENTS.md

## Project: PingWatch

A lightweight HTTP endpoint monitoring CLI built in Go with historical data storage, SSL monitoring, and uptime reporting.

## Architecture

```
pingwatch/
├── main.go              # Entry point
├── cmd/                 # CLI commands (Cobra)
│   ├── root.go          # Root command, version info
│   ├── check.go         # Main check command (with --db flag for history)
│   ├── init.go          # Config initialization
│   ├── show.go          # Config display
│   └── report.go        # Uptime/performance report command (NEW)
├── config/              # Configuration parsing
│   └── config.go        # YAML/JSON config loader
├── monitor/             # HTTP checking logic
│   ├── monitor.go       # HTTP client + validation + SSL checking
│   └── monitor_test.go
├── ssl/                 # SSL/TLS certificate checking (NEW)
│   ├── ssl.go           # Certificate retrieval and validation
│   └── ssl_test.go
├── history/             # SQLite historical data storage (NEW)
│   ├── history.go       # Store, report, percentile calculation
│   └── history_test.go
├── output/              # Output formatting
│   ├── output.go        # text/json/csv formatters (with SSL fields)
│   └── output_test.go
├── config.yaml          # Sample config (gitignored)
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

## Building

```bash
cd /root/workspace/pingwatch
go mod tidy
go build -o pingwatch
./pingwatch --help
```

## Testing

```bash
go test -vet=off ./... -v -count=1
```

## Key Dependencies

- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- `modernc.org/sqlite` — Pure-Go SQLite (no CGO needed)
- Standard library `crypto/tls` — TLS certificate inspection
- Standard library `net/http` — HTTP client

## New Features (merged from urlcheck)

### SSL Certificate Monitoring
- Automatically checks TLS certificates on HTTPS URLs
- Reports: subject, issuer, days remaining, validity status
- Expired or invalid certificates count as validation errors
- Uses standard library `crypto/tls` for certificate inspection

### SQLite Historical Storage
- Results stored in SQLite database via `--db` flag on `check` command
- WAL mode for concurrent safety and performance
- Batch inserts in transactions for efficiency
- Stores: timestamp, endpoint name, URL, status code, response time, success/failure, SSL info

### Uptime Reports
- New `report` command queries history database
- Calculates uptime percentage (successful / total * 100)
- Shows current status (UP/DOWN)
- Period options: 1h, 6h, 12h, 24h, 7d, 30d, 90d
- Custom date range via `--since` flag
- Output formats: text, JSON, CSV

### Response Time Percentiles
- P50 (median), P95, P99 calculated from historical data
- Also shows average, min, max response times
- Uses linear interpolation between sorted values
- All values rounded to 2 decimal places

## Adding a New Command

1. Create a new file in `cmd/` (e.g., `cmd/watch.go`)
2. Define a `*cobra.Command` with `Use`, `Short`, `Long`, `RunE`
3. Add it to `rootCmd` in `init()` of `cmd/root.go`
4. Add tests in the same file or a separate `_test.go`

## Config Structure

Endpoints are defined in YAML or JSON with:
- `url` (required) — endpoint URL
- `name` (optional) — display name, defaults to URL
- `method` (optional) — HTTP method, defaults to GET
- `expected_status` (optional) — expected status code, defaults to 200
- `timeout` (optional) — request timeout
- `headers` (optional) — map of header key/value pairs
- `body` (optional) — request body for POST/PUT
- `response_body_contains` (optional) — validate body contains string
- `response_body_not_contains` (optional) — validate body does NOT contain string
- `max_response_time_ms` (optional) — max acceptable response time in ms
- `retries` (optional) — number of retries on failure
- `retry_delay` (optional) — delay between retries