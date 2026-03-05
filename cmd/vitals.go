package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO: Implement vitals subcommands using the Play Developer Reporting API
// (playdeveloperreporting.googleapis.com). This requires a separate API client
// from the androidpublisher service used elsewhere in this CLI.

var vitalsCmd = &cobra.Command{
	Use:   "vitals",
	Short: "View app vitals and performance metrics",
	Long:  "Access crash rates, ANR rates, rendering metrics, and other app vitals from the Play Developer Reporting API.",
}

var vitalsCrashRateCmd = &cobra.Command{
	Use:   "crash-rate",
	Short: "View crash rate metrics",
	Long:  "Display crash rate metrics for the application.",
	RunE:  runVitalsStub,
}

var vitalsAnrRateCmd = &cobra.Command{
	Use:   "anr-rate",
	Short: "View ANR rate metrics",
	Long:  "Display Application Not Responding (ANR) rate metrics.",
	RunE:  runVitalsStub,
}

var vitalsErrorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "View error clusters",
	Long:  "Display error clusters and their details.",
	RunE:  runVitalsStub,
}

var vitalsSlowRenderingCmd = &cobra.Command{
	Use:   "slow-rendering",
	Short: "View slow rendering metrics",
	Long:  "Display slow rendering frame rate metrics.",
	RunE:  runVitalsStub,
}

var vitalsSlowStartCmd = &cobra.Command{
	Use:   "slow-start",
	Short: "View slow start-up metrics",
	Long:  "Display cold, warm, and hot start-up time metrics.",
	RunE:  runVitalsStub,
}

var vitalsAnomaliesCmd = &cobra.Command{
	Use:   "anomalies",
	Short: "View detected anomalies",
	Long:  "Display anomalies detected in app vitals metrics.",
	RunE:  runVitalsStub,
}

func init() {
	vitalsCmd.AddCommand(vitalsCrashRateCmd)
	vitalsCmd.AddCommand(vitalsAnrRateCmd)
	vitalsCmd.AddCommand(vitalsErrorsCmd)
	vitalsCmd.AddCommand(vitalsSlowRenderingCmd)
	vitalsCmd.AddCommand(vitalsSlowStartCmd)
	vitalsCmd.AddCommand(vitalsAnomaliesCmd)
	rootCmd.AddCommand(vitalsCmd)
}

func runVitalsStub(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not yet implemented — requires Play Developer Reporting API client (playdeveloperreporting.googleapis.com)")
}
