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

	b, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read job file: %v", err)
	}

	var job worker.Job
	if err := json.Unmarshal(b, &job); err != nil {
		t.Fatalf("unmarshal job: %v", err)
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

	queuePath := filepath.Join(jobsDir, "new", "sample.json")
	if err := os.MkdirAll(filepath.Dir(queuePath), 0o755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}
	b, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}
	if err := os.WriteFile(queuePath, b, 0o644); err != nil {
		t.Fatalf("write queued job: %v", err)
	}

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
