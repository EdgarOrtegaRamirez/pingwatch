package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/EdgarOrtegaRamirez/pingwatch/config"
	"github.com/spf13/cobra"
)

var showConfigFile string

var showCmd = &cobra.Command{
	Use:   "show [path]",
	Short: "Show parsed config from a file",
	Long:  "Display the parsed configuration from a YAML or JSON file.",
	Run: func(cmd *cobra.Command, args []string) {
		path := showConfigFile
		if len(args) > 0 {
			path = args[0]
		}
		if path == "" {
			fmt.Fprintln(os.Stderr, "Error: specify a config file path")
			os.Exit(1)
		}

		cfg, err := config.LoadConfig(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(cfg)
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().StringVarP(&showConfigFile, "config", "c", "", "Config file path")
}
