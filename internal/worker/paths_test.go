package worker

import (
	"path/filepath"
	"testing"
)

func TestOutputPaths(t *testing.T) {
	paths := NewOutputPaths("generated")

	if got, want := paths.Screenshot("abc"), filepath.Join("generated", "screenshots", "abc.jpg"); got != want {
		t.Fatalf("screenshot path = %q, want %q", got, want)
	}
	if got, want := paths.Preview("abc"), filepath.Join("generated", "screenshots", "abc.mp4"); got != want {
		t.Fatalf("preview path = %q, want %q", got, want)
	}
	if got, want := paths.WebP("abc"), filepath.Join("generated", "screenshots", "abc.webp"); got != want {
		t.Fatalf("webp path = %q, want %q", got, want)
	}
	if got, want := paths.SpriteImage("abc"), filepath.Join("generated", "vtt", "abc_sprite.jpg"); got != want {
		t.Fatalf("sprite image path = %q, want %q", got, want)
	}
	if got, want := paths.SpriteVTT("abc"), filepath.Join("generated", "vtt", "abc_thumbs.vtt"); got != want {
		t.Fatalf("sprite vtt path = %q, want %q", got, want)
	}
	if got, want := paths.Transcode("abc"), filepath.Join("generated", "transcodes", "abc.mp4"); got != want {
		t.Fatalf("transcode path = %q, want %q", got, want)
	}
}
