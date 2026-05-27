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
- Can create job files from Stash GraphQL scene IDs
- Can process queued jobs from `jobs/new`, `jobs/pending`, `jobs/completed`, and `jobs/failed`
- Cross-compiles for Windows

## Dependencies

This worker imports official packages from `github.com/stashapp/stash`.

The module now builds directly against the published Stash module version declared in `go.mod`.

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

### Worker config

You can keep shared worker settings in a JSON config file and pass it with `--config`.

```json
{
  "jobs_dir": "./jobs",
  "generated_dir": "/srv/stash/.stash/generated",
  "ffmpeg_path": "",
  "ffprobe_path": "",
  "stash_graphql_endpoint": "http://localhost:9999/graphql",
  "stash_api_key": ""
}
```

`generated_dir` should be the worker host path that writes directly into the Stash server's generated assets directory.

### Run a single job directly

```bash
stash-gen-worker \
  --config ./worker-config.json \
  --input /path/to/scene.mp4 \
  --checksum abc123 \
  --generated /srv/stash/.stash/generated \
  --preview \
  --webp \
  --screenshot \
  --sprite \
  --transcode
```

### JSON job file

```json
{
  "schema_version": 1,
  "scene_id": "123",
  "scene_title": "Sample Scene",
  "input_path": "/path/to/scene.mp4",
  "checksum": "abc123",
  "generated_dir": "/srv/stash/.stash/generated",
  "preview": true,
  "webp": true,
  "screenshot": true,
  "sprite": true,
  "transcode": false,
  "overwrite": true,
  "retry_count": 0,
  "max_retries": 3,
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
stash-gen-worker --config ./worker-config.json --job ./job.json
```

### Create a queued job from Stash GraphQL

This creates a JSON job in `jobs/new/` using a Stash scene ID and the scene's first file path plus MD5 fingerprint.

```bash
stash-gen-worker generate-job \
  --config ./worker-config.json \
  --scene-id 123
```

You can also override the queue or generated destination at creation time:

```bash
stash-gen-worker generate-job \
  --scene-id 123 \
  --stash-url http://localhost:9999/graphql \
  --stash-api-key your-api-key \
  --jobs-dir ./jobs \
  --generated /srv/stash/.stash/generated
```

If no asset flags are supplied for `generate-job`, the worker enables preview, webp, screenshot, and sprite generation by default.

### Process queued jobs

Process a single queued job:

```bash
stash-gen-worker run-next --config ./worker-config.json
```

Process all currently queued jobs:

```bash
stash-gen-worker run-queue --config ./worker-config.json
```

Recover stale pending jobs faster (default is 15 minutes):

```bash
stash-gen-worker run-next --config ./worker-config.json --pending-timeout 2m
```

Requeue jobs from failed/pending back to new:

```bash
stash-gen-worker requeue --config ./worker-config.json
```

Queued job files move through:

- `jobs/new/` for newly created jobs
- `jobs/pending/` once a worker claims a job
- `jobs/completed/` after a successful run
- `jobs/failed/` after exhausting retry attempts or invalid job input

Moving `jobs/new/` to `jobs/pending/` happens before execution to reduce the chance of double-running the same job.

Retry behavior:

- Jobs track `retry_count` and `max_retries` in JSON
- Worker failures requeue the job to `jobs/new/` until retry limit is reached
- Once retries are exhausted, jobs move to `jobs/failed/`

Recovery behavior:

- `run-next` and `run-queue` automatically move stale `jobs/pending/` entries back to `jobs/new/` (configurable with `--pending-timeout`)
- `requeue` can manually move jobs from `jobs/failed/` and/or `jobs/pending/` back to `jobs/new/`
- New job creation rejects duplicates already present in `jobs/new/` or `jobs/pending/` for the same scene/checksum.

Retention policy:

- Completed and failed job files are retained in their directories until manually cleaned up.
