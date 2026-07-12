package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configDir string
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new pingwatch config file",
	Long: `Creates a sample pingwatch configuration file.
Defaults to config.yaml in the current directory, or the specified path.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := "config.yaml"
		if len(args) > 0 {
			path = args[0]
		}

		// Create directory if needed
		if configDir != "" {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
				os.Exit(1)
			}
			if !osPathAbsolute(path) {
				path = configDir + "/" + path
			}
		}

		// Check if file already exists
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "Config file already exists: %s\n", path)
			os.Exit(1)
		}

		sample := `# PingWatch Configuration File
# https://github.com/EdgarOrtegaRamirez/pingwatch

# Default settings for all endpoints
defaults:
  interval: 60s
  timeout: 10s
  expected_status: 200
  method: GET
  retries: 1
  retry_delay: 1s

# Endpoints to monitor
endpoints:
  - url: https://example.com
    name: Example
    expected_status: 200
    max_response_time_ms: 5000

  - url: https://httpbin.org/get
    name: HTTPBin
    expected_status: 200
    max_response_time_ms: 3000

  - url: https://httpbin.org/post
    name: HTTPBin POST
    method: POST
    body: '{"test": true}'
    expected_status: 200
    response_body_contains: '"test": true'
`

		if err := os.WriteFile(path, []byte(sample), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Config file created: %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&configDir, "dir", "d", "", "Directory to create config in")
}

func osPathAbsolute(path string) bool {
	return len(path) > 0 && path[0] == '/'
}
