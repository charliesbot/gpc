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

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage store listing images",
	Long:  "List, upload, and delete images for store listings.",
}

var imagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List images for a listing",
	Long:  "List all images for the specified language and image type.",
	RunE:  runImagesList,
}

var imagesUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an image",
	Long:  "Upload an image for the specified language and image type.",
	RunE:  runImagesUpload,
}

var imagesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an image",
	Long:  "Delete a specific image by ID for the specified language and image type.",
	RunE:  runImagesDelete,
}

var (
	imagesLanguage  string
	imagesImageType string
	imagesFile      string
	imagesImageID   string
)

func init() {
	imagesListCmd.Flags().StringVar(&imagesLanguage, "language", "", "BCP-47 language code (e.g. en-US)")
	imagesListCmd.Flags().StringVar(&imagesImageType, "image-type", "", "image type (e.g. phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvScreenshots, wearScreenshots, icon, featureGraphic, tvBanner)")
	_ = imagesListCmd.MarkFlagRequired("language")
	_ = imagesListCmd.MarkFlagRequired("image-type")

	imagesUploadCmd.Flags().StringVar(&imagesLanguage, "language", "", "BCP-47 language code (e.g. en-US)")
	imagesUploadCmd.Flags().StringVar(&imagesImageType, "image-type", "", "image type (e.g. phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvScreenshots, wearScreenshots, icon, featureGraphic, tvBanner)")
	imagesUploadCmd.Flags().StringVar(&imagesFile, "file", "", "path to the image file to upload")
	_ = imagesUploadCmd.MarkFlagRequired("language")
	_ = imagesUploadCmd.MarkFlagRequired("image-type")
	_ = imagesUploadCmd.MarkFlagRequired("file")

	imagesDeleteCmd.Flags().StringVar(&imagesLanguage, "language", "", "BCP-47 language code (e.g. en-US)")
	imagesDeleteCmd.Flags().StringVar(&imagesImageType, "image-type", "", "image type (e.g. phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvScreenshots, wearScreenshots, icon, featureGraphic, tvBanner)")
	imagesDeleteCmd.Flags().StringVar(&imagesImageID, "image-id", "", "ID of the image to delete")
	_ = imagesDeleteCmd.MarkFlagRequired("language")
	_ = imagesDeleteCmd.MarkFlagRequired("image-type")
	_ = imagesDeleteCmd.MarkFlagRequired("image-id")

	imagesCmd.AddCommand(imagesListCmd)
	imagesCmd.AddCommand(imagesUploadCmd)
	imagesCmd.AddCommand(imagesDeleteCmd)
	rootCmd.AddCommand(imagesCmd)
}

func newImagesService() (context.Context, *androidpublisher.Service, string, error) {
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

func runImagesList(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newImagesService()
	if err != nil {
		return err
	}

	var result any

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Images.List(pkg, editID, imagesLanguage, imagesImageType).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing images for language %q, type %q: %w", imagesLanguage, imagesImageType, err)
		}

		format := getFormatter()
		if format == "table" {
			rows := make([][]string, 0, len(resp.Images))
			for _, img := range resp.Images {
				rows = append(rows, []string{img.Id, img.Url})
			}
			result = output.TableData{
				Headers: []string{"ID", "URL"},
				Rows:    rows,
			}
		} else {
			result = resp.Images
		}
		return nil
	}); err != nil {
		return err
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}

func runImagesUpload(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newImagesService()
	if err != nil {
		return err
	}

	file, err := os.Open(imagesFile)
	if err != nil {
		return fmt.Errorf("opening image file %q: %w", imagesFile, err)
	}
	defer file.Close()

	var result any

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Images.Upload(pkg, editID, imagesLanguage, imagesImageType).Media(file).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("uploading image for language %q, type %q: %w", imagesLanguage, imagesImageType, err)
		}
		result = resp.Image
		return nil
	}); err != nil {
		return err
	}

	f := output.New(getFormatter(), os.Stdout)
	return f.Format(result)
}

func runImagesDelete(cmd *cobra.Command, args []string) error {
	ctx, svc, pkg, err := newImagesService()
	if err != nil {
		return err
	}

	if err := edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		if err := svc.Edits.Images.Delete(pkg, editID, imagesLanguage, imagesImageType, imagesImageID).Context(ctx).Do(); err != nil {
			return fmt.Errorf("deleting image %q for language %q, type %q: %w", imagesImageID, imagesLanguage, imagesImageType, err)
		}
		return nil
	}); err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Image %q deleted successfully.\n", imagesImageID)
	}
	return nil
}
