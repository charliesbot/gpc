package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
)

var purchasesCmd = &cobra.Command{
	Use:   "purchases",
	Short: "Manage purchases and subscriptions",
	Long:  "Query and manage in-app product purchases and subscriptions.",
}

var purchasesProductGetCmd = &cobra.Command{
	Use:   "product-get",
	Short: "Get an in-app product purchase",
	Long:  "Retrieve the status of an in-app product purchase by SKU and purchase token.",
	RunE:  runPurchasesProductGet,
}

var purchasesSubGetCmd = &cobra.Command{
	Use:   "sub-get",
	Short: "Get a subscription purchase",
	Long:  "Retrieve the status of a subscription purchase by subscription ID and purchase token.",
	RunE:  runPurchasesSubGet,
}

var purchasesSubCancelCmd = &cobra.Command{
	Use:   "sub-cancel",
	Short: "Cancel a subscription purchase",
	Long:  "Cancel a user's subscription by subscription ID and purchase token.",
	RunE:  runPurchasesSubCancel,
}

func init() {
	purchasesProductGetCmd.Flags().String("sku", "", "in-app product SKU")
	_ = purchasesProductGetCmd.MarkFlagRequired("sku")
	purchasesProductGetCmd.Flags().String("token", "", "purchase token")
	_ = purchasesProductGetCmd.MarkFlagRequired("token")

	purchasesSubGetCmd.Flags().String("sub-id", "", "subscription ID")
	_ = purchasesSubGetCmd.MarkFlagRequired("sub-id")
	purchasesSubGetCmd.Flags().String("token", "", "purchase token")
	_ = purchasesSubGetCmd.MarkFlagRequired("token")

	purchasesSubCancelCmd.Flags().String("sub-id", "", "subscription ID")
	_ = purchasesSubCancelCmd.MarkFlagRequired("sub-id")
	purchasesSubCancelCmd.Flags().String("token", "", "purchase token")
	_ = purchasesSubCancelCmd.MarkFlagRequired("token")

	purchasesCmd.AddCommand(purchasesProductGetCmd)
	purchasesCmd.AddCommand(purchasesSubGetCmd)
	purchasesCmd.AddCommand(purchasesSubCancelCmd)
	rootCmd.AddCommand(purchasesCmd)
}

func runPurchasesProductGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	sku, _ := cmd.Flags().GetString("sku")
	token, _ := cmd.Flags().GetString("token")

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	purchase, err := svc.Purchases.Products.Get(pkg, sku, token).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting product purchase: %w", err)
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(purchase)
}

func runPurchasesSubGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	subID, _ := cmd.Flags().GetString("sub-id")
	token, _ := cmd.Flags().GetString("token")

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	subscription, err := svc.Purchases.Subscriptions.Get(pkg, subID, token).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting subscription purchase: %w", err)
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(subscription)
}

func runPurchasesSubCancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	subID, _ := cmd.Flags().GetString("sub-id")
	token, _ := cmd.Flags().GetString("token")

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	err = svc.Purchases.Subscriptions.Cancel(pkg, subID, token).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling subscription: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Subscription %s cancelled successfully.\n", subID)
	}

	return nil
}
