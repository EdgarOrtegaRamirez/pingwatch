# AGENTS.md

## Project: PingWatch

A lightweight HTTP endpoint monitoring CLI built in Go.

## Architecture

```
pingwatch/
├── main.go              # Entry point
├── cmd/                 # CLI commands (Cobra)
│   ├── root.go          # Root command, version info
│   ├── check.go         # Main check command
│   ├── init.go          # Config initialization
│   └── show.go          # Config display
├── config/              # Configuration parsing
│   └── config.go        # YAML/JSON config loader
├── monitor/             # HTTP checking logic
│   └── monitor.go       # HTTP client + validation
├── output/              # Output formatting
│   └── output.go        # text/json/csv formatters
├── config.yaml          # Sample config (gitignored)
├── go.mod
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
go test ./... -v -race
```

## Key Dependencies

- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- Standard library `net/http` — HTTP client

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
