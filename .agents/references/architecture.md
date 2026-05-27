# Architecture and repository intent

This repository is a standalone Go worker CLI that generates Stash-compatible scene assets without requiring a full Stash server checkout.

## Current shape

- `cmd/stash-gen-worker/` contains the CLI entrypoint.
- `internal/worker/` contains job parsing, output path conventions, and orchestration around the vendored generator.
- `third_party/stash/` contains the minimal upstream Stash dependency cone copied into this repository.

## Design intent

- Keep the worker self-contained so it builds in CI and on distributed worker nodes.
- Reuse upstream Stash asset-generation logic instead of reimplementing it.
- Keep local glue small and obvious so future upstream syncs stay manageable.

## Output layout

Generated assets are written under the configured generated directory using Stash-like subdirectories:

- `screenshots/<checksum>.jpg`
- `screenshots/<checksum>.mp4`
- `screenshots/<checksum>.webp`
- `vtt/<checksum>_sprite.jpg`
- `vtt/<checksum>_thumbs.vtt`
- `transcodes/<checksum>.mp4`
