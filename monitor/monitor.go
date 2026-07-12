package monitor

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EdgarOrtegaRamirez/pingwatch/config"
)

// checkResult holds the raw result of a single HTTP request
type checkResult struct {
	StatusCode   int
	Body         string
	ResponseTime time.Duration
	Err          error
}

// Result represents the outcome of a single endpoint check
type Result struct {
	URL              string    `json:"url"`
	Name             string    `json:"name"`
	StatusCode       int       `json:"status_code"`
	ResponseTime     time.Duration `json:"response_time_ms"`
	ResponseTimeMs   float64   `json:"response_time_ms_val"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	ResponseBody     string    `json:"-"`
	ValidationErrors []string  `json:"validation_errors,omitempty"`
}

// Monitor performs HTTP checks against configured endpoints
type Monitor struct {
	client *http.Client
}

// NewMonitor creates a new Monitor with the given timeout
func NewMonitor(timeout time.Duration) *Monitor {
	return &Monitor{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// CheckEndpoint performs a single check against an endpoint
func (m *Monitor) CheckEndpoint(ep config.EndpointConfig, retries int, retryDelay time.Duration) Result {
	var lastErr error
	var result Result

	result.URL = ep.URL
	result.Name = ep.Name

	for attempt := 1; attempt <= retries; attempt++ {
		res, err := m.doRequest(ep)
		if err != nil {
			lastErr = err
			if attempt < retries {
				time.Sleep(retryDelay)
			}
			continue
		}

		// Create fresh result for this attempt to avoid stale validation errors
		result = Result{
			URL:  ep.URL,
			Name: ep.Name,
		}
		result.StatusCode = res.StatusCode
		result.ResponseTime = res.ResponseTime
		result.ResponseTimeMs = float64(res.ResponseTime.Milliseconds())
		result.ResponseBody = res.Body

		// Validate status code
		if res.StatusCode != *ep.ExpectedStatus {
			result.ValidationErrors = append(result.ValidationErrors,
				fmt.Sprintf("expected status %d, got %d", *ep.ExpectedStatus, res.StatusCode))
		}

		// Validate response body contains string
		if ep.ResponseBodyContains != nil && *ep.ResponseBodyContains != "" {
			if !strings.Contains(res.Body, *ep.ResponseBodyContains) {
				result.ValidationErrors = append(result.ValidationErrors,
					fmt.Sprintf("response body does not contain: %s", *ep.ResponseBodyContains))
			}
		}

		// Validate response body does NOT contain string
		if ep.ResponseBodyNotContains != nil && *ep.ResponseBodyNotContains != "" {
			if strings.Contains(res.Body, *ep.ResponseBodyNotContains) {
				result.ValidationErrors = append(result.ValidationErrors,
					fmt.Sprintf("response body contains (should not): %s", *ep.ResponseBodyNotContains))
			}
		}

		// Validate response time
		if ep.MaxResponseTime != nil && *ep.MaxResponseTime > 0 {
			if res.ResponseTime.Milliseconds() > int64(*ep.MaxResponseTime) {
				result.ValidationErrors = append(result.ValidationErrors,
					fmt.Sprintf("response time %dms exceeds max %dms",
						res.ResponseTime.Milliseconds(), *ep.MaxResponseTime))
			}
		}

		// Overall success
		result.Success = len(result.ValidationErrors) == 0
		if !result.Success {
			for _, ve := range result.ValidationErrors {
				result.ErrorMessage = ve
				break
			}
			// Retry if there are validation errors
			if attempt < retries {
				time.Sleep(retryDelay)
				continue
			}
		}

		return result
	}

	// All retries failed
	result.Success = false
	result.ErrorMessage = fmt.Sprintf("request failed after %d retries: %v", retries, lastErr)
	result.ResponseTimeMs = -1

	return result
}

func (m *Monitor) doRequest(ep config.EndpointConfig) (*checkResult, error) {
	method := "GET"
	if ep.Method != nil {
		method = *ep.Method
	}
	req, err := http.NewRequest(method, ep.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range ep.Headers {
		req.Header.Set(k, v)
	}

	// Set body if provided
	if ep.Body != nil && *ep.Body != "" {
		req.Body = io.NopCloser(strings.NewReader(*ep.Body))
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := m.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Truncate body for result storage (max 1KB)
	bodyStr := string(body)
	if len(bodyStr) > 1024 {
		bodyStr = bodyStr[:1024] + "...[truncated]"
	}

	return &checkResult{
		StatusCode:   resp.StatusCode,
		Body:         bodyStr,
		ResponseTime: elapsed,
	}, nil
}
