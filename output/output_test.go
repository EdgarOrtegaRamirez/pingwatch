package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/EdgarOrtegaRamirez/pingwatch/monitor"
)

func TestPrintTextSuccess(t *testing.T) {
	var buf bytes.Buffer
	printer := NewPrinter(FormatText, &buf)

	results := []monitor.Result{
		{
			URL:            "https://example.com",
			Name:           "Example",
			StatusCode:     200,
			ResponseTime:   50 * time.Millisecond,
			ResponseTimeMs: 50.0,
			Success:        true,
		},
	}

	if err := printer.PrintResults(results); err != nil {
		t.Fatalf("PrintResults failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Error("expected success marker in output")
	}
	if !strings.Contains(output, "Example") {
		t.Error("expected endpoint name in output")
	}
}

func TestPrintTextFailure(t *testing.T) {
	var buf bytes.Buffer
	printer := NewPrinter(FormatText, &buf)

	results := []monitor.Result{
		{
			URL:            "https://example.com",
			Name:           "Example",
			StatusCode:     500,
			ResponseTime:   100 * time.Millisecond,
			ResponseTimeMs: 100.0,
			Success:        false,
			ErrorMessage:   "expected status 200, got 500",
		},
	}

	if err := printer.PrintResults(results); err != nil {
		t.Fatalf("PrintResults failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Error("expected failure marker in output")
	}
	if !strings.Contains(output, "500") {
		t.Error("expected status code in output")
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	printer := NewPrinter(FormatJSON, &buf)

	results := []monitor.Result{
		{
			URL:            "https://example.com",
			Name:           "Example",
			StatusCode:     200,
			ResponseTime:   50 * time.Millisecond,
			ResponseTimeMs: 50.0,
			Success:        true,
		},
	}

	if err := printer.PrintResults(results); err != nil {
		t.Fatalf("PrintResults failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"success": true`) {
		t.Error("expected JSON success field")
	}
	if !strings.Contains(output, `"status_code": 200`) {
		t.Error("expected JSON status_code field")
	}
}

func TestPrintCSV(t *testing.T) {
	var buf bytes.Buffer
	printer := NewPrinter(FormatCSV, &buf)

	results := []monitor.Result{
		{
			URL:            "https://example.com",
			Name:           "Example",
			StatusCode:     200,
			ResponseTime:   50 * time.Millisecond,
			ResponseTimeMs: 50.0,
			Success:        true,
		},
	}

	if err := printer.PrintResults(results); err != nil {
		t.Fatalf("PrintResults failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines (header + data), got %d", len(lines))
	}

	// Check header
	header := lines[0]
	if !strings.Contains(header, "name") || !strings.Contains(header, "url") {
		t.Error("expected CSV header with name and url columns")
	}

	// Check data
	data := lines[1]
	if !strings.Contains(data, "Example") {
		t.Error("expected endpoint name in CSV data")
	}
}

func TestPrintSummary(t *testing.T) {
	var buf bytes.Buffer
	printer := NewPrinter(FormatText, &buf)

	printer.PrintSummary(3, 2, 1)
	output := buf.String()

	if !strings.Contains(output, "Total: 3") {
		t.Error("expected total count in summary")
	}
	if !strings.Contains(output, "Passed: 2") {
		t.Error("expected passed count in summary")
	}
	if !strings.Contains(output, "Failed: 1") {
		t.Error("expected failed count in summary")
	}
	if !strings.Contains(output, "FAILED") {
		t.Error("expected FAILED status in summary")
	}
}

func TestPrintSummaryAllPassed(t *testing.T) {
	var buf bytes.Buffer
	printer := NewPrinter(FormatText, &buf)

	printer.PrintSummary(3, 3, 0)
	output := buf.String()

	if !strings.Contains(output, "ALL PASSED") {
		t.Error("expected ALL PASSED status in summary")
	}
}

func TestFormatFromStr(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
		hasError bool
	}{
		{"text", FormatText, false},
		{"json", FormatJSON, false},
		{"csv", FormatCSV, false},
		{"", FormatText, false},
		{"TEXT", FormatText, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		got, err := FormatFromStr(tt.input)
		if tt.hasError && err == nil {
			t.Errorf("FormatFromStr(%q) expected error, got nil", tt.input)
		}
		if !tt.hasError && got != tt.expected {
			t.Errorf("FormatFromStr(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDurationToText(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{500 * time.Microsecond, "500µs"},
		{50 * time.Millisecond, "50ms"},
		{1500 * time.Millisecond, "1.50s"},
	}

	for _, tt := range tests {
		got := DurationToText(tt.input)
		if got != tt.expected {
			t.Errorf("DurationToText(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
