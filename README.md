# gpc — Google Play Console CLI

A command-line interface for the Google Play Developer API.

## Install

```bash
# Homebrew
brew install charliesbot/tap/gpc

# Go
go install github.com/charliesbot/gpc@latest

# Binary
# Download from https://github.com/charliesbot/gpc/releases
```

## Quick Start

```bash
# Set up authentication
gpc auth init --key-file service-account.json

# Deploy an AAB to internal testing
gpc deploy --file app.aab --track internal --package com.example.app

# List tracks
gpc tracks list --package com.example.app

# Manage listings
gpc listings list --package com.example.app
gpc listings update --package com.example.app --language en-US --title "My App"
```

## Authentication

`gpc` uses a Google service account. [Create one](https://developers.google.com/android-publisher/getting_started) with access to the Google Play Developer API.

Credentials are resolved in order:

1. `--key-file` flag
2. `GPC_KEY_FILE` environment variable
3. `GPC_KEY_JSON` environment variable (inline JSON, useful for CI)
4. System keyring (stored via `gpc auth init`)
5. Application Default Credentials

```bash
# Store credentials in system keyring
gpc auth init --key-file path/to/key.json

# Or use environment variables
export GPC_KEY_FILE=path/to/key.json
export GPC_PACKAGE=com.example.app
```

## Commands

| Command | Description |
|---------|-------------|
| `gpc auth init\|status\|test` | Manage authentication |
| `gpc deploy` | Upload and release (flagship) |
| `gpc bundles list\|upload` | App bundles |
| `gpc apks list\|upload` | APK files |
| `gpc tracks list\|get\|update` | Release tracks |
| `gpc apps list\|info` | App details |
| `gpc listings list\|get\|update\|delete` | Store listings |
| `gpc images list\|upload\|delete` | Screenshots & images |
| `gpc reviews list\|get\|reply` | User reviews |
| `gpc testers get\|update` | Test track testers |
| `gpc subscriptions list\|get\|create\|update\|archive` | Subscriptions |
| `gpc products list\|get\|create\|update\|delete` | In-app products |
| `gpc purchases product-get\|sub-get\|sub-cancel` | Purchase verification |
| `gpc orders get\|refund` | Order management |
| `gpc vitals crash-rate\|anr-rate\|...` | Android Vitals |
| `gpc edits insert\|validate\|commit` | Raw edit operations |

## Global Flags

```
--package, -p   Android package name
--key-file      Path to service account JSON key
--output, -o    Output format: json (default) or table
--quiet, -q     Suppress non-essential output
--verbose, -v   Enable verbose output
```

## CI/CD

```yaml
# GitHub Actions
- name: Deploy to Play Store
  run: |
    gpc deploy \
      --file app.aab \
      --track production \
      --rollout 10 \
      --release-notes '{"en-US": "Bug fixes and improvements"}'
  env:
    GPC_KEY_JSON: ${{ secrets.PLAY_STORE_KEY }}
    GPC_PACKAGE: com.example.app
```

## License

MIT
