package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
)

var editsCmd = &cobra.Command{
	Use:   "edits",
	Short: "Manage edits directly",
	Long:  "Low-level escape hatch to insert, validate, and commit edits manually.",
}

var editsInsertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Create a new edit",
	Long:  "Insert a new edit for the specified package and print the edit ID.",
	RunE:  runEditsInsert,
}

var editsValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate an edit",
	Long:  "Validate an existing edit by its edit ID.",
	RunE:  runEditsValidate,
}

var editsCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit an edit",
	Long:  "Commit an existing edit by its edit ID, making it live.",
	RunE:  runEditsCommit,
}

func init() {
	editsValidateCmd.Flags().String("edit-id", "", "edit ID to validate")
	_ = editsValidateCmd.MarkFlagRequired("edit-id")

	editsCommitCmd.Flags().String("edit-id", "", "edit ID to commit")
	_ = editsCommitCmd.MarkFlagRequired("edit-id")

	editsCmd.AddCommand(editsInsertCmd)
	editsCmd.AddCommand(editsValidateCmd)
	editsCmd.AddCommand(editsCommitCmd)
	rootCmd.AddCommand(editsCmd)
}

func runEditsInsert(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	edit, err := svc.Edits.Insert(pkg, nil).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("inserting edit: %w", err)
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(edit)
}

func runEditsValidate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	editID, _ := cmd.Flags().GetString("edit-id")

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	result, err := svc.Edits.Validate(pkg, editID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating edit: %w", err)
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}

func runEditsCommit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	editID, _ := cmd.Flags().GetString("edit-id")

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	result, err := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing edit: %w", err)
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}
