package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
)

// requiredScopes defines the OAuth2 scopes needed for Google Play Developer API access.
var requiredScopes = []string{
	androidpublisher.AndroidpublisherScope,
}

// CredentialOptions configures how credentials are resolved.
type CredentialOptions struct {
	// KeyFile is the path to a service account JSON key file,
	// typically provided via the --key-file CLI flag.
	KeyFile string

	// UseKeyring enables loading credentials from the system keychain
	// as part of the resolution chain.
	UseKeyring bool
}

// ResolveCredentials walks the credential resolution chain and returns the
// first successfully obtained credentials. The resolution order is:
//
//  1. Explicit key file path (--key-file flag / CredentialOptions.KeyFile)
//  2. GPC_KEY_FILE environment variable (path to a JSON key file)
//  3. GPC_KEY_JSON environment variable (inline JSON key content)
//  4. System keyring (if UseKeyring is true)
//  5. Application Default Credentials (ADC)
//
// If every step fails, an error summarising the chain is returned.
func ResolveCredentials(opts CredentialOptions) (*google.Credentials, error) {
	ctx := context.Background()

	// Step 1: Explicit key file path from the --key-file flag.
	if opts.KeyFile != "" {
		creds, err := credentialsFromFile(ctx, opts.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading key file %q: %w", opts.KeyFile, err)
		}
		return creds, nil
	}

	// Step 2: GPC_KEY_FILE environment variable.
	if path := os.Getenv("GPC_KEY_FILE"); path != "" {
		creds, err := credentialsFromFile(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("loading key file from GPC_KEY_FILE (%q): %w", path, err)
		}
		return creds, nil
	}

	// Step 3: GPC_KEY_JSON environment variable (raw JSON content).
	if jsonKey := os.Getenv("GPC_KEY_JSON"); jsonKey != "" {
		creds, err := credentialsFromJSON(ctx, []byte(jsonKey))
		if err != nil {
			return nil, fmt.Errorf("parsing credentials from GPC_KEY_JSON: %w", err)
		}
		return creds, nil
	}

	// Step 4: System keyring.
	if opts.UseKeyring {
		jsonKey, err := LoadCredential()
		if err == nil && len(jsonKey) > 0 {
			creds, err := credentialsFromJSON(ctx, jsonKey)
			if err != nil {
				return nil, fmt.Errorf("parsing credentials from keyring: %w", err)
			}
			return creds, nil
		}
		// Keyring miss is non-fatal; fall through to ADC.
	}

	// Step 5: Application Default Credentials (ADC).
	creds, err := google.FindDefaultCredentials(ctx, requiredScopes...)
	if err != nil {
		return nil, fmt.Errorf(
			"no credentials found: set --key-file, GPC_KEY_FILE, GPC_KEY_JSON, "+
				"store a key in the system keyring, or configure Application Default Credentials: %w",
			err,
		)
	}
	return creds, nil
}

// credentialsFromFile reads a service account JSON key file from disk and
// returns configured credentials.
func credentialsFromFile(ctx context.Context, path string) (*google.Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return credentialsFromJSON(ctx, data)
}

// credentialsFromJSON parses raw service account JSON and returns configured
// credentials with the required Android Publisher scopes.
func credentialsFromJSON(ctx context.Context, jsonKey []byte) (*google.Credentials, error) {
	creds, err := google.CredentialsFromJSON(ctx, jsonKey, requiredScopes...)
	if err != nil {
		return nil, fmt.Errorf("parsing service account JSON: %w", err)
	}
	return creds, nil
}
