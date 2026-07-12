# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | ✅                 |

## Security Considerations

### TLS Certificate Validation
- PingWatch validates TLS certificates by default when checking HTTPS endpoints
- `InsecureSkipVerify` is always set to `false` — certificate errors are reported as validation failures
- Expired, self-signed, or invalid certificates will cause the check to fail

### Safe HTTP Practices
- HTTP redirects are not followed by default (`http.ErrUseLastResponse`)
- Response bodies are truncated to 1KB to prevent memory exhaustion
- Request timeouts are enforced and configurable

### Credential Management
- No hardcoded credentials or API tokens in the codebase
- Headers (including authentication tokens) are loaded from config files only
- Never commit `.env` or config files with real credentials to version control

### SQLite Database
- The history database is a local file — protect it with appropriate file permissions
- WAL mode is used for safe concurrent access
- No network exposure — the database is purely local

### Input Validation
- All CLI arguments are validated before use
- Config files are parsed with strict YAML/JSON parsing — malformed files are rejected
- URL validation is performed via Go's standard library HTTP client

## Reporting a Vulnerability

If you discover a security vulnerability, please open a GitHub issue with the "security" label.
Do not disclose security vulnerabilities publicly until they have been addressed.