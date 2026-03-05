package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
)

var subscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Manage subscriptions",
	Long:  "List, view, create, update, and archive subscriptions via the Monetization API.",
}

var subscriptionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all subscriptions",
	Long:  "Retrieve all subscriptions for the specified package.",
	RunE:  runSubscriptionsList,
}

var subscriptionsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a subscription",
	Long:  "Retrieve a single subscription by its product ID.",
	RunE:  runSubscriptionsGet,
}

var subscriptionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a subscription",
	Long:  "Create a new subscription from a JSON configuration file.",
	RunE:  runSubscriptionsCreate,
}

var subscriptionsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a subscription",
	Long:  "Update an existing subscription from a JSON configuration file.",
	RunE:  runSubscriptionsUpdate,
}

var subscriptionsArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a subscription",
	Long:  "Archive an existing subscription by its product ID.",
	RunE:  runSubscriptionsArchive,
}

func init() {
	subscriptionsGetCmd.Flags().String("product-id", "", "subscription product ID")
	_ = subscriptionsGetCmd.MarkFlagRequired("product-id")

	subscriptionsCreateCmd.Flags().String("product-id", "", "subscription product ID")
	subscriptionsCreateCmd.Flags().String("config", "", "path to JSON configuration file")
	_ = subscriptionsCreateCmd.MarkFlagRequired("product-id")
	_ = subscriptionsCreateCmd.MarkFlagRequired("config")

	subscriptionsUpdateCmd.Flags().String("product-id", "", "subscription product ID")
	subscriptionsUpdateCmd.Flags().String("config", "", "path to JSON configuration file")
	_ = subscriptionsUpdateCmd.MarkFlagRequired("product-id")
	_ = subscriptionsUpdateCmd.MarkFlagRequired("config")

	subscriptionsArchiveCmd.Flags().String("product-id", "", "subscription product ID")
	_ = subscriptionsArchiveCmd.MarkFlagRequired("product-id")

	subscriptionsCmd.AddCommand(subscriptionsListCmd)
	subscriptionsCmd.AddCommand(subscriptionsGetCmd)
	subscriptionsCmd.AddCommand(subscriptionsCreateCmd)
	subscriptionsCmd.AddCommand(subscriptionsUpdateCmd)
	subscriptionsCmd.AddCommand(subscriptionsArchiveCmd)
	rootCmd.AddCommand(subscriptionsCmd)
}

func runSubscriptionsList(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	resp, err := svc.Monetization.Subscriptions.List(pkg).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing subscriptions: %w", err)
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(subscriptionsToTable(resp.Subscriptions))
	}
	return formatter.Format(resp)
}

func runSubscriptionsGet(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	productID, _ := cmd.Flags().GetString("product-id")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	sub, err := svc.Monetization.Subscriptions.Get(pkg, productID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting subscription %q: %w", productID, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(subscriptionsToTable([]*androidpublisher.Subscription{sub}))
	}
	return formatter.Format(sub)
}

func runSubscriptionsCreate(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	productID, _ := cmd.Flags().GetString("product-id")
	configPath, _ := cmd.Flags().GetString("config")

	sub, err := loadSubscriptionConfig(configPath)
	if err != nil {
		return err
	}
	sub.ProductId = productID

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	created, err := svc.Monetization.Subscriptions.Create(pkg, sub).ProductId(productID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating subscription %q: %w", productID, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)
	return formatter.Format(created)
}

func runSubscriptionsUpdate(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	productID, _ := cmd.Flags().GetString("product-id")
	configPath, _ := cmd.Flags().GetString("config")

	sub, err := loadSubscriptionConfig(configPath)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	updated, err := svc.Monetization.Subscriptions.Patch(pkg, productID, sub).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating subscription %q: %w", productID, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)
	return formatter.Format(updated)
}

func runSubscriptionsArchive(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	productID, _ := cmd.Flags().GetString("product-id")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	req := &androidpublisher.ArchiveSubscriptionRequest{}

	archived, err := svc.Monetization.Subscriptions.Archive(pkg, productID, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("archiving subscription %q: %w", productID, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)
	return formatter.Format(archived)
}

func loadSubscriptionConfig(path string) (*androidpublisher.Subscription, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var sub androidpublisher.Subscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, fmt.Errorf("parsing subscription config: %w", err)
	}

	return &sub, nil
}

func subscriptionsToTable(subs []*androidpublisher.Subscription) output.TableData {
	headers := []string{"Product ID", "Package Name", "Archived"}
	var rows [][]string

	for _, s := range subs {
		archived := "false"
		if s.Archived {
			archived = "true"
		}
		rows = append(rows, []string{s.ProductId, s.PackageName, archived})
	}

	return output.TableData{Headers: headers, Rows: rows}
}
