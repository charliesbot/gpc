package edits

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

// WithEdit creates an edit, runs fn with the edit ID, validates, and commits.
// On any error in fn or validation, the edit is deleted (cleaned up).
func WithEdit(ctx context.Context, svc *androidpublisher.Service, pkg string, fn func(editID string) error) error {
	edit, err := svc.Edits.Insert(pkg, nil).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating edit: %w", err)
	}

	editID := edit.Id

	if err := fn(editID); err != nil {
		_ = deleteEdit(ctx, svc, pkg, editID)
		return fmt.Errorf("executing edit operation: %w", err)
	}

	if _, err := svc.Edits.Validate(pkg, editID).Context(ctx).Do(); err != nil {
		_ = deleteEdit(ctx, svc, pkg, editID)
		return fmt.Errorf("validating edit: %w", err)
	}

	if _, err := svc.Edits.Commit(pkg, editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("committing edit: %w", err)
	}

	return nil
}

// deleteEdit attempts to clean up an edit. The error is intentionally
// discarded by callers because the original error is more relevant.
func deleteEdit(ctx context.Context, svc *androidpublisher.Service, pkg, editID string) error {
	err := svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting edit: %w", err)
	}
	return nil
}
