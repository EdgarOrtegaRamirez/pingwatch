package monitor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EdgarOrtegaRamirez/pingwatch/config"
)

func newTestServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func newTestServerFunc(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestCheckEndpointSuccess(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"status": "ok"}`)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second

	ep := config.EndpointConfig{
		URL:            server.URL,
		Name:           "Test Endpoint",
		Method:         &method,
		ExpectedStatus: &expectedStatus,
		Retries:        &retries,
		RetryDelay:     &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMessage)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
	if result.ResponseTime <= 0 {
		t.Errorf("expected positive response time, got %v", result.ResponseTime)
	}
	if len(result.ValidationErrors) > 0 {
		t.Errorf("expected no validation errors, got: %v", result.ValidationErrors)
	}
}

func TestCheckEndpointWrongStatus(t *testing.T) {
	server := newTestServer(http.StatusNotFound, `{"error": "not found"}`)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second

	ep := config.EndpointConfig{
		URL:            server.URL,
		Name:           "Test Endpoint",
		Method:         &method,
		ExpectedStatus: &expectedStatus,
		Retries:        &retries,
		RetryDelay:     &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if result.Success {
		t.Error("expected failure for wrong status code")
	}
	if result.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, result.StatusCode)
	}
	if len(result.ValidationErrors) == 0 {
		t.Error("expected validation errors for wrong status code")
	}
}

func TestCheckEndpointResponseContains(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data": "hello world"}`)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second
	contains := `"hello world"`

	ep := config.EndpointConfig{
		URL:                server.URL,
		Name:               "Test Endpoint",
		Method:             &method,
		ExpectedStatus:     &expectedStatus,
		ResponseBodyContains: &contains,
		Retries:            &retries,
		RetryDelay:         &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMessage)
	}
}

func TestCheckEndpointResponseNotContains(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"data": "hello world"}`)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second
	notContains := `"error"`

	ep := config.EndpointConfig{
		URL:                   server.URL,
		Name:                  "Test Endpoint",
		Method:                &method,
		ExpectedStatus:        &expectedStatus,
		ResponseBodyNotContains: &notContains,
		Retries:               &retries,
		RetryDelay:            &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMessage)
	}
}

func TestCheckEndpointResponseNotContainsFail(t *testing.T) {
	server := newTestServer(http.StatusOK, `{"error": "something failed"}`)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second
	notContains := `"error"`

	ep := config.EndpointConfig{
		URL:                   server.URL,
		Name:                  "Test Endpoint",
		Method:                &method,
		ExpectedStatus:        &expectedStatus,
		ResponseBodyNotContains: &notContains,
		Retries:               &retries,
		RetryDelay:            &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if result.Success {
		t.Error("expected failure when response contains forbidden string")
	}
}

func TestCheckEndpointMaxResponseTime(t *testing.T) {
	dummyBody := `{"ok": true}`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(dummyBody))
	})
	server := newTestServerFunc(handler)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second
	maxRT := int(50) // 50ms max

	ep := config.EndpointConfig{
		URL:             server.URL,
		Name:            "Slow Endpoint",
		Method:          &method,
		ExpectedStatus:  &expectedStatus,
		MaxResponseTime: &maxRT,
		Retries:         &retries,
		RetryDelay:      &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if result.Success {
		t.Error("expected failure due to response time exceeding max")
	}
	if len(result.ValidationErrors) == 0 {
		t.Error("expected validation error for slow response")
	}
}

func TestCheckEndpointPOST(t *testing.T) {
	receivedBody := ""
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created": true}`))
	})
	server := newTestServerFunc(handler)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "POST"
	expectedStatus := http.StatusCreated
	retries := 1
	retryDelay := 0 * time.Second
	body := `{"test": true}`

	ep := config.EndpointConfig{
		URL:            server.URL,
		Name:           "POST Endpoint",
		Method:         &method,
		ExpectedStatus: &expectedStatus,
		Body:           &body,
		Retries:        &retries,
		RetryDelay:     &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMessage)
	}
	if result.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, result.StatusCode)
	}
	if !strings.Contains(receivedBody, `"test"`) {
		t.Errorf("expected server to receive body with 'test', got: %s", receivedBody)
	}
}

func TestCheckEndpointWithHeaders(t *testing.T) {
	receivedHeader := ""
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Test-Header")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	})
	server := newTestServerFunc(handler)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second

	ep := config.EndpointConfig{
		URL:            server.URL,
		Name:           "Header Endpoint",
		Method:         &method,
		ExpectedStatus: &expectedStatus,
		Headers: map[string]string{
			"X-Test-Header": "test-value-123",
		},
		Retries:    &retries,
		RetryDelay: &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMessage)
	}
	if receivedHeader != "test-value-123" {
		t.Errorf("expected header 'test-value-123', got: %s", receivedHeader)
	}
}

func TestCheckEndpointUnreachable(t *testing.T) {
	mon := NewMonitor(1 * time.Second)
	method := "GET"
	expectedStatus := http.StatusOK
	retries := 1
	retryDelay := 0 * time.Second

	ep := config.EndpointConfig{
		URL:            "http://localhost:59999/unreachable",
		Name:           "Unreachable",
		Method:         &method,
		ExpectedStatus: &expectedStatus,
		Retries:        &retries,
		RetryDelay:     &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	if result.Success {
		t.Error("expected failure for unreachable endpoint")
	}
	if result.ErrorMessage == "" {
		t.Error("expected error message for unreachable endpoint")
	}
}

func TestCheckEndpointRetries(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	})
	server := newTestServerFunc(handler)
	defer server.Close()

	mon := NewMonitor(5 * time.Second)
	method := "GET"
	retries := 3
	retryDelay := 10 * time.Millisecond

	ep := config.EndpointConfig{
		URL:            server.URL,
		Name:           "Retry Endpoint",
		Method:         &method,
		ExpectedStatus: func() *int { v := http.StatusOK; return &v }(),
		Retries:        &retries,
		RetryDelay:     &retryDelay,
	}

	result := mon.CheckEndpoint(ep, retries, retryDelay)

	// Should have succeeded after retries
	if !result.Success {
		t.Errorf("expected success after retries, got failure: %s", result.ErrorMessage)
	}
	// Should have made multiple attempts
	if callCount < 2 {
		t.Errorf("expected multiple attempts, got %d", callCount)
	}
}
