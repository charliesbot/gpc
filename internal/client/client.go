package client

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// NewService creates an authenticated androidpublisher.Service using the
// provided Google credentials. The returned service can be used to interact
// with the Google Play Developer API v3.
func NewService(ctx context.Context, creds *google.Credentials) (*androidpublisher.Service, error) {
	svc, err := androidpublisher.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create androidpublisher service: %w", err)
	}
	return svc, nil
}
