package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO: Implement reports subcommands using a GCS client. Financial and other
// Play Console reports are stored in Google Cloud Storage buckets and require
// the Cloud Storage API to list and download.

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Access Play Console reports",
	Long:  "List and download financial and other reports from Google Play Console via GCS.",
}

var reportsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available reports",
	Long:  "List available reports in the Play Console GCS bucket.",
	RunE:  runReportsStub,
}

var reportsDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a report",
	Long:  "Download a specific report from the Play Console GCS bucket.",
	RunE:  runReportsStub,
}

func init() {
	reportsCmd.AddCommand(reportsListCmd)
	reportsCmd.AddCommand(reportsDownloadCmd)
	rootCmd.AddCommand(reportsCmd)
}

func runReportsStub(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not yet implemented — requires GCS client for financial reports")
}
