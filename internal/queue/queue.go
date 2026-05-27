package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mr-szgz/stash-gen-worker/internal/worker"
)

type Queue struct {
	root         string
	newDir       string
	pendingDir   string
	completedDir string
	failedDir    string
}

type QueuedJob struct {
	queue Queue
	Path  string
	Name  string
	Job   worker.Job
}

func New(root string) Queue {
	return Queue{
		root:         root,
		newDir:       filepath.Join(root, "new"),
		pendingDir:   filepath.Join(root, "pending"),
		completedDir: filepath.Join(root, "completed"),
		failedDir:    filepath.Join(root, "failed"),
	}
}

func (q Queue) Ensure() error {
	for _, dir := range []string{q.root, q.newDir, q.pendingDir, q.completedDir, q.failedDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating jobs directory %s: %w", dir, err)
		}
	}
	return nil
}

func (q Queue) Enqueue(job worker.Job, nameHint string) (string, error) {
	if err := q.Ensure(); err != nil {
		return "", err
	}

	name := q.jobFileName(nameHint)
	target := filepath.Join(q.newDir, name)
	tmp, err := os.CreateTemp(q.newDir, "job-*.json")
	if err != nil {
		return "", fmt.Errorf("creating temporary job file: %w", err)
	}

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(job); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("writing job file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("closing job file: %w", err)
	}

	if err := os.Rename(tmp.Name(), target); err != nil {
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("moving job file into queue: %w", err)
	}

	return target, nil
}

func (q Queue) AcquireNext() (*QueuedJob, error) {
	if err := q.Ensure(); err != nil {
		return nil, err
	}

	entries, err := filepath.Glob(filepath.Join(q.newDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("listing new jobs: %w", err)
	}
	sort.Strings(entries)

	for _, path := range entries {
		name := filepath.Base(path)
		pendingPath := filepath.Join(q.pendingDir, name)
		if err := os.Rename(path, pendingPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("moving job to pending: %w", err)
		}

		job, err := readJob(pendingPath)
		if err != nil {
			failedPath := filepath.Join(q.failedDir, name)
			moveErr := os.Rename(pendingPath, failedPath)
			if moveErr != nil {
				return nil, fmt.Errorf("reading pending job: %w (moving to failed: %v)", err, moveErr)
			}
			return nil, fmt.Errorf("reading pending job: %w", err)
		}

		return &QueuedJob{
			queue: q,
			Path:  pendingPath,
			Name:  name,
			Job:   job,
		}, nil
	}

	return nil, nil
}

func (j QueuedJob) MarkCompleted() (string, error) {
	return j.move(filepath.Join(j.queue.completedDir, j.Name))
}

func (j QueuedJob) MarkFailed() (string, error) {
	return j.move(filepath.Join(j.queue.failedDir, j.Name))
}

func (j QueuedJob) move(target string) (string, error) {
	if err := os.Rename(j.Path, target); err != nil {
		return "", fmt.Errorf("moving job file: %w", err)
	}
	return target, nil
}

func (q Queue) jobFileName(nameHint string) string {
	timestamp := time.Now().UTC().Format("20060102T150405.000Z")
	name := sanitizeName(nameHint)
	if name == "" {
		name = "job"
	}
	return fmt.Sprintf("%s-%s.json", timestamp, name)
}

func sanitizeName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-', r == '_':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func readJob(path string) (worker.Job, error) {
	var job worker.Job

	b, err := os.ReadFile(path)
	if err != nil {
		return worker.Job{}, fmt.Errorf("reading job file: %w", err)
	}
	if err := json.Unmarshal(b, &job); err != nil {
		return worker.Job{}, fmt.Errorf("parsing job file: %w", err)
	}

	return job, nil
}
