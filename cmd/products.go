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

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage in-app products",
	Long:  "List, view, create, update, and delete in-app products (managed products).",
}

var productsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all in-app products",
	Long:  "Retrieve all in-app products for the specified package.",
	RunE:  runProductsList,
}

var productsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an in-app product",
	Long:  "Retrieve a single in-app product by its SKU.",
	RunE:  runProductsGet,
}

var productsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an in-app product",
	Long:  "Create a new in-app product from a JSON configuration file.",
	RunE:  runProductsCreate,
}

var productsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an in-app product",
	Long:  "Update an existing in-app product from a JSON configuration file.",
	RunE:  runProductsUpdate,
}

var productsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an in-app product",
	Long:  "Delete an in-app product by its SKU.",
	RunE:  runProductsDelete,
}

func init() {
	productsGetCmd.Flags().String("sku", "", "in-app product SKU")
	_ = productsGetCmd.MarkFlagRequired("sku")

	productsCreateCmd.Flags().String("config", "", "path to JSON configuration file")
	_ = productsCreateCmd.MarkFlagRequired("config")

	productsUpdateCmd.Flags().String("sku", "", "in-app product SKU")
	productsUpdateCmd.Flags().String("config", "", "path to JSON configuration file")
	_ = productsUpdateCmd.MarkFlagRequired("sku")
	_ = productsUpdateCmd.MarkFlagRequired("config")

	productsDeleteCmd.Flags().String("sku", "", "in-app product SKU")
	_ = productsDeleteCmd.MarkFlagRequired("sku")

	productsCmd.AddCommand(productsListCmd)
	productsCmd.AddCommand(productsGetCmd)
	productsCmd.AddCommand(productsCreateCmd)
	productsCmd.AddCommand(productsUpdateCmd)
	productsCmd.AddCommand(productsDeleteCmd)
	rootCmd.AddCommand(productsCmd)
}

func runProductsList(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	resp, err := svc.Inappproducts.List(pkg).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing in-app products: %w", err)
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(productsToTable(resp.Inappproduct))
	}
	return formatter.Format(resp)
}

func runProductsGet(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	sku, _ := cmd.Flags().GetString("sku")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	product, err := svc.Inappproducts.Get(pkg, sku).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting in-app product %q: %w", sku, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(productsToTable([]*androidpublisher.InAppProduct{product}))
	}
	return formatter.Format(product)
}

func runProductsCreate(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	configPath, _ := cmd.Flags().GetString("config")

	product, err := loadProductConfig(configPath)
	if err != nil {
		return err
	}
	product.PackageName = pkg

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	created, err := svc.Inappproducts.Insert(pkg, product).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating in-app product: %w", err)
	}

	formatter := output.New(getFormatter(), os.Stdout)
	return formatter.Format(created)
}

func runProductsUpdate(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	sku, _ := cmd.Flags().GetString("sku")
	configPath, _ := cmd.Flags().GetString("config")

	product, err := loadProductConfig(configPath)
	if err != nil {
		return err
	}
	product.PackageName = pkg
	product.Sku = sku

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	updated, err := svc.Inappproducts.Update(pkg, sku, product).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating in-app product %q: %w", sku, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)
	return formatter.Format(updated)
}

func runProductsDelete(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	sku, _ := cmd.Flags().GetString("sku")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	err = svc.Inappproducts.Delete(pkg, sku).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting in-app product %q: %w", sku, err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "In-app product %q deleted successfully.\n", sku)
	}

	return nil
}

func loadProductConfig(path string) (*androidpublisher.InAppProduct, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var product androidpublisher.InAppProduct
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, fmt.Errorf("parsing product config: %w", err)
	}

	return &product, nil
}

func productsToTable(products []*androidpublisher.InAppProduct) output.TableData {
	headers := []string{"SKU", "Status", "Purchase Type", "Default Price"}
	var rows [][]string

	for _, p := range products {
		defaultPrice := ""
		if p.DefaultPrice != nil {
			defaultPrice = p.DefaultPrice.Currency + " " + p.DefaultPrice.PriceMicros
		}
		rows = append(rows, []string{p.Sku, p.Status, p.PurchaseType, defaultPrice})
	}

	return output.TableData{Headers: headers, Rows: rows}
}
