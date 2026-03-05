package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO: Implement grants subcommands using the Play Developer API grants
// endpoint. Grants control the permissions assigned to users on a per-app
// or account-wide basis.

var grantsCmd = &cobra.Command{
	Use:   "grants",
	Short: "Manage user grants and permissions",
	Long:  "Create, update, and delete permission grants for developer account users.",
}

var grantsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a grant",
	Long:  "Create a new permission grant for a user.",
	RunE:  runGrantsStub,
}

var grantsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a grant",
	Long:  "Update an existing permission grant for a user.",
	RunE:  runGrantsStub,
}

var grantsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a grant",
	Long:  "Delete a permission grant from a user.",
	RunE:  runGrantsStub,
}

func init() {
	grantsCmd.AddCommand(grantsCreateCmd)
	grantsCmd.AddCommand(grantsUpdateCmd)
	grantsCmd.AddCommand(grantsDeleteCmd)
	rootCmd.AddCommand(grantsCmd)
}

func runGrantsStub(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not yet implemented — requires Play Developer API grants endpoint")
}
