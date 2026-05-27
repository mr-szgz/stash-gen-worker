package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mr-szgz/stash-gen-worker/internal/worker"
)

func TestGenerateJobQueuesSceneFromGraphQL(t *testing.T) {
	t.Parallel()

	jobsDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"findScene": {
					"id": "123",
					"title": "Sample Scene",
					"files": [
						{"path": "/library/sample.mp4", "fingerprint": "abc123"}
					]
				}
			}
		}`))
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{
		"generate-job",
		"--scene-id", "123",
		"--stash-url", server.URL,
		"--jobs-dir", jobsDir,
		"--generated", "/srv/stash/generated",
	}, &stdout)
	if err != nil {
		t.Fatalf("run generate-job: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(jobsDir, "new", "*.json"))
	if err != nil {
		t.Fatalf("glob job files: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 queued job, got %d", len(matches))
	}

	var job worker.Job
	readJobFromPath(t, matches[0], &job)

	if job.SchemaVersion != worker.CurrentJobSchemaVersion {
		t.Fatalf("expected schema version %d, got %d", worker.CurrentJobSchemaVersion, job.SchemaVersion)
	}
	if job.SceneID != "123" {
		t.Fatalf("expected scene id 123, got %q", job.SceneID)
	}
	if job.InputPath != "/library/sample.mp4" {
		t.Fatalf("expected input path to be queued, got %q", job.InputPath)
	}
	if job.Checksum != "abc123" {
		t.Fatalf("expected checksum abc123, got %q", job.Checksum)
	}
	if job.GeneratedDir != "/srv/stash/generated" {
		t.Fatalf("expected generated dir override, got %q", job.GeneratedDir)
	}
	if !job.Preview || !job.WebP || !job.Screenshot || !job.Sprite {
		t.Fatalf("expected default asset generation flags to be enabled: %+v", job)
	}
	if job.Transcode {
		t.Fatalf("expected transcode to remain disabled by default")
	}
}

func TestRunNextMovesJobToCompleted(t *testing.T) {
	jobsDir := t.TempDir()
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	job.GeneratedDir = ""

	writeQueuedJob(t, filepath.Join(jobsDir, "new", "sample.json"), job)

	originalExecute := executeJob
	t.Cleanup(func() {
		executeJob = originalExecute
	})

	var gotJob worker.Job
	var stdout bytes.Buffer
	executeJob = func(_ context.Context, job worker.Job, _, _ string, _ io.Writer) error {
		gotJob = job
		return nil
	}

	if err := run([]string{"run-next", "--jobs-dir", jobsDir, "--generated", "/srv/stash/generated"}, &stdout); err != nil {
		t.Fatalf("run run-next: %v", err)
	}

	if gotJob.GeneratedDir != "/srv/stash/generated" {
		t.Fatalf("expected generated dir default to be applied, got %q", gotJob.GeneratedDir)
	}

	completed, err := filepath.Glob(filepath.Join(jobsDir, "completed", "*.json"))
	if err != nil {
		t.Fatalf("glob completed jobs: %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("expected job to move to completed, got %d files", len(completed))
	}
}

func TestRunNextRetriesThenCompletes(t *testing.T) {
	jobsDir := t.TempDir()
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	job.MaxRetries = 2
	writeQueuedJob(t, filepath.Join(jobsDir, "new", "sample.json"), job)

	originalExecute := executeJob
	t.Cleanup(func() {
		executeJob = originalExecute
	})

	attempts := 0
	executeJob = func(_ context.Context, _ worker.Job, _, _ string, _ io.Writer) error {
		attempts++
		if attempts == 1 {
			return io.ErrUnexpectedEOF
		}
		return nil
	}

	if err := run([]string{"run-next", "--jobs-dir", jobsDir}, io.Discard); err != nil {
		t.Fatalf("first run-next: %v", err)
	}

	newJobs, err := filepath.Glob(filepath.Join(jobsDir, "new", "*.json"))
	if err != nil {
		t.Fatalf("glob new jobs: %v", err)
	}
	if len(newJobs) != 1 {
		t.Fatalf("expected failed attempt to requeue one job, got %d", len(newJobs))
	}

	var retried worker.Job
	readJobFromPath(t, newJobs[0], &retried)
	if retried.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", retried.RetryCount)
	}
	if retried.LastError == "" {
		t.Fatalf("expected last error to be populated")
	}

	if err := run([]string{"run-next", "--jobs-dir", jobsDir}, io.Discard); err != nil {
		t.Fatalf("second run-next: %v", err)
	}

	completed, err := filepath.Glob(filepath.Join(jobsDir, "completed", "*.json"))
	if err != nil {
		t.Fatalf("glob completed jobs: %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("expected job to complete on retry, got %d", len(completed))
	}
}

func TestRunNextMovesToFailedAfterRetryLimit(t *testing.T) {
	jobsDir := t.TempDir()
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	job.MaxRetries = 1
	writeQueuedJob(t, filepath.Join(jobsDir, "new", "sample.json"), job)

	originalExecute := executeJob
	t.Cleanup(func() {
		executeJob = originalExecute
	})

	executeJob = func(_ context.Context, _ worker.Job, _, _ string, _ io.Writer) error {
		return io.ErrClosedPipe
	}

	if err := run([]string{"run-next", "--jobs-dir", jobsDir}, io.Discard); err != nil {
		t.Fatalf("first run-next: %v", err)
	}
	if err := run([]string{"run-next", "--jobs-dir", jobsDir}, io.Discard); err == nil {
		t.Fatalf("second run-next should fail after retry limit")
	}

	failed, err := filepath.Glob(filepath.Join(jobsDir, "failed", "*.json"))
	if err != nil {
		t.Fatalf("glob failed jobs: %v", err)
	}
	if len(failed) != 1 {
		t.Fatalf("expected one failed job, got %d", len(failed))
	}

	var failedJob worker.Job
	readJobFromPath(t, failed[0], &failedJob)
	if failedJob.RetryCount != 2 {
		t.Fatalf("expected retry count 2 after two failures, got %d", failedJob.RetryCount)
	}
}

func TestRunNextRecoversStalePending(t *testing.T) {
	jobsDir := t.TempDir()
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	pendingPath := filepath.Join(jobsDir, "pending", "stale.json")
	writeQueuedJob(t, pendingPath, job)

	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(pendingPath, old, old); err != nil {
		t.Fatalf("set stale pending timestamp: %v", err)
	}

	originalExecute := executeJob
	t.Cleanup(func() {
		executeJob = originalExecute
	})
	executeJob = func(_ context.Context, _ worker.Job, _, _ string, _ io.Writer) error {
		return nil
	}

	if err := run([]string{"run-next", "--jobs-dir", jobsDir, "--pending-timeout", "30m"}, io.Discard); err != nil {
		t.Fatalf("run-next with pending recovery: %v", err)
	}

	completed, err := filepath.Glob(filepath.Join(jobsDir, "completed", "*.json"))
	if err != nil {
		t.Fatalf("glob completed jobs: %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("expected stale pending job to be recovered and completed, got %d", len(completed))
	}
}

func TestRequeueCommandMovesFailedToNew(t *testing.T) {
	jobsDir := t.TempDir()
	job := worker.DefaultJob()
	job.InputPath = "/library/sample.mp4"
	job.Checksum = "abc123"
	writeQueuedJob(t, filepath.Join(jobsDir, "failed", "failed.json"), job)

	if err := run([]string{"requeue", "--jobs-dir", jobsDir, "--failed"}, io.Discard); err != nil {
		t.Fatalf("run requeue command: %v", err)
	}

	newJobs, err := filepath.Glob(filepath.Join(jobsDir, "new", "*.json"))
	if err != nil {
		t.Fatalf("glob requeued jobs: %v", err)
	}
	if len(newJobs) != 1 {
		t.Fatalf("expected one requeued job in new queue, got %d", len(newJobs))
	}
}

func writeQueuedJob(t *testing.T, path string, job worker.Job) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir queue dir: %v", err)
	}
	b, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal queued job: %v", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write queued job: %v", err)
	}
}

func readJobFromPath(t *testing.T, path string, out *worker.Job) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read job file: %v", err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("unmarshal job: %v", err)
	}
}
