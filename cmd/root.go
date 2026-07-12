package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "pingwatch",
	Short: "Lightweight HTTP endpoint monitoring CLI",
	Long: `PingWatch is a CLI tool for monitoring HTTP endpoints. It checks endpoints at configurable intervals,
validates response status codes, response times, and response bodies, and reports results
in multiple formats. Designed for CI/CD pipelines and local development.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
	rootCmd.SetVersionTemplate("PingWatch version {{.Version}}\n")
}
