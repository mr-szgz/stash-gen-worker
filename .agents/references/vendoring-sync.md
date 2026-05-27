# Vendoring and sync strategy

The repo intentionally vendors a small subset of `stashapp/stash` under `third_party/stash/`.

## Rules

- Prefer copying upstream files unchanged except for import-path rewrites to this module.
- Keep worker-specific changes outside `third_party/stash/` whenever possible.
- Treat `third_party/stash/` as a pinned mirror of the minimal dependency cone needed for standalone builds.

## Current upstream pin

- Upstream repo: `stashapp/stash`
- Upstream ref: `4187d164b349f8442a4f31c72bb477302590a9a4`

## Sync process

- Linux/macOS: `scripts/sync-stash-vendor.sh`
- Windows PowerShell: `scripts/sync-stash-vendor.ps1`
- GitHub Actions: `.github/workflows/update-vendored-stash.yml`

The sync scripts:

1. download the pinned upstream files into `third_party/stash/`
2. rewrite imports from `github.com/stashapp/stash/pkg/...` to local `third_party/stash/pkg/...`
3. leave local glue files alone

After syncing, run:

- `go mod tidy`
- `go test ./...`
- `go build ./cmd/stash-gen-worker`
