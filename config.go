package main

import (
	"fmt"
	"github.com/signalfx/sfxdatagen/internal/metrics"
	"github.com/spf13/cobra"
	"os"
)

var (
	metricsCfg *metrics.Config
)

// rootCmd is the root command on which will be run children commands
var rootCmd = &cobra.Command{
	Use:     "sfxdatagen",
	Short:   "sfxdatagen sends datapoints to ingest to simulate a load",
	Example: "sfxdatagen",
}

// metricsCmd is the command responsible for sending metrics
var metricsCmd = &cobra.Command{
	Use:     "metrics",
	Short:   fmt.Sprintf("Simulates a client generating metrics"),
	Example: "sfxdatagen metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		return metrics.Start(metricsCfg)
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)

	metricsCfg = new(metrics.Config)
	metricsCfg.Flags(metricsCmd.Flags())

	// Disabling completion command for end user
	// https://github.com/spf13/cobra/blob/master/shell_completions.md
	rootCmd.CompletionOptions.DisableDefaultCmd = true

}

// Execute tries to run the input command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
