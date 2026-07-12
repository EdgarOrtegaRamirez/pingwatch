package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/EdgarOrtegaRamirez/pingwatch/config"
	"github.com/EdgarOrtegaRamirez/pingwatch/history"
	"github.com/EdgarOrtegaRamirez/pingwatch/monitor"
	"github.com/EdgarOrtegaRamirez/pingwatch/output"
	"github.com/spf13/cobra"
)

var (
	configFile           string
	outputFormat         string
	singleURL            string
	singleMethod         string
	singleTimeout        int
	singleStatus         int
	singleBody           string
	singleContains       string
	dbPath               string
)

var checkCmd = &cobra.Command{
	Use:   "check [urls...]",
	Short: "Check HTTP endpoints",
	Long: `Check one or more HTTP endpoints and report their status.
Can be used with a config file or with command-line arguments.

Examples:
  pingwatch check https://example.com https://api.example.com
  pingwatch check --config config.yaml
  pingwatch check --url https://example.com --timeout 5000
  pingwatch check --url https://example.com --method POST --body '{"key":"value"}'
  pingwatch check --config config.yaml --db history.db`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg *config.MonitorConfig
		var results []monitor.Result

		mon := monitor.NewMonitor(10 * time.Second)

		// If config file is specified, use it
		if configFile != "" {
			var err error
			cfg, err = config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			for _, ep := range cfg.Endpoints {
				result := mon.CheckEndpoint(ep, *ep.Retries, *ep.RetryDelay)
				results = append(results, result)
			}
		} else if singleURL != "" {
			// Single URL mode
			timeout := time.Duration(singleTimeout) * time.Millisecond
			if singleTimeout == 0 {
				timeout = 10000 * time.Millisecond
			}
			mon = monitor.NewMonitor(timeout)

			method := "GET"
			if singleMethod != "" {
				method = singleMethod
			}

			expectedStatus := 200
			if singleStatus != 0 {
				expectedStatus = singleStatus
			}

			ep := config.EndpointConfig{
				URL:            singleURL,
				Name:           singleURL,
				Method:         &method,
				ExpectedStatus: &expectedStatus,
				Timeout:        &timeout,
				Retries:        func() *int { i := 1; return &i }(),
				RetryDelay:     func() *time.Duration { d := 0 * time.Second; return &d }(),
			}

			if singleBody != "" {
				ep.Body = &singleBody
			}
			if singleContains != "" {
				ep.ResponseBodyContains = &singleContains
			}

			result := mon.CheckEndpoint(ep, 1, 0)
			results = append(results, result)
		} else if len(args) > 0 {
			// URL arguments
			for _, url := range args {
				method := "GET"
				timeout := 10 * time.Second
				expectedStatus := 200

				ep := config.EndpointConfig{
					URL:            url,
					Name:           url,
					Method:         &method,
					ExpectedStatus: &expectedStatus,
					Timeout:        &timeout,
					Retries:        func() *int { i := 1; return &i }(),
					RetryDelay:     func() *time.Duration { d := 0 * time.Second; return &d }(),
				}

				result := mon.CheckEndpoint(ep, 1, 0)
				results = append(results, result)
			}
		} else {
			return fmt.Errorf("provide URLs as arguments, use --config for file-based config, or --url for a single URL")
		}

		// Save to history DB if specified
		if dbPath != "" {
			if err := saveResultsToDB(dbPath, results); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save results to history: %v\n", err)
			}
		}

		// Print results
		fmtObj, err := output.FormatFromStr(outputFormat)
		if err != nil {
			return err
		}
		printer := output.NewPrinter(fmtObj, nil)

		if err := printer.PrintResults(results); err != nil {
			return err
		}

		passed := 0
		failed := 0
		for _, r := range results {
			if r.Success {
				passed++
			} else {
				failed++
			}
		}
		printer.PrintSummary(len(results), passed, failed)

		if failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func saveResultsToDB(path string, results []monitor.Result) error {
	store, err := history.NewStore(path)
	if err != nil {
		return fmt.Errorf("failed to open history database: %w", err)
	}
	defer store.Close()

	records := make([]history.CheckRecord, len(results))
	for i, r := range results {
		var sslDays *int
		var sslValid *bool
		if r.SSL != nil {
			sslDays = &r.SSL.DaysRemaining
			sslValid = &r.SSL.IsValid
		}

		records[i] = history.CheckRecord{
			Timestamp:      time.Now(),
			EndpointName:   r.Name,
			URL:            r.URL,
			StatusCode:     r.StatusCode,
			ResponseTimeMs: r.ResponseTimeMs,
			Success:        r.Success,
			ErrorMessage:   r.ErrorMessage,
			SSLDaysLeft:    sslDays,
			SSLValid:       sslValid,
		}
	}

	return store.SaveChecks(records)
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file path (YAML or JSON)")
	checkCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text, json, csv")
	checkCmd.Flags().StringVarP(&singleURL, "url", "u", "", "Single URL to check")
	checkCmd.Flags().StringVarP(&singleMethod, "method", "m", "", "HTTP method (GET, POST, PUT, DELETE)")
	checkCmd.Flags().IntVarP(&singleTimeout, "timeout", "t", 0, "Timeout in milliseconds")
	checkCmd.Flags().IntVarP(&singleStatus, "expected-status", "s", 0, "Expected HTTP status code")
	checkCmd.Flags().StringVarP(&singleBody, "body", "b", "", "Request body (for POST/PUT)")
	checkCmd.Flags().StringVarP(&singleContains, "contains", "", "", "Response body must contain this string")
	checkCmd.Flags().StringVarP(&dbPath, "db", "d", "", "Save results to SQLite history database")
}