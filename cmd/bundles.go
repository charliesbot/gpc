package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/charliesbot/gpc/internal/auth"
	"github.com/charliesbot/gpc/internal/client"
	"github.com/charliesbot/gpc/internal/edits"
	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
)

var bundlesCmd = &cobra.Command{
	Use:   "bundles",
	Short: "Manage Android App Bundles",
	Long:  "List and upload Android App Bundles (AAB) for your application.",
}

var bundlesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bundles",
	Long:  "List all uploaded Android App Bundles for the specified package.",
	RunE:  runBundlesList,
}

var bundlesUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an App Bundle",
	Long:  "Upload an Android App Bundle (AAB) file within an edit.",
	RunE:  runBundlesUpload,
}

var bundlesUploadFile string

func init() {
	bundlesUploadCmd.Flags().StringVar(&bundlesUploadFile, "file", "", "path to the AAB file to upload")
	_ = bundlesUploadCmd.MarkFlagRequired("file")

	bundlesCmd.AddCommand(bundlesListCmd)
	bundlesCmd.AddCommand(bundlesUploadCmd)
	rootCmd.AddCommand(bundlesCmd)
}

func runBundlesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile: flagKeyFile,
	})
	if err != nil {
		return fmt.Errorf("resolving credentials: %w", err)
	}

	svc, err := client.NewService(ctx, creds)
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	var bundles []*androidpublisher.Bundle

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Bundles.List(pkg, editID).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing bundles: %w", err)
		}
		bundles = resp.Bundles
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		rows := make([][]string, 0, len(bundles))
		for _, b := range bundles {
			rows = append(rows, []string{
				strconv.FormatInt(b.VersionCode, 10),
				b.Sha256,
			})
		}
		return formatter.Format(output.TableData{
			Headers: []string{"Version Code", "SHA-256"},
			Rows:    rows,
		})
	}

	return formatter.Format(bundles)
}

func runBundlesUpload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile: flagKeyFile,
	})
	if err != nil {
		return fmt.Errorf("resolving credentials: %w", err)
	}

	svc, err := client.NewService(ctx, creds)
	if err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	f, err := os.Open(bundlesUploadFile)
	if err != nil {
		return fmt.Errorf("opening bundle file %q: %w", bundlesUploadFile, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("reading file info: %w", err)
	}

	var bundle *androidpublisher.Bundle

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		bar := output.NewProgressBar(fi.Size(), "Uploading bundle...")
		reader := io.TeeReader(f, bar)

		upload := svc.Edits.Bundles.Upload(pkg, editID)
		upload.Media(reader)

		resp, err := upload.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("uploading bundle: %w", err)
		}

		bar.Finish()
		bundle = resp
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(output.TableData{
			Headers: []string{"Version Code", "SHA-256"},
			Rows: [][]string{
				{
					strconv.FormatInt(bundle.VersionCode, 10),
					bundle.Sha256,
				},
			},
		})
	}

	return formatter.Format(bundle)
}
