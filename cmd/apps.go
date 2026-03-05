package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charliesbot/gpc/internal/auth"
	"github.com/charliesbot/gpc/internal/client"
	"github.com/charliesbot/gpc/internal/edits"
	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage applications",
	Long:  "List and inspect applications in your Google Play Developer account.",
}

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long:  "List all applications associated with your developer account.",
	RunE:  runAppsList,
}

var appsInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show application details",
	Long:  "Show details for a specific application using the Edits API.",
	RunE:  runAppsInfo,
}

func init() {
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsInfoCmd)
	rootCmd.AddCommand(appsCmd)
}

func runAppsList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not yet implemented — the Play Developer API does not support listing apps directly")
}

func runAppsInfo(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile:    flagKeyFile,
		UseKeyring: true,
	})
	if err != nil {
		return fmt.Errorf("resolving credentials: %w", err)
	}

	ctx := context.Background()
	svc, err := client.NewService(ctx, creds)
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	var result any

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		details, err := svc.Edits.Details.Get(pkg, editID).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("fetching app details: %w", err)
		}
		result = details
		return nil
	}); err != nil {
		return err
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}
