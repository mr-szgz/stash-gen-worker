# stash-gen-worker

Standalone worker CLI for generating Stash-compatible scene assets outside the main Stash server.

## Features

- Reuses official Stash generation packages
- Generates scene preview MP4
- Generates preview WebP
- Generates screenshot JPG
- Generates sprite JPG + VTT
- Generates transcodes
- Supports JSON job files for distributed workers
- Cross-compiles for Windows

## Important note about dependencies

This worker imports official packages from `github.com/stashapp/stash`.

The current `go.mod` is configured with:

- a `require` on `github.com/stashapp/stash`
- a local `replace github.com/stashapp/stash => ../stash`

That means the app is immediately compilable when the Stash source tree is checked out adjacent to this repository:

- `../stash-gen-worker`
- `../stash`

This is the safest way to keep the worker aligned with Stash internals while the CLI is being developed.

If you want, the next step can be to vendor/fork only the minimal Stash packages needed so this repo builds fully standalone without a local sibling checkout.

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

They can be supplied explicitly or discovered on PATH.

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
