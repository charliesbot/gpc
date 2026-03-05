# Architecture

`gpc` is a CLI for the Google Play Developer API, written in Go using [Cobra](https://github.com/spf13/cobra).

## Project Layout

```
main.go                    Entry point — delegates to cmd.Execute()
cmd/
  root.go                  Root Cobra command, global flags (--package, --key-file, --output, etc.)
  auth.go                  gpc auth init|status|test
  deploy.go                gpc deploy (flagship command)
  bundles.go               gpc bundles list|upload
  apks.go                  gpc apks list|upload
  tracks.go                gpc tracks list|get|update
  apps.go                  gpc apps list|info
  listings.go              gpc listings list|get|update|delete
  images.go                gpc images list|upload|delete
  reviews.go               gpc reviews list|get|reply
  testers.go               gpc testers get|update
  subscriptions.go         gpc subscriptions list|get|create|update|archive
  products.go              gpc products list|get|create|update|delete
  purchases.go             gpc purchases product-get|sub-get|sub-cancel
  orders.go                gpc orders get|refund
  vitals.go                gpc vitals crash-rate|anr-rate|... (stubbed)
  reports.go               gpc reports list|download (stubbed)
  users.go                 gpc users list|create|delete (stubbed)
  grants.go                gpc grants create|update|delete (stubbed)
  edits.go                 gpc edits insert|validate|commit (escape hatch)
internal/
  auth/
    credentials.go         Credential resolution chain
    keyring.go             System keychain integration (macOS/Linux/Windows)
  client/
    client.go              Authenticated androidpublisher.Service factory
    retry.go               Exponential backoff with jitter for 429/5xx
  config/
    config.go              ~/.config/gpc/config.json read/write
  edits/
    transaction.go         WithEdit() — insert → fn → validate → commit (or cleanup)
  output/
    formatter.go           Formatter interface, New() factory
    json.go                JSON pretty-print (default)
    table.go               ASCII table output (--output table)
    progress.go            Upload progress bar (stderr)
  errors/
    errors.go              CLIError with code, message, wrapped error
```

## How to Add a New Command

1. Create `cmd/<name>.go`
2. Define a parent `cobra.Command` and subcommands
3. In `init()`, call `rootCmd.AddCommand(parentCmd)`
4. Use helpers from `cmd/root.go`:
   - `getPackage()` — resolves package name from `--package` flag or `GPC_PACKAGE` env
   - `getFormatter()` — returns `"json"` or `"table"` based on `--output` flag
5. Authenticate and create a service:
   ```go
   creds, err := auth.ResolveCredentials(auth.CredentialOptions{KeyFile: flagKeyFile, UseKeyring: true})
   svc, err := client.NewService(ctx, creds)
   ```
6. For edit-scoped operations, wrap in `edits.WithEdit()`:
   ```go
   edits.WithEdit(ctx, svc, pkg, func(editID string) error {
       // API calls using editID
       return nil
   })
   ```
7. Output results via the formatter:
   ```go
   output.New(getFormatter(), os.Stdout).Format(result)
   ```

## Edit Transaction Pattern

Many Play Developer API operations require an "edit" — a transaction that groups changes:

1. **Insert** a new edit (gets an edit ID)
2. **Perform operations** using the edit ID
3. **Validate** the edit
4. **Commit** the edit

If any step fails, the edit is **deleted** (rolled back). The `edits.WithEdit()` helper manages this lifecycle automatically.

```go
edits.WithEdit(ctx, svc, pkg, func(editID string) error {
    // Upload a bundle, update a track, etc.
    return nil
})
// Edit is validated and committed on success, deleted on failure.
```

## Auth Credential Resolution

Credentials are resolved in priority order (first match wins):

```
--key-file flag → GPC_KEY_FILE env → GPC_KEY_JSON env → system keyring → ADC
```

- `--key-file`: Path to a service account JSON key file
- `GPC_KEY_FILE`: Same, via environment variable
- `GPC_KEY_JSON`: Inline JSON content (useful in CI)
- System keyring: Stored via `gpc auth init`
- ADC: Application Default Credentials (gcloud auth)

## Output Formatting

- **JSON** (default): Pretty-printed to stdout. Machine-parseable.
- **Table** (`--output table`): Human-readable ASCII tables.

All commands use the `output.Formatter` interface. Auth commands are the exception — they print directly to stdout/stderr since they're human-facing.

## Testing

Unit tests use `httptest.Server` to mock HTTP responses:

```go
ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(expected)
}))
defer ts.Close()
```

Run tests:
```bash
make test
```

## Dependencies

| Library | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `text/tabwriter` (stdlib) | Table output |
| `github.com/99designs/keyring` | Secure credential storage |
| `google.golang.org/api/androidpublisher/v3` | Play Developer API |
| `golang.org/x/oauth2/google` | Service account auth |

## Config

Config lives at `~/.config/gpc/config.json`:

```json
{
  "default_package": "com.example.app",
  "default_key_file": "/path/to/key.json",
  "default_output": "json"
}
```
