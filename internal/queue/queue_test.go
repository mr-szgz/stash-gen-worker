package queue

import (
	"path/filepath"
	"testing"

	"github.com/mr-szgz/stash-gen-worker/internal/worker"
)

func TestQueueAcquireAndComplete(t *testing.T) {
	t.Parallel()

	q := New(t.TempDir())
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"

	path, err := q.Enqueue(job, "Sample Scene")
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}
	if filepath.Base(path) == "" {
		t.Fatalf("expected queued file path")
	}

	queued, err := q.AcquireNext()
	if err != nil {
		t.Fatalf("acquire next: %v", err)
	}
	if queued == nil {
		t.Fatalf("expected queued job")
	}
	if queued.Job.InputPath != job.InputPath {
		t.Fatalf("expected input path %q, got %q", job.InputPath, queued.Job.InputPath)
	}

	completedPath, err := queued.MarkCompleted()
	if err != nil {
		t.Fatalf("mark completed: %v", err)
	}
	if filepath.Dir(completedPath) != filepath.Join(q.root, "completed") {
		t.Fatalf("expected completed dir, got %q", completedPath)
	}
}
