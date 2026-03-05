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
	"google.golang.org/api/androidpublisher/v3"
)

var listingsCmd = &cobra.Command{
	Use:   "listings",
	Short: "Manage store listings",
	Long:  "List, view, update, and delete store listings for an application.",
}

var listingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all store listings",
	Long:  "List all store listings for the specified application.",
	RunE:  runListingsList,
}

var listingsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a store listing by language",
	Long:  "Get a specific store listing for the given language code.",
	RunE:  runListingsGet,
}

var listingsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a store listing",
	Long:  "Update a store listing for the given language code with new title, short description, or full description.",
	RunE:  runListingsUpdate,
}

var listingsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a store listing",
	Long:  "Delete the store listing for the given language code.",
	RunE:  runListingsDelete,
}

var (
	listingsLanguage  string
	listingsTitle     string
	listingsShortDesc string
	listingsFullDesc  string
)

func init() {
	listingsGetCmd.Flags().StringVar(&listingsLanguage, "language", "", "BCP-47 language code (e.g. en-US)")
	_ = listingsGetCmd.MarkFlagRequired("language")

	listingsUpdateCmd.Flags().StringVar(&listingsLanguage, "language", "", "BCP-47 language code (e.g. en-US)")
	listingsUpdateCmd.Flags().StringVar(&listingsTitle, "title", "", "listing title")
	listingsUpdateCmd.Flags().StringVar(&listingsShortDesc, "short-desc", "", "short description")
	listingsUpdateCmd.Flags().StringVar(&listingsFullDesc, "full-desc", "", "full description")
	_ = listingsUpdateCmd.MarkFlagRequired("language")

	listingsDeleteCmd.Flags().StringVar(&listingsLanguage, "language", "", "BCP-47 language code (e.g. en-US)")
	_ = listingsDeleteCmd.MarkFlagRequired("language")

	listingsCmd.AddCommand(listingsListCmd)
	listingsCmd.AddCommand(listingsGetCmd)
	listingsCmd.AddCommand(listingsUpdateCmd)
	listingsCmd.AddCommand(listingsDeleteCmd)
	rootCmd.AddCommand(listingsCmd)
}

func newListingsService() (context.Context, *androidpublisher.Service, string, error) {
	pkg, err := getPackage()
	if err != nil {
		return nil, nil, "", err
	}

	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile:    flagKeyFile,
		UseKeyring: true,
	})
	if err != nil {
		return nil, nil, "", fmt.Errorf("resolving credentials: %w", err)
	}

	ctx := context.Background()
	svc, err := client.NewService(ctx, creds)
	if err != nil {
		return nil, nil, "", fmt.Errorf("creating API client: %w", err)
	}

	return ctx, svc, pkg, nil
}

func runListingsList(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newListingsService()
	if err != nil {
		return err
	}

	var result any

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Listings.List(pkg, editID).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing store listings: %w", err)
		}

		format := getFormatter()
		if format == "table" {
			rows := make([][]string, 0, len(resp.Listings))
			for _, l := range resp.Listings {
				rows = append(rows, []string{l.Language, l.Title, l.ShortDescription})
			}
			result = output.TableData{
				Headers: []string{"Language", "Title", "Short Description"},
				Rows:    rows,
			}
		} else {
			result = resp.Listings
		}
		return nil
	}); err != nil {
		return err
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}

func runListingsGet(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newListingsService()
	if err != nil {
		return err
	}

	var result any

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		listing, err := svc.Edits.Listings.Get(pkg, editID, listingsLanguage).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting listing for language %q: %w", listingsLanguage, err)
		}
		result = listing
		return nil
	}); err != nil {
		return err
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}

func runListingsUpdate(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newListingsService()
	if err != nil {
		return err
	}

	listing := &androidpublisher.Listing{
		Language: listingsLanguage,
	}
	if listingsTitle != "" {
		listing.Title = listingsTitle
	}
	if listingsShortDesc != "" {
		listing.ShortDescription = listingsShortDesc
	}
	if listingsFullDesc != "" {
		listing.FullDescription = listingsFullDesc
	}

	var result any

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		updated, err := svc.Edits.Listings.Update(pkg, editID, listingsLanguage, listing).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating listing for language %q: %w", listingsLanguage, err)
		}
		result = updated
		return nil
	}); err != nil {
		return err
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}

func runListingsDelete(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newListingsService()
	if err != nil {
		return err
	}

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		if err := svc.Edits.Listings.Delete(pkg, editID, listingsLanguage).Context(ctx).Do(); err != nil {
			return fmt.Errorf("deleting listing for language %q: %w", listingsLanguage, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Listing for language %q deleted successfully.\n", listingsLanguage)
	}
	return nil
}
