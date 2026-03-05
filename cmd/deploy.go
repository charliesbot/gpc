package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charliesbot/gpc/internal/auth"
	"github.com/charliesbot/gpc/internal/client"
	"github.com/charliesbot/gpc/internal/edits"
	clierrors "github.com/charliesbot/gpc/internal/errors"
	"github.com/charliesbot/gpc/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"
)

// Flag variables scoped to the deploy command.
var (
	deployFile             string
	deployTrack            string
	deployRollout          float64
	deployReleaseName      string
	deployReleaseNotes     string
	deployReleaseNotesFile string
	deployStatus           string
	deployMappingFile      string
)

// validStatuses enumerates the allowed values for the --status flag.
var validStatuses = map[string]bool{
	"draft":       true,
	"inProgress":  true,
	"halted":      true,
	"completed":   true,
}

// deployResult holds the structured output printed after a successful deploy.
type deployResult struct {
	VersionCode int64  `json:"versionCode"`
	Track       string `json:"track"`
	Status      string `json:"status"`
	PackageName string `json:"packageName"`
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Upload and release an APK or AAB to a Google Play track",
	Long: `Upload an Android App Bundle (AAB) or APK to Google Play and assign it
to a release track. This command creates an API edit, uploads the artifact,
optionally uploads a ProGuard/R8 mapping file, configures release notes,
and commits the edit.

Examples:
  gpc deploy --file app.aab --package com.example.app
  gpc deploy --file app.apk --track production --rollout 10 --status inProgress
  gpc deploy --file app.aab --release-notes '{"en-US": "Bug fixes"}'
  gpc deploy --file app.aab --release-notes-file notes.json --mapping-file mapping.txt`,
	RunE: runDeploy,
}

func init() {
	deployCmd.Flags().StringVar(&deployFile, "file", "", "path to AAB or APK file (required)")
	deployCmd.Flags().StringVar(&deployTrack, "track", "internal", "release track (e.g. internal, alpha, beta, production)")
	deployCmd.Flags().Float64Var(&deployRollout, "rollout", 100, "rollout percentage (0-100)")
	deployCmd.Flags().StringVar(&deployReleaseName, "release-name", "", "human-readable release name")
	deployCmd.Flags().StringVar(&deployReleaseNotes, "release-notes", "", `release notes as JSON (e.g. '{"en-US": "Bug fixes"}')`)
	deployCmd.Flags().StringVar(&deployReleaseNotesFile, "release-notes-file", "", "path to JSON file containing release notes")
	deployCmd.Flags().StringVar(&deployStatus, "status", "completed", "release status: draft, inProgress, halted, completed")
	deployCmd.Flags().StringVar(&deployMappingFile, "mapping-file", "", "path to ProGuard/R8 mapping file")

	_ = deployCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// --- Validate inputs ---

	if err := validateDeployFlags(); err != nil {
		return err
	}

	pkg, err := getPackage()
	if err != nil {
		return err
	}

	fileType, err := detectFileType(deployFile)
	if err != nil {
		return err
	}

	releaseNotes, err := resolveReleaseNotes(deployReleaseNotes, deployReleaseNotesFile)
	if err != nil {
		return err
	}

	// --- Authenticate and create service ---

	creds, err := auth.ResolveCredentials(auth.CredentialOptions{
		KeyFile: flagKeyFile,
	})
	if err != nil {
		return clierrors.AuthError("failed to resolve credentials", err)
	}

	svc, err := client.NewService(ctx, creds)
	if err != nil {
		return clierrors.AuthError("failed to create API service", err)
	}

	// --- Execute the deploy within an edit transaction ---

	var result deployResult

	err = edits.WithEdit(ctx, svc, pkg, func(editID string) error {
		// 1. Upload the artifact (AAB or APK).
		versionCode, uploadErr := uploadArtifact(ctx, svc, pkg, editID, deployFile, fileType)
		if uploadErr != nil {
			return uploadErr
		}

		// 2. Upload mapping file if provided.
		if deployMappingFile != "" {
			if mappingErr := uploadMapping(ctx, svc, pkg, editID, versionCode, deployMappingFile); mappingErr != nil {
				return mappingErr
			}
		}

		// 3. Update the track with release information.
		if trackErr := updateTrack(ctx, svc, pkg, editID, versionCode, releaseNotes); trackErr != nil {
			return trackErr
		}

		result = deployResult{
			VersionCode: versionCode,
			Track:       deployTrack,
			Status:      deployStatus,
			PackageName: pkg,
		}

		return nil
	})
	if err != nil {
		return clierrors.EditError("deploy failed", err)
	}

	// --- Print result ---

	if !flagQuiet {
		formatter := output.New(getFormatter(), os.Stdout)
		if fmtErr := formatter.Format(result); fmtErr != nil {
			return fmt.Errorf("formatting output: %w", fmtErr)
		}
	}

	return nil
}

// validateDeployFlags checks flag values for correctness before any network calls.
func validateDeployFlags() error {
	if !validStatuses[deployStatus] {
		return fmt.Errorf("invalid --status %q: must be one of draft, inProgress, halted, completed", deployStatus)
	}

	if deployRollout < 0 || deployRollout > 100 {
		return fmt.Errorf("invalid --rollout %.2f: must be between 0 and 100", deployRollout)
	}

	if deployReleaseNotes != "" && deployReleaseNotesFile != "" {
		return fmt.Errorf("--release-notes and --release-notes-file are mutually exclusive")
	}

	return nil
}

// detectFileType returns "aab" or "apk" based on the file extension, or an
// error if the extension is not recognised.
func detectFileType(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".aab":
		return "aab", nil
	case ".apk":
		return "apk", nil
	default:
		return "", fmt.Errorf("unsupported file type %q: must be .aab or .apk", ext)
	}
}

// resolveReleaseNotes parses release notes from the --release-notes flag value
// or from the file at --release-notes-file. The returned slice is nil when
// neither flag is set.
func resolveReleaseNotes(raw, filePath string) ([]*androidpublisher.LocalizedText, error) {
	var jsonData string

	switch {
	case raw != "":
		jsonData = raw
	case filePath != "":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading release notes file %q: %w", filePath, err)
		}
		jsonData = string(data)
	default:
		return nil, nil
	}

	// Parse the JSON map of language -> text.
	var notesMap map[string]string
	if err := json.Unmarshal([]byte(jsonData), &notesMap); err != nil {
		return nil, fmt.Errorf("parsing release notes JSON: %w", err)
	}

	notes := make([]*androidpublisher.LocalizedText, 0, len(notesMap))
	for lang, text := range notesMap {
		notes = append(notes, &androidpublisher.LocalizedText{
			Language: lang,
			Text:     text,
		})
	}

	return notes, nil
}

// uploadArtifact uploads an AAB or APK and returns the resulting version code.
func uploadArtifact(
	ctx context.Context,
	svc *androidpublisher.Service,
	pkg, editID, filePath, fileType string,
) (int64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, clierrors.UploadError(
			fmt.Sprintf("opening %s file %q", fileType, filePath), err,
		)
	}
	defer f.Close()

	// Determine file size for the progress bar.
	fi, err := f.Stat()
	if err != nil {
		return 0, clierrors.UploadError("reading file info", err)
	}

	var reader io.Reader = f

	// Wrap with a progress bar unless quiet mode is enabled.
	if !flagQuiet {
		bar := output.NewProgressBar(fi.Size(), fmt.Sprintf("Uploading %s...", filepath.Base(filePath)))
		reader = io.TeeReader(f, bar)
		defer bar.Finish()
	}

	var versionCode int64

	switch fileType {
	case "aab":
		bundle, uploadErr := svc.Edits.Bundles.Upload(pkg, editID).
			Media(reader).
			Context(ctx).
			Do()
		if uploadErr != nil {
			return 0, clierrors.UploadError("uploading AAB", uploadErr)
		}
		versionCode = bundle.VersionCode

	case "apk":
		apk, uploadErr := svc.Edits.Apks.Upload(pkg, editID).
			Media(reader).
			Context(ctx).
			Do()
		if uploadErr != nil {
			return 0, clierrors.UploadError("uploading APK", uploadErr)
		}
		versionCode = int64(apk.VersionCode)
	}

	if flagVerbose {
		fmt.Fprintf(os.Stderr, "Uploaded %s with version code %d\n", fileType, versionCode)
	}

	return versionCode, nil
}

// uploadMapping uploads a ProGuard/R8 deobfuscation mapping file for the
// given version code.
func uploadMapping(
	ctx context.Context,
	svc *androidpublisher.Service,
	pkg, editID string,
	versionCode int64,
	mappingPath string,
) error {
	f, err := os.Open(mappingPath)
	if err != nil {
		return clierrors.UploadError(
			fmt.Sprintf("opening mapping file %q", mappingPath), err,
		)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return clierrors.UploadError("reading mapping file info", err)
	}

	var reader io.Reader = f

	if !flagQuiet {
		bar := output.NewProgressBar(fi.Size(), "Uploading mapping file...")
		reader = io.TeeReader(f, bar)
		defer bar.Finish()
	}

	_, err = svc.Edits.Deobfuscationfiles.Upload(pkg, editID, versionCode, "proguard").
		Media(reader).
		Context(ctx).
		Do()
	if err != nil {
		return clierrors.UploadError("uploading mapping file", err)
	}

	if flagVerbose {
		fmt.Fprintf(os.Stderr, "Uploaded mapping file for version code %d\n", versionCode)
	}

	return nil
}

// updateTrack assigns the uploaded version code to the specified release track
// with the configured rollout, status, and optional release notes.
func updateTrack(
	ctx context.Context,
	svc *androidpublisher.Service,
	pkg, editID string,
	versionCode int64,
	releaseNotes []*androidpublisher.LocalizedText,
) error {
	// Convert rollout percentage to a fraction (0.0 - 1.0).
	rolloutFraction := deployRollout / 100.0

	release := &androidpublisher.TrackRelease{
		VersionCodes: []int64{versionCode},
		Status:       deployStatus,
	}

	// Only set UserFraction when the rollout is a staged rollout (not 100%).
	// The API rejects UserFraction for "completed" status or 100% rollouts.
	if rolloutFraction < 1.0 {
		release.UserFraction = rolloutFraction
	}

	if deployReleaseName != "" {
		release.Name = deployReleaseName
	}

	if releaseNotes != nil {
		release.ReleaseNotes = releaseNotes
	}

	track := &androidpublisher.Track{
		Track:    deployTrack,
		Releases: []*androidpublisher.TrackRelease{release},
	}

	_, err := svc.Edits.Tracks.Update(pkg, editID, deployTrack, track).
		Context(ctx).
		Do()
	if err != nil {
		return clierrors.APIError(
			fmt.Sprintf("updating track %q", deployTrack), err,
		)
	}

	if flagVerbose {
		fmt.Fprintf(os.Stderr, "Updated track %q: version %d, status %q, rollout %.0f%%\n",
			deployTrack, versionCode, deployStatus, deployRollout)
	}

	return nil
}
