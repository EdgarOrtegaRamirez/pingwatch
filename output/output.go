package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/EdgarOrtegaRamirez/pingwatch/monitor"
)

// Format is the output format type
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

// Printer handles output formatting
type Printer struct {
	writer io.Writer
	format Format
}

// NewPrinter creates a new Printer
func NewPrinter(format Format, writer io.Writer) *Printer {
	if writer == nil {
		writer = os.Stdout
	}
	return &Printer{writer: writer, format: format}
}

// PrintResults outputs the monitor results in the configured format
func (p *Printer) PrintResults(results []monitor.Result) error {
	switch p.format {
	case FormatJSON:
		return p.printJSON(results)
	case FormatCSV:
		return p.printCSV(results)
	default:
		return p.printText(results)
	}
}

// PrintSummary outputs a summary line
func (p *Printer) PrintSummary(total, passed, failed int) {
	if p.format == FormatText {
		fmt.Fprintf(p.writer, "\n=== Summary ===\n")
		fmt.Fprintf(p.writer, "Total: %d | Passed: %d | Failed: %d\n", total, passed, failed)
		if failed > 0 {
			fmt.Fprintf(p.writer, "Status: FAILED\n")
		} else {
			fmt.Fprintf(p.writer, "Status: ALL PASSED\n")
		}
	}
}

func (p *Printer) printText(results []monitor.Result) error {
	for _, r := range results {
		status := "✓"
		if !r.Success {
			status = "✗"
		}

		line := fmt.Sprintf("%s %s [%d] %.0fms",
			status, r.Name, r.StatusCode, r.ResponseTimeMs)

		// Add SSL info
		if r.SSL != nil {
			if r.SSL.IsValid {
				line += fmt.Sprintf(" | SSL: %s (%d days)", r.SSL.Subject, r.SSL.DaysRemaining)
			} else if r.SSL.Error != "" {
				line += fmt.Sprintf(" | SSL: %s", r.SSL.Error)
			}
		}

		if !r.Success {
			line += fmt.Sprintf(" — %s", r.ErrorMessage)
			if len(r.ValidationErrors) > 1 {
				for _, ve := range r.ValidationErrors[1:] {
					line += fmt.Sprintf("\n  • %s", ve)
				}
			}
		}

		fmt.Fprintln(p.writer, line)
	}
	return nil
}

func (p *Printer) printJSON(results []monitor.Result) error {
	encoder := json.NewEncoder(p.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func (p *Printer) printCSV(results []monitor.Result) error {
	w := csv.NewWriter(p.writer)
	defer w.Flush()

	// Header
	headers := []string{"name", "url", "status_code", "response_time_ms", "success", "error_message"}
	headers = append(headers, "ssl_subject", "ssl_days_remaining", "ssl_valid")
	if err := w.Write(headers); err != nil {
		return err
	}

	for _, r := range results {
		sslSubject := ""
		sslDays := ""
		sslValid := ""
		if r.SSL != nil {
			sslSubject = r.SSL.Subject
			sslDays = fmt.Sprintf("%d", r.SSL.DaysRemaining)
			sslValid = fmt.Sprintf("%t", r.SSL.IsValid)
		}

		record := []string{
			r.Name,
			r.URL,
			fmt.Sprintf("%d", r.StatusCode),
			fmt.Sprintf("%.0f", r.ResponseTimeMs),
			fmt.Sprintf("%t", r.Success),
			r.ErrorMessage,
			sslSubject,
			sslDays,
			sslValid,
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// FormatFromStr converts a string to Format type
func FormatFromStr(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	default:
		return "", fmt.Errorf("unknown output format: %s (valid: text, json, csv)", s)
	}
}

// DurationToText converts a duration to a human-readable string
func DurationToText(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
