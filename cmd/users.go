package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO: Implement users subcommands using the Play Developer API users
// endpoint. The Users resource requires the developer account ID and operates
// on the developer-level (not per-app).

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage developer account users",
	Long:  "List, create, and delete users with access to the Play Developer account.",
}

var (
	usersDeveloperID string
)

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long:  "List all users with access to the developer account. Requires --developer-id.",
	RunE:  runUsersStub,
}

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a user",
	Long:  "Create a new user with access to the developer account.",
	RunE:  runUsersStub,
}

var usersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  "Remove a user's access from the developer account.",
	RunE:  runUsersStub,
}

func init() {
	usersListCmd.Flags().StringVar(&usersDeveloperID, "developer-id", "", "developer account ID")
	_ = usersListCmd.MarkFlagRequired("developer-id")

	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	rootCmd.AddCommand(usersCmd)
}

func runUsersStub(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not yet implemented — requires Play Developer API users endpoint")
}
