package history

import (
	"os"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-history-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	return store
}

func TestNewStore(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	defer os.Remove(store.Path())

	if store.Path() == "" {
		t.Error("expected non-empty path")
	}
}

func TestSaveAndGetReport(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	defer os.Remove(store.Path())

	now := time.Now().UTC()

	records := []CheckRecord{
		{
			Timestamp:      now.Add(-2 * time.Hour),
			EndpointName:   "example",
			URL:            "https://example.com",
			StatusCode:     200,
			ResponseTimeMs: 50.0,
			Success:        true,
		},
		{
			Timestamp:      now.Add(-1 * time.Hour),
			EndpointName:   "example",
			URL:            "https://example.com",
			StatusCode:     200,
			ResponseTimeMs: 75.0,
			Success:        true,
		},
		{
			Timestamp:      now.Add(-30 * time.Minute),
			EndpointName:   "example",
			URL:            "https://example.com",
			StatusCode:     500,
			ResponseTimeMs: 100.0,
			Success:        false,
			ErrorMessage:   "expected status 200, got 500",
		},
		{
			Timestamp:      now.Add(-5 * time.Minute),
			EndpointName:   "example",
			URL:            "https://example.com",
			StatusCode:     200,
			ResponseTimeMs: 60.0,
			Success:        true,
		},
	}

	if err := store.SaveChecks(records); err != nil {
		t.Fatalf("SaveChecks failed: %v", err)
	}

	// Get report for last 3 hours
	reports, err := store.GetUptimeReport(3 * time.Hour)
	if err != nil {
		t.Fatalf("GetUptimeReport failed: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	r := reports[0]
	if r.EndpointName != "example" {
		t.Errorf("expected endpoint 'example', got '%s'", r.EndpointName)
	}
	if r.TotalChecks != 4 {
		t.Errorf("expected 4 total checks, got %d", r.TotalChecks)
	}
	if r.SuccessfulChecks != 3 {
		t.Errorf("expected 3 successful checks, got %d", r.SuccessfulChecks)
	}
	if r.FailedChecks != 1 {
		t.Errorf("expected 1 failed check, got %d", r.FailedChecks)
	}

	// Uptime: 3/4 = 75%
	expectedUptime := 75.0
	if r.UptimePercent != expectedUptime {
		t.Errorf("expected uptime %.2f%%, got %.2f%%", expectedUptime, r.UptimePercent)
	}

	// Current status should be UP (last check was successful)
	if r.CurrentStatus != "UP" {
		t.Errorf("expected current status 'UP', got '%s'", r.CurrentStatus)
	}
}

func TestEmptyStore(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	defer os.Remove(store.Path())

	reports, err := store.GetUptimeReport(24 * time.Hour)
	if err != nil {
		t.Fatalf("GetUptimeReport failed: %v", err)
	}

	if len(reports) != 0 {
		t.Errorf("expected 0 reports for empty store, got %d", len(reports))
	}
}

func TestPercentiles(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		p        float64
		expected float64
	}{
		{"single element", []float64{42.0}, 50, 42.0},
		{"p50 even", []float64{10, 20, 30, 40}, 50, 25.0},
		{"p50 odd", []float64{10, 20, 30, 40, 50}, 50, 30.0},
		{"p95", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 95, 19.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(tt.data, tt.p)
			if got != tt.expected {
				t.Errorf("percentile(%v, %.0f) = %.2f, want %.2f", tt.data, tt.p, got, tt.expected)
			}
		})
	}
}

func TestMultipleEndpoints(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	defer os.Remove(store.Path())

	now := time.Now().UTC()

	records := []CheckRecord{
		{
			Timestamp:      now.Add(-1 * time.Hour),
			EndpointName:   "api-1",
			URL:            "https://api1.example.com",
			StatusCode:     200,
			ResponseTimeMs: 100.0,
			Success:        true,
		},
		{
			Timestamp:      now.Add(-1 * time.Hour),
			EndpointName:   "api-2",
			URL:            "https://api2.example.com",
			StatusCode:     503,
			ResponseTimeMs: 200.0,
			Success:        false,
			ErrorMessage:   "service unavailable",
		},
	}

	if err := store.SaveChecks(records); err != nil {
		t.Fatalf("SaveChecks failed: %v", err)
	}

	reports, err := store.GetUptimeReport(2 * time.Hour)
	if err != nil {
		t.Fatalf("GetUptimeReport failed: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}

	// Should be sorted by endpoint name
	if reports[0].EndpointName != "api-1" {
		t.Errorf("expected first report 'api-1', got '%s'", reports[0].EndpointName)
	}
	if reports[1].EndpointName != "api-2" {
		t.Errorf("expected second report 'api-2', got '%s'", reports[1].EndpointName)
	}

	if reports[0].CurrentStatus != "UP" {
		t.Errorf("expected api-1 status 'UP', got '%s'", reports[0].CurrentStatus)
	}
	if !stringsHasPrefix(reports[1].CurrentStatus, "DOWN") {
		t.Errorf("expected api-2 status starting with 'DOWN', got '%s'", reports[1].CurrentStatus)
	}
}

func TestSaveCheckSingle(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	defer os.Remove(store.Path())

	record := CheckRecord{
		Timestamp:      time.Now().UTC(),
		EndpointName:   "test",
		URL:            "https://test.example.com",
		StatusCode:     200,
		ResponseTimeMs: 42.0,
		Success:        true,
	}

	if err := store.SaveCheck(record); err != nil {
		t.Fatalf("SaveCheck failed: %v", err)
	}
}

func TestLatestStatus(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	defer os.Remove(store.Path())

	now := time.Now().UTC()

	records := []CheckRecord{
		{Timestamp: now.Add(-1 * time.Hour), EndpointName: "ep1", URL: "https://ep1.com", Success: true, ResponseTimeMs: 10},
		{Timestamp: now.Add(-30 * time.Minute), EndpointName: "ep2", URL: "https://ep2.com", Success: false, ErrorMessage: "error", ResponseTimeMs: 20},
		{Timestamp: now.Add(-10 * time.Minute), EndpointName: "ep1", URL: "https://ep1.com", Success: false, ErrorMessage: "timeout", ResponseTimeMs: 30},
	}

	if err := store.SaveChecks(records); err != nil {
		t.Fatalf("SaveChecks failed: %v", err)
	}

	status, err := store.GetLatestStatus()
	if err != nil {
		t.Fatalf("GetLatestStatus failed: %v", err)
	}

	if len(status) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(status))
	}

	// ep1's latest check was a failure
	if status["ep1"] != false {
		t.Error("expected ep1 status to be false (last check failed)")
	}
	// ep2's only check was a failure
	if status["ep2"] != false {
		t.Error("expected ep2 status to be false")
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
