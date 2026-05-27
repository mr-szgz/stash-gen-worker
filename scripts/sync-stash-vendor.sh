#!/usr/bin/env bash
set -euo pipefail

UPSTREAM_REPO="${UPSTREAM_REPO:-https://raw.githubusercontent.com/stashapp/stash}"
UPSTREAM_REF="${UPSTREAM_REF:-4187d164b349f8442a4f31c72bb477302590a9a4}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

files=(
  pkg/scene/generate/generator.go
  pkg/scene/generate/preview.go
  pkg/scene/generate/screenshot.go
  pkg/scene/generate/sprite.go
  pkg/scene/generate/transcode.go
  pkg/ffmpeg/codec.go
  pkg/ffmpeg/downloader.go
  pkg/ffmpeg/ffmpeg.go
  pkg/ffmpeg/ffprobe.go
  pkg/ffmpeg/filter.go
  pkg/ffmpeg/format.go
  pkg/ffmpeg/options.go
  pkg/ffmpeg/types.go
  pkg/ffmpeg/transcoder/screenshot.go
  pkg/ffmpeg/transcoder/splice.go
  pkg/ffmpeg/transcoder/transcode.go
  pkg/fsutil/lock_manager.go
  pkg/exec/command.go
  pkg/exec/shell_nonwindows.go
  pkg/exec/shell_windows.go
  pkg/utils/date.go
  pkg/utils/vtt.go
  pkg/models/json/json_time.go
)

for file in "${files[@]}"; do
  dest="$REPO_ROOT/third_party/stash/$file"
  mkdir -p "$(dirname "$dest")"
  curl -fsSL "$UPSTREAM_REPO/$UPSTREAM_REF/$file" -o "$dest"
done

python <<'PY' "$REPO_ROOT"
from pathlib import Path
import sys
root = Path(sys.argv[1])
old = 'github.com/stashapp/stash/pkg/'
new = 'github.com/mr-szgz/stash-gen-worker/third_party/stash/pkg/'
for path in (root / 'third_party' / 'stash').rglob('*.go'):
    text = path.read_text()
    if old in text:
        path.write_text(text.replace(old, new))
PY

echo "Synced vendored Stash files from ${UPSTREAM_REF}."
