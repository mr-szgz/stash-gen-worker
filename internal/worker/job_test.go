package worker

import "testing"

func TestJobApplyDefaults(t *testing.T) {
	t.Parallel()

	job := Job{}
	job.ApplyDefaults("/srv/stash/generated")

	if job.GeneratedDir != "/srv/stash/generated" {
		t.Fatalf("expected generated dir default, got %q", job.GeneratedDir)
	}
	if job.SchemaVersion != CurrentJobSchemaVersion {
		t.Fatalf("expected schema version %d, got %d", CurrentJobSchemaVersion, job.SchemaVersion)
	}
	if job.PreviewOptions.Segments != 12 {
		t.Fatalf("expected default preview segments, got %d", job.PreviewOptions.Segments)
	}
	if job.PreviewOptions.Preset != "veryfast" {
		t.Fatalf("expected default preview preset, got %q", job.PreviewOptions.Preset)
	}
	if job.SpriteOptions.Count != 25 {
		t.Fatalf("expected default sprite count, got %d", job.SpriteOptions.Count)
	}
	if job.SpriteOptions.Size != 320 {
		t.Fatalf("expected default sprite size, got %d", job.SpriteOptions.Size)
	}
	if job.MaxRetries != DefaultMaxRetries {
		t.Fatalf("expected default max retries %d, got %d", DefaultMaxRetries, job.MaxRetries)
	}
}

func TestPreviewOptionsToStashAppliesDefaults(t *testing.T) {
	t.Parallel()

	got := (PreviewOptions{}).ToStash()
	if got.Segments != 12 {
		t.Fatalf("expected default segments, got %d", got.Segments)
	}
	if got.SegmentDuration != 0.5 {
		t.Fatalf("expected default segment duration, got %f", got.SegmentDuration)
	}
	if got.Preset != "veryfast" {
		t.Fatalf("expected default preset, got %q", got.Preset)
	}
}
