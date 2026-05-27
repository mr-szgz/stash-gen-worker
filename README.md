# stash-gen-worker

Standalone worker CLI for generating Stash-compatible scene assets outside the main Stash server.

## Features

- self-contained Go module with no sibling `../stash` checkout requirement
- reuses a minimal vendored subset of upstream Stash generation code
- generates scene preview MP4
- generates preview WebP
- generates screenshot JPG
- generates sprite JPG + VTT
- generates transcodes
- supports JSON job files for distributed workers
- includes sync scripts and a manual GitHub Actions workflow for upstream vendor refreshes

## Repository layout

- `cmd/stash-gen-worker/` - CLI entrypoint
- `internal/worker/` - worker orchestration and output path helpers
- `third_party/stash/` - vendored upstream Stash files used by the worker
- `scripts/` - upstream sync scripts (`.sh` and `.ps1`)
- `.agents/references/` - practical grounding docs for future agents

## Build

### Linux/macOS

```bash
go build ./cmd/stash-gen-worker
```

### Windows cross-compile

```bash
GOOS=windows GOARCH=amd64 go build -o stash-gen-worker.exe ./cmd/stash-gen-worker
```

## Runtime requirements

The worker needs:

- `ffmpeg`
- `ffprobe`

They can be supplied explicitly or discovered on `PATH`.

## Usage

### Simple command flags

```bash
stash-gen-worker \
  --input /path/to/scene.mp4 \
  --checksum abc123 \
  --generated ./generated \
  --preview \
  --webp \
  --screenshot \
  --sprite \
  --transcode
```

### JSON job file

```json
{
  "input_path": "/path/to/scene.mp4",
  "checksum": "abc123",
  "generated_dir": "./generated",
  "preview": true,
  "webp": true,
  "screenshot": true,
  "sprite": true,
  "transcode": false,
  "overwrite": true,
  "preview_options": {
    "segments": 12,
    "segment_duration": 0.5,
    "exclude_start": "0",
    "exclude_end": "0",
    "preset": "veryfast",
    "audio": false,
    "fallback": false,
    "use_vsync_2": false
  },
  "sprite_options": {
    "count": 25,
    "size": 320
  }
}
```

Run it with:

```bash
stash-gen-worker --job ./job.json
```

## Vendored upstream sync

The repo vendors a minimal subset of `stashapp/stash` under `third_party/stash/`.

- Linux/macOS: `scripts/sync-stash-vendor.sh`
- PowerShell: `scripts/sync-stash-vendor.ps1`
- GitHub Actions: manually run `Update vendored Stash files`

The workflow syncs vendored files, runs `go mod tidy`, runs `go test ./...`, and opens a PR with the refresh.
