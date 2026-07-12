package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigYAML(t *testing.T) {
	yamlContent := `
defaults:
  interval: 30s
  timeout: 5s
  expected_status: 200
  method: GET
  retries: 3

endpoints:
  - url: https://example.com
    name: Example
  - url: https://api.example.com
    name: API
    expected_status: 201
    max_response_time_ms: 1000
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Defaults.Interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", cfg.Defaults.Interval)
	}
	if cfg.Defaults.Retries != 3 {
		t.Errorf("expected default retries 3, got %d", cfg.Defaults.Retries)
	}
	if len(cfg.Endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(cfg.Endpoints))
	}
	if cfg.Endpoints[0].Name != "Example" {
		t.Errorf("expected endpoint name 'Example', got '%s'", cfg.Endpoints[0].Name)
	}
	if cfg.Endpoints[1].ExpectedStatus == nil || *cfg.Endpoints[1].ExpectedStatus != 201 {
		t.Errorf("expected endpoint 2 status 201, got %v", cfg.Endpoints[1].ExpectedStatus)
	}
}

func TestLoadConfigJSON(t *testing.T) {
	jsonContent := `{
  "defaults": {
    "interval": 60000000000,
    "timeout": 10000000000,
    "expected_status": 200
  },
  "endpoints": [
    {
      "url": "https://example.com",
      "name": "Example"
    }
  ]
}`
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(cfg.Endpoints))
	}
	if cfg.Endpoints[0].Name != "Example" {
		t.Errorf("expected endpoint name 'Example', got '%s'", cfg.Endpoints[0].Name)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	yamlContent := `
endpoints:
  - url: https://example.com
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify defaults are applied
	if cfg.Defaults.Interval != 60*time.Second {
		t.Errorf("expected default interval 60s, got %v", cfg.Defaults.Interval)
	}
	if cfg.Defaults.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", cfg.Defaults.Timeout)
	}
	if cfg.Defaults.ExpectedStatus != 200 {
		t.Errorf("expected default status 200, got %d", cfg.Defaults.ExpectedStatus)
	}
	if cfg.Defaults.Method != "GET" {
		t.Errorf("expected default method GET, got %s", cfg.Defaults.Method)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("invalid: yaml: content: ["); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestEndpointNameDefault(t *testing.T) {
	yamlContent := `
endpoints:
  - url: https://example.com
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Endpoint name should default to URL
	if cfg.Endpoints[0].Name != "https://example.com" {
		t.Errorf("expected endpoint name to default to URL, got '%s'", cfg.Endpoints[0].Name)
	}
}
