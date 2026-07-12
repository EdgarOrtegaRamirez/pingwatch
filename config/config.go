package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// MonitorConfig is the top-level configuration for pingwatch
type MonitorConfig struct {
	// Defaults apply to all endpoints unless overridden
	Defaults DefaultConfig `yaml:"defaults,omitempty"`
	// Endpoints to monitor
	Endpoints []EndpointConfig `yaml:"endpoints"`
	// Output format: "text", "json", "csv"
	OutputFormat string `yaml:"output_format,omitempty"`
	// Alert file path (optional)
	AlertFile string `yaml:"alert_file,omitempty"`
}

// DefaultConfig contains default settings for all endpoints
type DefaultConfig struct {
	// Interval between checks
	Interval time.Duration `yaml:"interval,omitempty"`
	// Timeout for each request
	Timeout time.Duration `yaml:"timeout,omitempty"`
	// Expected status code
	ExpectedStatus int `yaml:"expected_status,omitempty"`
	// Method to use
	Method string `yaml:"method,omitempty"`
	// Headers to send
	Headers map[string]string `yaml:"headers,omitempty"`
	// Body to send (for POST/PUT)
	Body string `yaml:"body,omitempty"`
	// Validate response body contains this string
	ResponseBodyContains string `yaml:"response_body_contains,omitempty"`
	// Validate response body does NOT contain this string
	ResponseBodyNotContains string `yaml:"response_body_not_contains,omitempty"`
	// Maximum response time in milliseconds
	MaxResponseTime int `yaml:"max_response_time_ms,omitempty"`
	// Number of retries on failure
	Retries int `yaml:"retries,omitempty"`
	// Retry delay
	RetryDelay time.Duration `yaml:"retry_delay,omitempty"`
}

// EndpointConfig defines a single endpoint to monitor
type EndpointConfig struct {
	// URL to monitor
	URL string `yaml:"url"`
	// Name for this endpoint (for display)
	Name string `yaml:"name,omitempty"`
	// Override defaults
	Interval *time.Duration `yaml:"interval,omitempty"`
	Timeout  *time.Duration `yaml:"timeout,omitempty"`
	// Expected status code
	ExpectedStatus *int `yaml:"expected_status,omitempty"`
	// Method to use
	Method *string `yaml:"method,omitempty"`
	// Headers to send
	Headers map[string]string `yaml:"headers,omitempty"`
	// Body to send (for POST/PUT)
	Body *string `yaml:"body,omitempty"`
	// Validate response body contains this string
	ResponseBodyContains *string `yaml:"response_body_contains,omitempty"`
	// Validate response body does NOT contain this string
	ResponseBodyNotContains *string `yaml:"response_body_not_contains,omitempty"`
	// Maximum response time in milliseconds
	MaxResponseTime *int `yaml:"max_response_time_ms,omitempty"`
	// Number of retries on failure
	Retries *int `yaml:"retries,omitempty"`
	// Retry delay
	RetryDelay *time.Duration `yaml:"retry_delay,omitempty"`
}

// LoadConfig reads a config file from the given path
func LoadConfig(path string) (*MonitorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &MonitorConfig{}

	// Detect format by extension or content
	switch {
	case hasYAMLExtension(path):
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case hasJSONExtension(path):
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, cfg); err != nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config (tried YAML and JSON): %w", err)
			}
		}
	}

	// Apply defaults
	applyDefaults(cfg)

	return cfg, nil
}

func hasYAMLExtension(path string) bool {
	ext := strings.ToLower(path)
	return strings.HasSuffix(ext, ".yaml") || strings.HasSuffix(ext, ".yml")
}

func hasJSONExtension(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".json")
}

// applyDefaults sets default values for unset fields
func applyDefaults(cfg *MonitorConfig) {
	if cfg.Defaults.Interval == 0 {
		cfg.Defaults.Interval = 60 * time.Second
	}
	if cfg.Defaults.Timeout == 0 {
		cfg.Defaults.Timeout = 10 * time.Second
	}
	if cfg.Defaults.ExpectedStatus == 0 {
		cfg.Defaults.ExpectedStatus = http.StatusOK
	}
	if cfg.Defaults.Method == "" {
		cfg.Defaults.Method = http.MethodGet
	}
	if cfg.Defaults.Retries == 0 {
		cfg.Defaults.Retries = 1
	}

	for i := range cfg.Endpoints {
		ep := &cfg.Endpoints[i]
		if ep.Interval == nil {
			v := cfg.Defaults.Interval
			ep.Interval = &v
		}
		if ep.Timeout == nil {
			v := cfg.Defaults.Timeout
			ep.Timeout = &v
		}
		if ep.ExpectedStatus == nil {
			v := cfg.Defaults.ExpectedStatus
			ep.ExpectedStatus = &v
		}
		if ep.Method == nil {
			v := cfg.Defaults.Method
			ep.Method = &v
		}
		if ep.Retries == nil {
			v := cfg.Defaults.Retries
			ep.Retries = &v
		}
		if ep.Name == "" {
			ep.Name = ep.URL
		}
	}
}
