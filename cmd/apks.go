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

var apksCmd = &cobra.Command{
	Use:   "apks",
	Short: "Manage APKs",
	Long:  "List and upload APKs for your application.",
}

var apksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all APKs",
	Long:  "List all uploaded APKs for the specified package.",
	RunE:  runApksList,
}

var apksUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an APK",
	Long:  "Upload an APK file within an edit.",
	RunE:  runApksUpload,
}

var apksUploadFile string

func init() {
	apksUploadCmd.Flags().StringVar(&apksUploadFile, "file", "", "path to the APK file to upload")
	_ = apksUploadCmd.MarkFlagRequired("file")

	apksCmd.AddCommand(apksListCmd)
	apksCmd.AddCommand(apksUploadCmd)
	rootCmd.AddCommand(apksCmd)
}

func runApksList(cmd *cobra.Command, args []string) error {
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

	var apks []*androidpublisher.Apk

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Apks.List(pkg, editID).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing APKs: %w", err)
		}
		apks = resp.Apks
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		rows := make([][]string, 0, len(apks))
		for _, a := range apks {
			rows = append(rows, []string{
				strconv.FormatInt(int64(a.VersionCode), 10),
				a.Binary.Sha256,
			})
		}
		return formatter.Format(output.TableData{
			Headers: []string{"Version Code", "SHA-256"},
			Rows:    rows,
		})
	}

	return formatter.Format(apks)
}

func runApksUpload(cmd *cobra.Command, args []string) error {
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

	f, err := os.Open(apksUploadFile)
	if err != nil {
		return fmt.Errorf("opening APK file %q: %w", apksUploadFile, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("reading file info: %w", err)
	}

	var apk *androidpublisher.Apk

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		bar := output.NewProgressBar(fi.Size(), "Uploading APK...")
		reader := io.TeeReader(f, bar)

		upload := svc.Edits.Apks.Upload(pkg, editID)
		upload.Media(reader)

		resp, err := upload.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("uploading APK: %w", err)
		}

		bar.Finish()
		apk = resp
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
					strconv.FormatInt(int64(apk.VersionCode), 10),
					apk.Binary.Sha256,
				},
			},
		})
	}

	return formatter.Format(apk)
}
