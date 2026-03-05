package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "Manage orders",
	Long:  "Query and manage orders including refunds.",
}

var ordersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an order",
	Long:  "Retrieve order details by order ID.",
	RunE:  runOrdersGet,
}

var ordersRefundCmd = &cobra.Command{
	Use:   "refund",
	Short: "Refund an order",
	Long:  "Issue a refund for an order, optionally revoking access.",
	RunE:  runOrdersRefund,
}

func init() {
	ordersGetCmd.Flags().String("order-id", "", "order ID")
	_ = ordersGetCmd.MarkFlagRequired("order-id")

	ordersRefundCmd.Flags().String("order-id", "", "order ID")
	_ = ordersRefundCmd.MarkFlagRequired("order-id")
	ordersRefundCmd.Flags().Bool("revoke", false, "revoke the subscription or in-app purchase in addition to refunding")

	ordersCmd.AddCommand(ordersGetCmd)
	ordersCmd.AddCommand(ordersRefundCmd)
	rootCmd.AddCommand(ordersCmd)
}

func runOrdersGet(cmd *cobra.Command, args []string) error {
	// The Play Developer API does not expose a direct orders.get endpoint.
	// Order information must be retrieved through the purchases endpoints.
	fmt.Fprintln(os.Stderr, "The Play Developer API does not have a direct orders.get endpoint.")
	fmt.Fprintln(os.Stderr, "Use purchases commands to verify purchase state:")
	fmt.Fprintln(os.Stderr, "  gpc purchases product-get --sku <sku> --token <token>")
	fmt.Fprintln(os.Stderr, "  gpc purchases sub-get --sub-id <id> --token <token>")
	return nil
}

func runOrdersRefund(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	orderID, _ := cmd.Flags().GetString("order-id")
	revoke, _ := cmd.Flags().GetBool("revoke")

	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	call := svc.Orders.Refund(pkg, orderID)
	if revoke {
		call = call.Revoke(true)
	}

	err = call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("refunding order: %w", err)
	}

	if !flagQuiet {
		msg := fmt.Sprintf("Order %s refunded successfully.", orderID)
		if revoke {
			msg = fmt.Sprintf("Order %s refunded and revoked successfully.", orderID)
		}
		fmt.Fprintln(os.Stderr, msg)
	}

	return nil
}
