package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charliesbot/gpc/internal/auth"
	"github.com/charliesbot/gpc/internal/client"
	"github.com/charliesbot/gpc/internal/edits"
	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
)

var tracksCmd = &cobra.Command{
	Use:   "tracks",
	Short: "Manage release tracks",
	Long:  "List, get details of, and update release tracks for your application.",
}

var tracksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracks",
	Long:  "List all available release tracks for the specified package.",
	RunE:  runTracksList,
}

var tracksGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific track",
	Long:  "Get details of a specific release track by name.",
	RunE:  runTracksGet,
}

var tracksUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a track",
	Long:  "Update a release track with a new version code, rollout fraction, and status.",
	RunE:  runTracksUpdate,
}

var (
	trackName        string
	trackVersionCode int64
	trackRollout     float64
	trackStatus      string
)

func init() {
	tracksGetCmd.Flags().StringVar(&trackName, "track", "", "track name (e.g. production, beta, alpha, internal)")
	_ = tracksGetCmd.MarkFlagRequired("track")

	tracksUpdateCmd.Flags().StringVar(&trackName, "track", "", "track name (e.g. production, beta, alpha, internal)")
	tracksUpdateCmd.Flags().Int64Var(&trackVersionCode, "version-code", 0, "version code to release")
	tracksUpdateCmd.Flags().Float64Var(&trackRollout, "rollout", 1.0, "rollout fraction (0.0 to 1.0)")
	tracksUpdateCmd.Flags().StringVar(&trackStatus, "status", "completed", "release status (completed, draft, halted, inProgress)")
	_ = tracksUpdateCmd.MarkFlagRequired("track")
	_ = tracksUpdateCmd.MarkFlagRequired("version-code")

	tracksCmd.AddCommand(tracksListCmd)
	tracksCmd.AddCommand(tracksGetCmd)
	tracksCmd.AddCommand(tracksUpdateCmd)
	rootCmd.AddCommand(tracksCmd)
}

func runTracksList(cmd *cobra.Command, args []string) error {
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

	var tracks []*androidpublisher.Track

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Tracks.List(pkg, editID).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing tracks: %w", err)
		}
		tracks = resp.Tracks
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		rows := make([][]string, 0, len(tracks))
		for _, t := range tracks {
			rows = append(rows, []string{
				t.Track,
				formatReleases(t.Releases),
			})
		}
		return formatter.Format(output.TableData{
			Headers: []string{"Track", "Releases"},
			Rows:    rows,
		})
	}

	return formatter.Format(tracks)
}

func runTracksGet(cmd *cobra.Command, args []string) error {
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

	var track *androidpublisher.Track

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		resp, err := svc.Edits.Tracks.Get(pkg, editID, trackName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting track %q: %w", trackName, err)
		}
		track = resp
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		rows := make([][]string, 0, len(track.Releases))
		for _, r := range track.Releases {
			versionCodes := make([]string, 0, len(r.VersionCodes))
			for _, vc := range r.VersionCodes {
				versionCodes = append(versionCodes, strconv.FormatInt(vc, 10))
			}
			rows = append(rows, []string{
				r.Name,
				r.Status,
				strings.Join(versionCodes, ", "),
				fmt.Sprintf("%.2f", r.UserFraction),
			})
		}
		return formatter.Format(output.TableData{
			Headers: []string{"Name", "Status", "Version Codes", "Rollout"},
			Rows:    rows,
		})
	}

	return formatter.Format(track)
}

func runTracksUpdate(cmd *cobra.Command, args []string) error {
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

	var track *androidpublisher.Track

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		release := &androidpublisher.TrackRelease{
			VersionCodes: []int64{trackVersionCode},
			Status:       trackStatus,
		}

		if trackStatus == "inProgress" {
			release.UserFraction = trackRollout
		}

		trackUpdate := &androidpublisher.Track{
			Track: trackName,
			Releases: []*androidpublisher.TrackRelease{
				release,
			},
		}

		resp, err := svc.Edits.Tracks.Update(pkg, editID, trackName, trackUpdate).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating track %q: %w", trackName, err)
		}
		track = resp
		return nil
	})
	if err != nil {
		return err
	}

	formatter := output.New(getFormatter(), os.Stdout)

	if getFormatter() == "table" {
		rows := make([][]string, 0, len(track.Releases))
		for _, r := range track.Releases {
			versionCodes := make([]string, 0, len(r.VersionCodes))
			for _, vc := range r.VersionCodes {
				versionCodes = append(versionCodes, strconv.FormatInt(vc, 10))
			}
			rows = append(rows, []string{
				r.Name,
				r.Status,
				strings.Join(versionCodes, ", "),
				fmt.Sprintf("%.2f", r.UserFraction),
			})
		}
		return formatter.Format(output.TableData{
			Headers: []string{"Name", "Status", "Version Codes", "Rollout"},
			Rows:    rows,
		})
	}

	return formatter.Format(track)
}

// formatReleases produces a compact summary of releases for table display.
func formatReleases(releases []*androidpublisher.TrackRelease) string {
	if len(releases) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(releases))
	for _, r := range releases {
		parts = append(parts, fmt.Sprintf("%s(%s)", r.Status, r.Name))
	}
	return strings.Join(parts, ", ")
}
