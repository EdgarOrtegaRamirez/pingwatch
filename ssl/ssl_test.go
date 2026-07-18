package ssl

import (
	"testing"
)

func TestExtractHostnameHTTPS(t *testing.T) {
	hostname := ExtractHostname("https://example.com/path?query=1")
	if hostname != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", hostname)
	}
}

func TestExtractHostnameHTTP(t *testing.T) {
	hostname := ExtractHostname("http://api.example.com/v1/users")
	if hostname != "api.example.com" {
		t.Errorf("expected 'api.example.com', got '%s'", hostname)
	}
}

func TestExtractHostnameNoProtocol(t *testing.T) {
	hostname := ExtractHostname("example.com:8080")
	if hostname != "example.com:8080" {
		t.Errorf("expected 'example.com:8080', got '%s'", hostname)
	}
}

func TestExtractHostnameWithPort(t *testing.T) {
	hostname := ExtractHostname("https://example.com:8443/path")
	if hostname != "example.com:8443" {
		t.Errorf("expected 'example.com:8443', got '%s'", hostname)
	}
}

func TestExtractHostnamePlainDomain(t *testing.T) {
	hostname := ExtractHostname("https://example.com")
	if hostname != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", hostname)
	}
}

func TestExtractHostnameSubdomain(t *testing.T) {
	hostname := ExtractHostname("https://sub.domain.example.co.uk/api/v2")
	if hostname != "sub.domain.example.co.uk" {
		t.Errorf("expected 'sub.domain.example.co.uk', got '%s'", hostname)
	}
}
