package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EdgarOrtegaRamirez/pingwatch/history"
	"github.com/spf13/cobra"
)

var (
	reportDBPath   string
	reportFormat   string
	reportPeriod   string
	reportSinceStr string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate uptime and performance reports",
	Long: `Generate uptime, performance, and SSL certificate reports from historical check data.

Reports include:
  - Uptime percentage over the specified period
  - Response time percentiles (P50/P95/P99)
  - Total, successful, and failed checks
  - Current endpoint status

Period options: 1h, 6h, 12h, 24h, 7d, 30d (default: 24h)

Examples:
  pingwatch report --db history.db
  pingwatch report --db history.db --period 7d
  pingwatch report --db history.db --period 30d --format json
  pingwatch report --db history.db --since "2026-07-01T00:00:00Z"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportDBPath == "" {
			return fmt.Errorf("--db flag is required (path to history database)")
		}

		// Open the history store
		store, err := history.NewStore(reportDBPath)
		if err != nil {
			return fmt.Errorf("failed to open history database: %w", err)
		}
		defer store.Close()

		var reports []history.UptimeReport

		if reportSinceStr != "" {
			start, err := time.Parse(time.RFC3339, reportSinceStr)
			if err != nil {
				return fmt.Errorf("invalid --since value (use RFC3339 format like \"2026-07-01T00:00:00Z\"): %w", err)
			}
			reports, err = store.GetUptimeReportRange(start, time.Now())
			if err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}
		} else {
			// Parse period duration
			duration, err := parseDuration(reportPeriod)
			if err != nil {
				return fmt.Errorf("invalid period: %w", err)
			}
			reports, err = store.GetUptimeReport(duration)
			if err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}
		}

		if len(reports) == 0 {
			fmt.Println("No data found for the specified period.")
			return nil
		}

		// Output based on format
		switch strings.ToLower(reportFormat) {
		case "json":
			return printReportJSON(reports)
		case "csv":
			return printReportCSV(reports)
		default:
			return printReportText(reports)
		}
	},
}

func parseDuration(s string) (time.Duration, error) {
	// Support human-friendly units
	switch strings.ToLower(s) {
	case "1h":
		return time.Hour, nil
	case "6h":
		return 6 * time.Hour, nil
	case "12h":
		return 12 * time.Hour, nil
	case "24h", "1d", "day":
		return 24 * time.Hour, nil
	case "7d", "7days", "week":
		return 7 * 24 * time.Hour, nil
	case "30d", "30days", "month":
		return 30 * 24 * time.Hour, nil
	case "90d", "90days", "quarter":
		return 90 * 24 * time.Hour, nil
	default:
		return time.ParseDuration(s)
	}
}

func printReportText(reports []history.UptimeReport) error {
	for _, r := range reports {
		fmt.Println(strings.Repeat("━", 60))
		fmt.Printf("  %s\n", r.EndpointName)
		fmt.Printf("  URL: %s\n", r.URL)

		// Status badge
		statusBadge := "● UP"
		if strings.HasPrefix(r.CurrentStatus, "DOWN") {
			statusBadge = "◉ DOWN"
		}
		fmt.Printf("  Status: %s\n", statusBadge)

		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("  Period:             %s → %s\n", r.PeriodStart[:10], r.PeriodEnd[:10])
		fmt.Printf("  Total checks:       %d\n", r.TotalChecks)
		fmt.Printf("  Successful:         %d\n", r.SuccessfulChecks)
		fmt.Printf("  Failed:             %d\n", r.FailedChecks)
		fmt.Printf("  Uptime:             %.2f%%\n", r.UptimePercent)

		fmt.Println()
		fmt.Println("  Response Times:")
		fmt.Printf("    Average:          %.1f ms\n", r.AvgResponseMs)
		fmt.Printf("    Min:              %.1f ms\n", r.MinResponseMs)
		fmt.Printf("    Max:              %.1f ms\n", r.MaxResponseMs)
		fmt.Printf("    P50 (median):     %.1f ms\n", r.PercentileP50)
		fmt.Printf("    P95:              %.1f ms\n", r.PercentileP95)
		fmt.Printf("    P99:              %.1f ms\n", r.PercentileP99)

		fmt.Println()
	}

	// Overall summary
	total := len(reports)
	upCount := 0
	for _, r := range reports {
		if !strings.HasPrefix(r.CurrentStatus, "DOWN") {
			upCount++
		}
	}

	fmt.Println(strings.Repeat("━", 60))
	fmt.Printf("  Endpoints: %d | UP: %d | DOWN: %d\n", total, upCount, total-upCount)

	return nil
}

func printReportJSON(reports []history.UptimeReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(reports)
}

func printReportCSV(reports []history.UptimeReport) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header
	if err := w.Write([]string{
		"endpoint", "url", "total_checks", "successful", "failed",
		"uptime_pct", "avg_ms", "min_ms", "max_ms", "p50_ms", "p95_ms", "p99_ms",
		"current_status", "period_start", "period_end",
	}); err != nil {
		return err
	}

	for _, r := range reports {
		record := []string{
			r.EndpointName,
			r.URL,
			fmt.Sprintf("%d", r.TotalChecks),
			fmt.Sprintf("%d", r.SuccessfulChecks),
			fmt.Sprintf("%d", r.FailedChecks),
			fmt.Sprintf("%.2f", r.UptimePercent),
			fmt.Sprintf("%.1f", r.AvgResponseMs),
			fmt.Sprintf("%.1f", r.MinResponseMs),
			fmt.Sprintf("%.1f", r.MaxResponseMs),
			fmt.Sprintf("%.1f", r.PercentileP50),
			fmt.Sprintf("%.1f", r.PercentileP95),
			fmt.Sprintf("%.1f", r.PercentileP99),
			r.CurrentStatus,
			r.PeriodStart,
			r.PeriodEnd,
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&reportDBPath, "db", "d", "", "Path to history database (required)")
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "text", "Output format: text, json, csv")
	reportCmd.Flags().StringVarP(&reportPeriod, "period", "p", "24h", "Report period: 1h, 6h, 12h, 24h, 7d, 30d, 90d")
	reportCmd.Flags().StringVarP(&reportSinceStr, "since", "s", "", "Explicit start date (RFC3339 format, overrides --period)")
}
