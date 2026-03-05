# Product Requirements Document — `gpc` (Google Play Console CLI)

## Problem

There is no comprehensive, standalone CLI for the Google Play Developer API. Android developers and CI/CD pipelines rely on fragmented tooling — Gradle plugins that only handle uploads, web-only console access for metadata, and no scriptable interface for monetization, reviews, or vitals. The [App Store Connect CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) (`asc`) proves the demand: 72+ commands, 2.5k+ stars. The Play Store has no equivalent.

## Solution

**`gpc`** — a single Go binary that wraps the full Google Play Developer API surface into a scriptable, CI-friendly CLI. One tool for deploying, managing listings, handling subscriptions, reading reviews, and monitoring vitals.

## Target Users

1. **Android developers** who want to manage their apps from the terminal
2. **DevOps/CI engineers** building automated release pipelines
3. **Product teams** who need scriptable access to reviews, vitals, and monetization data

## Non-Goals

- GUI or TUI dashboard (this is a CLI tool)
- Replacing the Google Play Console web UI entirely
- Supporting non-Google-Play app stores

---

## Functional Requirements

### F1 — Authentication

| ID | Requirement | Priority |
|----|-------------|----------|
| F1.1 | Support Google service account JSON key files | P0 |
| F1.2 | Credential resolution chain: `--key-file` flag > `GPC_KEY_FILE` env > `GPC_KEY_JSON` env > system keyring > ADC | P0 |
| F1.3 | Secure credential storage in system keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager) | P0 |
| F1.4 | Guided setup via `gpc auth init` | P0 |
| F1.5 | Credential validation via `gpc auth test` | P0 |
| F1.6 | Auth status introspection via `gpc auth status` | P1 |

### F2 — Deployment (MVP)

| ID | Requirement | Priority |
|----|-------------|----------|
| F2.1 | Upload AAB files via `gpc bundles upload` | P0 |
| F2.2 | Upload APK files via `gpc apks upload` | P0 |
| F2.3 | Manage release tracks (internal, alpha, beta, production, custom) | P0 |
| F2.4 | Flagship `gpc deploy` command combining upload + track assignment in one step | P0 |
| F2.5 | Staged rollouts with configurable percentage | P0 |
| F2.6 | Release notes in multiple languages (JSON format) | P0 |
| F2.7 | ProGuard/R8 mapping file upload alongside builds | P1 |
| F2.8 | Upload progress bar | P1 |
| F2.9 | Draft, in-progress, halted, and completed release statuses | P0 |

### F3 — App Metadata & Listings

| ID | Requirement | Priority |
|----|-------------|----------|
| F3.1 | View app details | P1 |
| F3.2 | List, get, update, and delete store listings per language | P1 |
| F3.3 | Update title, short description, and full description | P1 |

### F4 — Images & Screenshots

| ID | Requirement | Priority |
|----|-------------|----------|
| F4.1 | List, upload, and delete screenshots and images by type and language | P1 |
| F4.2 | Batch upload from a directory (concurrent) | P2 |

### F5 — Reviews

| ID | Requirement | Priority |
|----|-------------|----------|
| F5.1 | List and read user reviews | P1 |
| F5.2 | Reply to reviews from the CLI | P1 |
| F5.3 | Pagination for large review sets | P2 |

### F6 — Testers

| ID | Requirement | Priority |
|----|-------------|----------|
| F6.1 | View testers per track | P1 |
| F6.2 | Update tester email lists per track | P1 |

### F7 — Monetization

| ID | Requirement | Priority |
|----|-------------|----------|
| F7.1 | CRUD for in-app products (managed products) | P1 |
| F7.2 | CRUD for subscriptions | P1 |
| F7.3 | Manage subscription base plans and offers | P2 |
| F7.4 | Archive subscriptions | P2 |

### F8 — Purchase Verification

| ID | Requirement | Priority |
|----|-------------|----------|
| F8.1 | Verify product purchases by SKU and token | P1 |
| F8.2 | Verify and cancel subscription purchases | P1 |
| F8.3 | Refund orders with optional revocation | P2 |

### F9 — Android Vitals

| ID | Requirement | Priority |
|----|-------------|----------|
| F9.1 | View crash rate metrics | P2 |
| F9.2 | View ANR rate metrics | P2 |
| F9.3 | Search error issues and reports | P2 |
| F9.4 | View slow rendering and slow start metrics | P2 |
| F9.5 | View anomaly alerts | P3 |

### F10 — Financial Reports

| ID | Requirement | Priority |
|----|-------------|----------|
| F10.1 | List available financial reports (from GCS) | P2 |
| F10.2 | Download financial reports | P2 |

### F11 — User & Permission Management

| ID | Requirement | Priority |
|----|-------------|----------|
| F11.1 | List, create, and delete developer account users | P3 |
| F11.2 | Manage user grants and permissions | P3 |

### F12 — Edit Escape Hatch

| ID | Requirement | Priority |
|----|-------------|----------|
| F12.1 | Manually insert, validate, and commit edits for advanced workflows | P2 |

---

## Non-Functional Requirements

### NF1 — Output

| ID | Requirement |
|----|-------------|
| NF1.1 | JSON output by default (machine-parseable, pipeable to `jq`) |
| NF1.2 | Table output opt-in via `--output table` |
| NF1.3 | Quiet mode (`--quiet`) suppresses non-essential output |
| NF1.4 | Verbose mode (`--verbose`) for debugging |

### NF2 — Reliability

| ID | Requirement |
|----|-------------|
| NF2.1 | Automatic retry with exponential backoff on 429 and 5xx errors |
| NF2.2 | Edit transactions are always cleaned up on failure (no orphaned edits) |
| NF2.3 | User-friendly error messages with actionable suggestions |

### NF3 — Security

| ID | Requirement |
|----|-------------|
| NF3.1 | Credentials never written to disk in plaintext (keyring only) |
| NF3.2 | `GPC_KEY_JSON` env var support for CI secrets injection |
| NF3.3 | No credentials logged in verbose mode |

### NF4 — Distribution

| ID | Requirement |
|----|-------------|
| NF4.1 | Single static binary (no runtime dependencies) |
| NF4.2 | Cross-platform: macOS (arm64, amd64), Linux (amd64, arm64), Windows (amd64) |
| NF4.3 | Homebrew tap for macOS/Linux |
| NF4.4 | GitHub Releases with GoReleaser |
| NF4.5 | Shell completions for bash, zsh, fish |
| NF4.6 | GitHub Action (`setup-gpc`) for CI integration |

### NF5 — Performance

| ID | Requirement |
|----|-------------|
| NF5.1 | CLI startup under 100ms |
| NF5.2 | Concurrent image uploads where possible |

---

## Command Reference

```
gpc auth init|status|test
gpc deploy --file <path> [--track <name>] [--rollout <pct>] [--status <s>] [--release-notes <json>]
gpc bundles list|upload
gpc apks list|upload
gpc tracks list|get|update
gpc apps list|info
gpc listings list|get|update|delete
gpc images list|upload|delete
gpc reviews list|get|reply
gpc testers get|update
gpc subscriptions list|get|create|update|archive
gpc products list|get|create|update|delete
gpc purchases product-get|sub-get|sub-cancel
gpc orders get|refund
gpc vitals crash-rate|anr-rate|errors|slow-rendering|slow-start|anomalies
gpc reports list|download
gpc users list|create|delete
gpc grants create|update|delete
gpc edits insert|validate|commit
```

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--package` | `-p` | — | Android package name |
| `--key-file` | — | — | Path to service account JSON key |
| `--output` | `-o` | `json` | Output format: `json` or `table` |
| `--quiet` | `-q` | `false` | Suppress non-essential output |
| `--verbose` | `-v` | `false` | Enable verbose/debug output |

---

## Milestones

### M1 — MVP (Steps 0-4)

**Success criteria:** `gpc auth init` followed by `gpc deploy --file app.aab --track internal` works end-to-end against a real Play Console account.

Includes: project setup, auth system, API client with retry, edit transactions, output formatting, deploy command, bundles/APKs/tracks management.

### M2 — Content Management (Steps 5-7)

**Success criteria:** Full store listing and image management from the CLI. Review replies work.

Includes: app metadata, listings CRUD, image upload/delete, reviews list/reply, testers management.

### M3 — Monetization (Step 8)

**Success criteria:** In-app products and subscriptions can be fully managed. Purchase verification works.

Includes: products CRUD, subscriptions CRUD with base plans and offers, purchase verification, order refunds.

### M4 — Observability (Step 9)

**Success criteria:** Crash rates, ANR rates, and error reports are queryable from the CLI.

Includes: Play Developer Reporting API client, vitals commands.

### M5 — Enterprise & Distribution (Steps 10-11)

**Success criteria:** `brew install charliesbot/tap/gpc` works. GitHub Action available. Financial reports downloadable.

Includes: GCS client for reports, user/grant management, GoReleaser, Homebrew tap, GitHub Action, shell completions, CI/CD docs.

---

## Success Metrics

| Metric | Target |
|--------|--------|
| GitHub stars (6 months) | 500+ |
| Total commands | 60+ |
| Platform coverage | macOS, Linux, Windows |
| API coverage | 90%+ of Play Developer API v3 |
| CI adoption | Used in 3+ open-source Android projects |

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Single binary, fast startup, excellent CLI ecosystem (Cobra, Charm) |
| CLI framework | Cobra | Industry standard (used by gh, kubectl, docker) |
| Auth storage | 99designs/keyring | Cross-platform keychain support without CGO on most platforms |
| Output default | JSON | Machine-parseable, pipeable to jq, CI-friendly |
| API client | google.golang.org/api | Official Google client libraries |
| Edit pattern | WithEdit() wrapper | Prevents orphaned edits, ensures cleanup on failure |

---

## Open Questions

1. Should `gpc apps list` attempt to work around the missing "list apps" API endpoint (e.g., by reading from config)?
2. Should we support YAML config in addition to JSON?
3. Should `gpc deploy` support watching the rollout status after deployment?
4. Should we add a `gpc init` command that creates a `.gpc.yml` project config file?
