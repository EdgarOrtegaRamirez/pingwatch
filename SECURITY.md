# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## Reporting a Vulnerability

Please report security vulnerabilities to the repository owner via GitHub Issues with the "security" label. Do not create public issues for security concerns.

## Security Considerations

PingWatch handles URLs, headers, and request bodies. The following security measures are in place:

1. **No automatic redirect following** — Redirects are not followed by default to prevent open redirect attacks
2. **Response body truncation** — Response bodies are truncated to 1KB to prevent memory issues with large responses
3. **Timeout enforcement** — All requests have configurable timeouts to prevent hanging
4. **Input validation** — URLs are validated before making requests
5. **No credential storage** — Headers with sensitive data (API keys, tokens) are passed in memory only, never persisted

## Known Limitations

- TLS certificate verification is enabled by default (Go's default behavior)
- Sensitive headers (e.g., Authorization) should not be stored in config files committed to version control
- Use environment variables for sensitive configuration in production environments
