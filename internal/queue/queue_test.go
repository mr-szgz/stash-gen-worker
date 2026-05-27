package queue

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestQueueEnqueueRejectsDuplicateActiveJob(t *testing.T) {
	t.Parallel()

	q := New(t.TempDir())
	job := worker.DefaultJob()
	job.SceneID = "123"
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"

	if _, err := q.Enqueue(job, "first"); err != nil {
		t.Fatalf("enqueue first job: %v", err)
	}
	if _, err := q.Enqueue(job, "duplicate"); !errors.Is(err, ErrDuplicateJob) {
		t.Fatalf("expected duplicate job error, got %v", err)
	}
}

func TestQueueAcquireMalformedJobMovesToFailed(t *testing.T) {
	t.Parallel()

	q := New(t.TempDir())
	if err := q.Ensure(); err != nil {
		t.Fatalf("ensure queue: %v", err)
	}

	path := filepath.Join(q.newDir, "broken.json")
	if err := os.WriteFile(path, []byte(`{"scene_id":"123"`), 0o644); err != nil {
		t.Fatalf("write malformed job: %v", err)
	}

	queued, err := q.AcquireNext()
	if err == nil {
		t.Fatalf("expected acquire error for malformed JSON")
	}
	if queued != nil {
		t.Fatalf("expected no queued job on malformed JSON")
	}

	failedPath := filepath.Join(q.failedDir, "broken.json")
	if _, statErr := os.Stat(failedPath); statErr != nil {
		t.Fatalf("expected malformed job to move to failed: %v", statErr)
	}
}

func TestRecoverPendingStaleMovesOnlyOldJobs(t *testing.T) {
	t.Parallel()

	q := New(t.TempDir())
	if err := q.Ensure(); err != nil {
		t.Fatalf("ensure queue: %v", err)
	}

	oldPath := filepath.Join(q.pendingDir, "old.json")
	newPath := filepath.Join(q.pendingDir, "new.json")
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	if err := writeJobAtomic(oldPath, job); err != nil {
		t.Fatalf("write old pending job: %v", err)
	}
	if err := writeJobAtomic(newPath, job); err != nil {
		t.Fatalf("write new pending job: %v", err)
	}

	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("set old timestamp: %v", err)
	}

	recovered, err := q.RecoverPendingStale(30 * time.Minute)
	if err != nil {
		t.Fatalf("recover stale pending: %v", err)
	}
	if recovered != 1 {
		t.Fatalf("expected one recovered pending job, got %d", recovered)
	}

	requeued, err := filepath.Glob(filepath.Join(q.newDir, "*.json"))
	if err != nil {
		t.Fatalf("glob requeued jobs: %v", err)
	}
	if len(requeued) != 1 {
		t.Fatalf("expected one job in new queue, got %d", len(requeued))
	}
	if _, statErr := os.Stat(newPath); statErr != nil {
		t.Fatalf("expected fresh pending job to remain pending: %v", statErr)
	}
}

func TestRequeueFailed(t *testing.T) {
	t.Parallel()

	q := New(t.TempDir())
	if err := q.Ensure(); err != nil {
		t.Fatalf("ensure queue: %v", err)
	}

	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	if err := writeJobAtomic(filepath.Join(q.failedDir, "failed-job.json"), job); err != nil {
		t.Fatalf("write failed job: %v", err)
	}

	count, err := q.RequeueFailed(0)
	if err != nil {
		t.Fatalf("requeue failed jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one failed job requeued, got %d", count)
	}

	newJobs, err := filepath.Glob(filepath.Join(q.newDir, "*.json"))
	if err != nil {
		t.Fatalf("glob new jobs: %v", err)
	}
	if len(newJobs) != 1 {
		t.Fatalf("expected one job in new dir after requeue, got %d", len(newJobs))
	}
}
