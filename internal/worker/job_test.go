package worker

import "testing"

func TestPreviewOptionsToStashDefaults(t *testing.T) {
	got := (PreviewOptions{}).ToStash()
	if got.Segments != 12 {
		t.Fatalf("segments = %d, want 12", got.Segments)
	}
	if got.SegmentDuration != 0.5 {
		t.Fatalf("segment duration = %v, want 0.5", got.SegmentDuration)
	}
	if got.Preset != "veryfast" {
		t.Fatalf("preset = %q, want veryfast", got.Preset)
	}
}
