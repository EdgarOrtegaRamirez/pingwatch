package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// CheckRecord represents a stored check result
type CheckRecord struct {
	ID            int64     `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	EndpointName  string    `json:"endpoint_name"`
	URL           string    `json:"url"`
	StatusCode    int       `json:"status_code"`
	ResponseTimeMs float64  `json:"response_time_ms"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	SSLDaysLeft   *int      `json:"ssl_days_left,omitempty"`
	SSLValid      *bool     `json:"ssl_valid,omitempty"`
}

// UptimeReport contains uptime statistics for an endpoint
type UptimeReport struct {
	EndpointName    string  `json:"endpoint_name"`
	URL             string  `json:"url"`
	TotalChecks     int     `json:"total_checks"`
	SuccessfulChecks int    `json:"successful_checks"`
	FailedChecks    int     `json:"failed_checks"`
	UptimePercent   float64 `json:"uptime_percent"`
	PeriodStart     string  `json:"period_start"`
	PeriodEnd       string  `json:"period_end"`
	CurrentStatus   string  `json:"current_status"`
	PercentileP50   float64 `json:"percentile_p50_ms"`
	PercentileP95   float64 `json:"percentile_p95_ms"`
	PercentileP99   float64 `json:"percentile_p99_ms"`
	AvgResponseMs   float64 `json:"avg_response_ms"`
	MinResponseMs   float64 `json:"min_response_ms"`
	MaxResponseMs   float64 `json:"max_response_ms"`
}

// Store manages the SQLite database for historical check data
type Store struct {
	db   *sql.DB
	mu   sync.Mutex
	path string
}

// NewStore creates or opens a history database
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas for performance and safety
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	store := &Store{db: db, path: path}
	if err := store.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

func (s *Store) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		endpoint_name TEXT NOT NULL,
		url TEXT NOT NULL,
		status_code INTEGER NOT NULL DEFAULT 0,
		response_time_ms REAL NOT NULL DEFAULT 0,
		success INTEGER NOT NULL DEFAULT 0,
		error_message TEXT DEFAULT '',
		ssl_days_left INTEGER,
		ssl_valid INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_checks_timestamp ON checks(timestamp);
	CREATE INDEX IF NOT EXISTS idx_checks_endpoint ON checks(endpoint_name);
	CREATE INDEX IF NOT EXISTS idx_checks_success ON checks(success);
	CREATE INDEX IF NOT EXISTS idx_checks_url ON checks(url);
	`
	_, err := s.db.Exec(schema)
	return err
}

// SaveCheck stores a check result in the database
func (s *Store) SaveCheck(record CheckRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO checks (timestamp, endpoint_name, url, status_code, response_time_ms, success, error_message, ssl_days_left, ssl_valid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.Timestamp.Format(time.RFC3339),
		record.EndpointName,
		record.URL,
		record.StatusCode,
		record.ResponseTimeMs,
		boolToInt(record.Success),
		record.ErrorMessage,
		record.SSLDaysLeft,
		record.SSLValid,
	)
	return err
}

// SaveChecks stores multiple check results in a transaction
func (s *Store) SaveChecks(records []CheckRecord) error {
	if len(records) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO checks (timestamp, endpoint_name, url, status_code, response_time_ms, success, error_message, ssl_days_left, ssl_valid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		_, err := stmt.Exec(
			record.Timestamp.Format(time.RFC3339),
			record.EndpointName,
			record.URL,
			record.StatusCode,
			record.ResponseTimeMs,
			boolToInt(record.Success),
			record.ErrorMessage,
			record.SSLDaysLeft,
			record.SSLValid,
		)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	return tx.Commit()
}

// GetUptimeReport generates an uptime report for a time period
func (s *Store) GetUptimeReport(since time.Duration) ([]UptimeReport, error) {
	end := time.Now()
	start := end.Add(-since)

	return s.GetUptimeReportRange(start, end)
}

// GetUptimeReportRange generates an uptime report for a specific time range
func (s *Store) GetUptimeReportRange(start, end time.Time) ([]UptimeReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(
		`SELECT endpoint_name, url, success, response_time_ms, timestamp, ssl_days_left, ssl_valid, error_message
		FROM checks
		WHERE timestamp >= ? AND timestamp <= ?
		ORDER BY endpoint_name, timestamp`,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query checks: %w", err)
	}
	defer rows.Close()

	// Group records by endpoint
	type record struct {
		success        bool
		responseTimeMs float64
		timestamp      string
		sslDaysLeft    *int
		sslValid       *bool
		errorMessage   string
	}

	endpointRecords := make(map[string][]record)
	endpointURLs := make(map[string]string)

	for rows.Next() {
		var name, url, ts, errMsg string
		var success int
		var rtMs float64
		var sslDays *int
		var sslValid *bool

		if err := rows.Scan(&name, &url, &success, &rtMs, &ts, &sslDays, &sslValid, &errMsg); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		endpointURLs[name] = url
		endpointRecords[name] = append(endpointRecords[name], record{
			success:        success == 1,
			responseTimeMs: rtMs,
			timestamp:      ts,
			sslDaysLeft:    sslDays,
			sslValid:       sslValid,
			errorMessage:   errMsg,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if len(endpointRecords) == 0 {
		return []UptimeReport{}, nil
	}

	var reports []UptimeReport
	latestTimestap := ""

	for name, records := range endpointRecords {
		report := UptimeReport{
			EndpointName: name,
			URL:          endpointURLs[name],
		}

		total := len(records)
		successCount := 0
		var rTimes []float64
		var sumRT float64

		for _, r := range records {
			if r.success {
				successCount++
			}
			rTimes = append(rTimes, r.responseTimeMs)
			sumRT += r.responseTimeMs

			if r.timestamp > latestTimestap {
				latestTimestap = r.timestamp
			}
		}

		report.TotalChecks = total
		report.SuccessfulChecks = successCount
		report.FailedChecks = total - successCount

		if total > 0 {
			report.UptimePercent = math.Round(float64(successCount)/float64(total)*10000) / 100
			report.AvgResponseMs = math.Round(sumRT/float64(total)*100) / 100
		}

		// Calculate percentiles
		sort.Float64s(rTimes)
		if len(rTimes) > 0 {
			report.MinResponseMs = rTimes[0]
			report.MaxResponseMs = rTimes[len(rTimes)-1]
			report.PercentileP50 = percentile(rTimes, 50)
			report.PercentileP95 = percentile(rTimes, 95)
			report.PercentileP99 = percentile(rTimes, 99)
		}

		report.PeriodStart = start.Format(time.RFC3339)
		report.PeriodEnd = end.Format(time.RFC3339)

		// Current status (last check)
		if len(records) > 0 {
			lastRec := records[len(records)-1]
			if lastRec.success {
				report.CurrentStatus = "UP"
			} else {
				report.CurrentStatus = "DOWN"
				if lastRec.errorMessage != "" {
					report.CurrentStatus += " (" + lastRec.errorMessage + ")"
				}
			}
		}

		reports = append(reports, report)
	}

	// Sort reports by endpoint name
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].EndpointName < reports[j].EndpointName
	})

	return reports, nil
}

// Percentile calculates the p-th percentile from sorted data
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper {
		return math.Round(sorted[lower]*100) / 100
	}

	fraction := rank - float64(lower)
	value := sorted[lower] + fraction*(sorted[upper]-sorted[lower])
	return math.Round(value*100) / 100
}

// GetLatestStatus returns the most recent status for each endpoint
func (s *Store) GetLatestStatus() (map[string]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(
		`SELECT endpoint_name, success
		FROM checks
		WHERE id IN (
				SELECT MAX(id) FROM checks GROUP BY endpoint_name
		)`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest status: %w", err)
	}
	defer rows.Close()

	status := make(map[string]bool)
	for rows.Next() {
		var name string
		var success int
		if err := rows.Scan(&name, &success); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		status[name] = success == 1
	}
	return status, rows.Err()
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Path returns the database file path
func (s *Store) Path() string {
	return s.path
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// MarshalJSON for UptimeReport to ensure proper JSON output
func (r UptimeReport) MarshalJSON() ([]byte, error) {
	type Alias UptimeReport
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&r),
	})
}