package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charliesbot/gpc/internal/auth"
	"github.com/charliesbot/gpc/internal/client"
	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
)

var reviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Manage app reviews",
	Long:  "List, view, and reply to user reviews on Google Play.",
}

var reviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List reviews for an app",
	Long:  "Retrieve all reviews for the specified package.",
	RunE:  runReviewsList,
}

var reviewsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific review",
	Long:  "Retrieve a single review by its review ID.",
	RunE:  runReviewsGet,
}

var reviewsReplyCmd = &cobra.Command{
	Use:   "reply",
	Short: "Reply to a review",
	Long:  "Send a developer reply to a specific review.",
	RunE:  runReviewsReply,
}

func init() {
	reviewsGetCmd.Flags().String("review-id", "", "review ID to retrieve")
	_ = reviewsGetCmd.MarkFlagRequired("review-id")

	reviewsReplyCmd.Flags().String("review-id", "", "review ID to reply to")
	reviewsReplyCmd.Flags().String("text", "", "reply text")
	_ = reviewsReplyCmd.MarkFlagRequired("review-id")
	_ = reviewsReplyCmd.MarkFlagRequired("text")

	reviewsCmd.AddCommand(reviewsListCmd)
	reviewsCmd.AddCommand(reviewsGetCmd)
	reviewsCmd.AddCommand(reviewsReplyCmd)
	rootCmd.AddCommand(reviewsCmd)
}

func runReviewsList(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	resp, err := svc.Reviews.List(pkg).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing reviews: %w", err)
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(reviewsToTable(resp.Reviews))
	}
	return formatter.Format(resp)
}

func runReviewsGet(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	reviewID, _ := cmd.Flags().GetString("review-id")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	review, err := svc.Reviews.Get(pkg, reviewID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting review %q: %w", reviewID, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		return formatter.Format(reviewsToTable([]*androidpublisher.Review{review}))
	}
	return formatter.Format(review)
}

func runReviewsReply(cmd *cobra.Command, args []string) error {
	pkg, err := getPackage()
	if err != nil {
		return err
	}

	reviewID, _ := cmd.Flags().GetString("review-id")
	text, _ := cmd.Flags().GetString("text")

	ctx := context.Background()
	svc, err := buildService(ctx)
	if err != nil {
		return err
	}

	req := &androidpublisher.ReviewsReplyRequest{
		ReplyText: text,
	}

	resp, err := svc.Reviews.Reply(pkg, reviewID, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replying to review %q: %w", reviewID, err)
	}

	formatter := output.New(getFormatter(), os.Stdout)
	return formatter.Format(resp)
}

func reviewsToTable(reviews []*androidpublisher.Review) output.TableData {
	headers := []string{"Review ID", "Author", "Rating", "Comment"}
	var rows [][]string

	for _, r := range reviews {
		author := ""
		rating := ""
		comment := ""

		if len(r.Comments) > 0 {
			uc := r.Comments[0].UserComment
			if uc != nil {
				author = uc.ReviewerLanguage
				rating = fmt.Sprintf("%d", uc.StarRating)
				comment = uc.Text
			}
		}

		rows = append(rows, []string{r.ReviewId, author, rating, comment})
	}

	return output.TableData{Headers: headers, Rows: rows}
}

func buildService(ctx context.Context) (*androidpublisher.Service, error) {
	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile:    flagKeyFile,
		UseKeyring: true,
	})
	if err != nil {
		return nil, fmt.Errorf("resolving credentials: %w", err)
	}

	svc, err := client.NewService(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}

	return svc, nil
}
