param(
    [string]$UpstreamRepo = "https://raw.githubusercontent.com/stashapp/stash",
    [string]$UpstreamRef = "4187d164b349f8442a4f31c72bb477302590a9a4"
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$Files = @(
    "pkg/scene/generate/generator.go",
    "pkg/scene/generate/preview.go",
    "pkg/scene/generate/screenshot.go",
    "pkg/scene/generate/sprite.go",
    "pkg/scene/generate/transcode.go",
    "pkg/ffmpeg/codec.go",
    "pkg/ffmpeg/downloader.go",
    "pkg/ffmpeg/ffmpeg.go",
    "pkg/ffmpeg/ffprobe.go",
    "pkg/ffmpeg/filter.go",
    "pkg/ffmpeg/format.go",
    "pkg/ffmpeg/options.go",
    "pkg/ffmpeg/types.go",
    "pkg/ffmpeg/transcoder/screenshot.go",
    "pkg/ffmpeg/transcoder/splice.go",
    "pkg/ffmpeg/transcoder/transcode.go",
    "pkg/fsutil/lock_manager.go",
    "pkg/exec/command.go",
    "pkg/exec/shell_nonwindows.go",
    "pkg/exec/shell_windows.go",
    "pkg/utils/date.go",
    "pkg/utils/vtt.go",
    "pkg/models/json/json_time.go"
)

foreach ($File in $Files) {
    $Destination = Join-Path $RepoRoot (Join-Path "third_party/stash" $File)
    New-Item -ItemType Directory -Force -Path (Split-Path $Destination) | Out-Null
    Invoke-WebRequest -Uri "$UpstreamRepo/$UpstreamRef/$File" -OutFile $Destination
}

$Old = 'github.com/stashapp/stash/pkg/'
$New = 'github.com/mr-szgz/stash-gen-worker/third_party/stash/pkg/'
Get-ChildItem -Path (Join-Path $RepoRoot 'third_party/stash') -Filter *.go -Recurse | ForEach-Object {
    $Content = Get-Content $_.FullName -Raw
    if ($Content.Contains($Old)) {
        $Content.Replace($Old, $New) | Set-Content -NoNewline $_.FullName
    }
}

Write-Host "Synced vendored Stash files from $UpstreamRef."
