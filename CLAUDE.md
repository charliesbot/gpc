# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`gpc` (Google Play Console CLI) is a Go CLI that wraps the Google Play Developer API v3 into a single binary. It uses Cobra for command structure, the official Google API client libraries, and targets CI/CD pipelines and Android developers.

Module path: `github.com/charliesbot/gpc`
Go version: 1.25.0

## Build & Run

```bash
make build          # outputs bin/gpc
make install        # go install to $GOPATH/bin
make test           # go test ./... -v
make lint           # golangci-lint run ./...
make build-all      # cross-compile for macOS/Linux/Windows (arm64+amd64)
make clean          # rm -rf bin/
```

Version is injected at build time via `-ldflags "-X main.version=$(VERSION)"`. The `main.go` calls `cmd.SetVersion(version)` before `cmd.Execute()`.

## Architecture

### Entry Point

`main.go` -> `cmd.SetVersion()` -> `cmd.Execute()` (runs the root Cobra command). Errors cause `os.Exit(1)` with no message (Cobra's `SilenceUsage` and `SilenceErrors` are both true on the root command).

### Package Layout

```
cmd/           Cobra commands. Each file defines commands, flags, init(), and RunE handlers.
internal/
  auth/        Credential resolution chain + system keyring integration (99designs/keyring)
  client/      API service factory (client.NewService) + retry logic with exponential backoff
  config/      ~/.config/gpc/config.json read/write
  edits/       WithEdit() transaction wrapper (insert -> fn -> validate -> commit, or cleanup)
  errors/      CLIError struct with machine-readable codes (auth_failed, edit_failed, etc.)
  output/      Formatter interface with JSON (default) and table implementations + progress bar
```

### How Commands Work (the standard pattern)

Every command follows this sequence in its `RunE` function:

1. **Validate inputs** -- check flags early, before any network calls
2. **Resolve package** -- `pkg, err := getPackage()` (checks `--package` flag, then `GPC_PACKAGE` env)
3. **Authenticate** -- `auth.ResolveCredentials(auth.CredentialOptions{KeyFile: flagKeyFile})` then `client.NewService(ctx, creds)`
4. **Call the API** -- either directly or wrapped in `edits.WithEdit(ctx, svc, pkg, func(editID string) error { ... })`
5. **Format output** -- `output.New(getFormatter(), os.Stdout).Format(result)`

Two auth patterns exist in the codebase:
- **Inline** (deploy.go, bundles.go, tracks.go): calls `auth.ResolveCredentials` + `client.NewService` directly
- **buildService helper** (reviews.go): wraps the two calls into a `buildService(ctx)` helper that lives at the bottom of the file

Both are acceptable. The `buildService` helper sets `UseKeyring: true` while the inline pattern in deploy.go does not -- be aware of this inconsistency.

### Edit Transaction Pattern (critical)

Many Play Developer API operations require an "edit" -- a transaction that groups changes. The `edits.WithEdit()` helper manages the full lifecycle:

```go
edits.WithEdit(ctx, svc, pkg, func(editID string) error {
    // upload, update track, etc. using editID
    return nil
})
```

- On success: validates then commits the edit
- On failure (fn returns error or validation fails): deletes the edit (cleanup)
- The delete error is intentionally discarded -- the original error is more important

Use `WithEdit` for any operation that touches: bundles, APKs, tracks, listings, images, or deobfuscation files. Reviews and purchases do NOT require edits.

### Auth Credential Resolution Chain

Resolution order (first match wins):
1. `--key-file` flag
2. `GPC_KEY_FILE` env var (path to JSON file)
3. `GPC_KEY_JSON` env var (inline JSON -- useful for CI secrets)
4. System keyring (macOS Keychain / Linux Secret Service / Windows Credential Manager)
5. Application Default Credentials (ADC / gcloud auth)

### Output Formatting

Two modes controlled by `--output` flag (default: `json`):
- **JSON**: pretty-printed with 2-space indent via `json.NewEncoder`. Machine-parseable.
- **Table**: uses stdlib `text/tabwriter`. Commands must build `output.TableData{Headers, Rows}` manually.

The table path requires an explicit `if getFormatter() == "table"` branch in each command. JSON mode accepts `any` type and marshals it directly. Auth commands are an exception -- they print directly to stdout/stderr since they are human-facing only.

### Error Handling

- `internal/errors` provides `CLIError` with `Code`, `Message`, and wrapped `Err`
- Constructor helpers: `clierrors.AuthError()`, `clierrors.EditError()`, `clierrors.UploadError()`, `clierrors.APIError()`
- Import alias convention: `clierrors "github.com/charliesbot/gpc/internal/errors"` (to avoid collision with stdlib `errors`)
- All errors use `fmt.Errorf("context: %w", err)` wrapping for the error chain
- CLIError is used in deploy.go; simpler commands use plain `fmt.Errorf`

### Retry Logic

`client.Retry(ctx, maxRetries, fn)` with exponential backoff (1s base, 30s cap, random jitter). Retries on:
- HTTP 429 (Too Many Requests)
- HTTP 5xx (server errors)
- Any error implementing `RetryableError` interface

### Progress Bar

`output.NewProgressBar(size, description)` implements `io.Writer`. Wrap upload streams with `io.TeeReader(file, bar)`. The bar renders to stderr so it does not pollute stdout JSON output. Call `bar.Finish()` after the upload completes. Skip the progress bar when `flagQuiet` is true.

## Adding a New Command

1. Create `cmd/<name>.go`
2. Define package-scoped flag variables at the top
3. Define parent `cobra.Command` + subcommands as package-scoped vars
4. In `init()`: register flags, mark required flags with `_ = cmd.MarkFlagRequired("name")`, add subcommands, call `rootCmd.AddCommand(parentCmd)`
5. Write `RunE` handler following the standard pattern above
6. For table output, provide a `<name>ToTable()` conversion function returning `output.TableData`

## Global Flags (defined in cmd/root.go)

| Flag | Short | Type | Default | Env Fallback |
|------|-------|------|---------|-------------|
| `--package` | `-p` | string | (none) | `GPC_PACKAGE` |
| `--key-file` | | string | (none) | `GPC_KEY_FILE` |
| `--output` | `-o` | string | `json` | (none) |
| `--quiet` | `-q` | bool | false | (none) |
| `--verbose` | `-v` | bool | false | (none) |

These are `PersistentFlags` on the root command, accessible to all subcommands via the package-level variables `flagPackage`, `flagKeyFile`, `flagOutput`, `flagQuiet`, `flagVerbose`.

## Conventions

- **Naming**: command files match their CLI noun (`bundles.go`, `tracks.go`, `reviews.go`). Flag vars use `<command><FlagName>` pattern (e.g. `deployFile`, `trackName`).
- **No tests yet**: the project has zero test files currently. Tests should use `httptest.NewServer` to mock HTTP responses per the architecture doc.
- **Linting**: golangci-lint with errcheck (including type assertions), govet, staticcheck, unused, gosimple, ineffassign, misspell (US locale), gofmt.
- **Error strings**: lowercase, no trailing punctuation, use `%w` for wrapping.
- **Verbose output**: goes to stderr via `fmt.Fprintf(os.Stderr, ...)`. Never to stdout.
- **Config path**: `~/.config/gpc/config.json` with `default_package`, `default_key_file`, `default_output` fields.
- **Required flags**: use `_ = cmd.MarkFlagRequired("name")` -- the error return is discarded with blank identifier.

## Key Dependencies

| Library | Import Path | Purpose |
|---------|------------|---------|
| Cobra | `github.com/spf13/cobra` | CLI framework |
| keyring | `github.com/99designs/keyring` | Cross-platform keychain (macOS/Linux/Windows) |
| Google API | `google.golang.org/api/androidpublisher/v3` | Play Developer API v3 client |
| oauth2 | `golang.org/x/oauth2/google` | Service account credential handling |
| tabwriter | `text/tabwriter` (stdlib) | Table output alignment |

## Docs

- `docs/ARCHITECTURE.md` -- project layout, how-to-add-a-command, edit pattern, auth chain
- `docs/PRD.md` -- full product requirements, milestones, command reference, non-functional requirements
