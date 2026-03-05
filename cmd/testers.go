package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charliesbot/gpc/internal/edits"
	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
)

var testersCmd = &cobra.Command{
	Use:   "testers",
	Short: "Manage testers for testing tracks",
	Long:  "Get and update tester lists for app testing tracks (e.g. alpha, beta, internal).",
}

var testersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get testers for a track",
	Long:  "Retrieve the tester configuration for the specified testing track.",
	RunE:  runTestersGet,
}

var testersUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update testers for a track",
	Long:  "Set the list of tester email addresses for the specified testing track.",
	RunE:  runTestersUpdate,
}

func init() {
	testersGetCmd.Flags().String("track", "", "testing track name (e.g. alpha, beta, internal)")
	_ = testersGetCmd.MarkFlagRequired("track")

	testersUpdateCmd.Flags().String("track", "", "testing track name (e.g. alpha, beta, internal)")
	testersUpdateCmd.Flags().String("emails", "", "comma-separated list of tester email addresses")
	_ = testersUpdateCmd.MarkFlagRequired("track")
	_ = testersUpdateCmd.MarkFlagRequired("emails")

	testersCmd.AddCommand(testersGetCmd)
	testersCmd.AddCommand(testersUpdateCmd)
	rootCmd.AddCommand(testersCmd)
}

func runTestersGet(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	track, _ := cmd.Flags().GetString("track")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	var testers *androidpublisher.Testers

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		t, getErr := svc.Edits.Testers.Get(pkg, editID, track).Context(ctx).Do()
		if getErr != nil {
			return fmt.Errorf("getting testers for track %q: %w", track, getErr)
		}
		testers = t
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(testersToTable(track, testers))
	}
	return formatter.Format(testers)
}

func runTestersUpdate(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	track, _ := cmd.Flags().GetString("track")
	emailsRaw, _ := cmd.Flags().GetString("emails")

	emails := parseEmails(emailsRaw)
	if len(emails) == 0 {
		return fmt.Errorf("at least one email address is required")
	}

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	var updated *androidpublisher.Testers

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		testers := &androidpublisher.Testers{
			GoogleGroups: emails,
		}

		t, updateErr := svc.Edits.Testers.Update(pkg, editID, track, testers).Context(ctx).Do()
		if updateErr != nil {
			return fmt.Errorf("updating testers for track %q: %w", track, updateErr)
		}
		updated = t
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(testersToTable(track, updated))
	}
	return formatter.Format(updated)
}

func parseEmails(raw string) []string {
	var emails []string
	for _, e := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(e)
		if trimmed != "" {
			emails = append(emails, trimmed)
		}
	}
	return emails
}

func testersToTable(track string, testers *androidpublisher.Testers) output.TableData {
	headers := []string{"Track", "Email"}
	var rows [][]string

	if testers != nil {
		for _, email := range testers.GoogleGroups {
			rows = append(rows, []string{track, email})
		}
	}

	return output.TableData{Headers: headers, Rows: rows}
}
