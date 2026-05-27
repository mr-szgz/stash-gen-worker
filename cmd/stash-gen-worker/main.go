package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	stashffmpeg "github.com/stashapp/stash/pkg/ffmpeg"

	"github.com/mr-szgz/stash-gen-worker/internal/queue"
	"github.com/mr-szgz/stash-gen-worker/internal/stashgraphql"
	"github.com/mr-szgz/stash-gen-worker/internal/worker"
)

var executeJob = runWorkerJob

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "generate-job":
			return runGenerateJob(args[1:], stdout)
		case "run-next":
			return runNextJob(args[1:], stdout)
		case "run-queue":
			return runQueue(args[1:], stdout)
		default:
			return fmt.Errorf("unknown command %q", args[0])
		}
	}

	return runSingleJob(args, stdout)
}

func runSingleJob(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("stash-gen-worker", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var configPath string
	var jobFile string
	var inputPath string
	var checksum string
	var generatedDir string
	var ffmpegPath string
	var ffprobePath string
	var preview bool
	var webp bool
	var screenshot bool
	var sprite bool
	var transcode bool
	var overwrite bool

	fs.StringVar(&configPath, "config", "", "path to worker config JSON file")
	fs.StringVar(&jobFile, "job", "", "path to JSON job file")
	fs.StringVar(&inputPath, "input", "", "input scene path")
	fs.StringVar(&checksum, "checksum", "", "scene checksum")
	fs.StringVar(&generatedDir, "generated", "", "generated output root")
	fs.StringVar(&ffmpegPath, "ffmpeg", "", "path to ffmpeg executable")
	fs.StringVar(&ffprobePath, "ffprobe", "", "path to ffprobe executable")
	fs.BoolVar(&preview, "preview", false, "generate preview mp4")
	fs.BoolVar(&webp, "webp", false, "generate preview webp")
	fs.BoolVar(&screenshot, "screenshot", false, "generate screenshot jpg")
	fs.BoolVar(&sprite, "sprite", false, "generate sprite jpg and vtt")
	fs.BoolVar(&transcode, "transcode", false, "generate transcode mp4")
	fs.BoolVar(&overwrite, "overwrite", true, "overwrite existing files")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := worker.LoadConfig(configPath)
	if err != nil {
		return err
	}

	resolvedGeneratedDir := firstNonEmpty(generatedDir, cfg.GeneratedDir, worker.DefaultConfig().GeneratedDir)
	resolvedFFMpeg := firstNonEmpty(ffmpegPath, cfg.FFMpegPath)
	resolvedFFProbe := firstNonEmpty(ffprobePath, cfg.FFProbePath)

	var job worker.Job
	if jobFile != "" {
		job, err = readJobFile(jobFile)
		if err != nil {
			return err
		}
	} else {
		job = worker.DefaultJob()
		job.InputPath = inputPath
		job.Checksum = checksum
		job.GeneratedDir = resolvedGeneratedDir
		job.Preview = preview
		job.WebP = webp
		job.Screenshot = screenshot
		job.Sprite = sprite
		job.Transcode = transcode
		job.Overwrite = overwrite
	}

	job.ApplyDefaults(resolvedGeneratedDir)
	if err := validateJob(job); err != nil {
		return err
	}

	if err := executeJob(context.Background(), job, resolvedFFMpeg, resolvedFFProbe, stdout); err != nil {
		return err
	}

	return nil
}

func runGenerateJob(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("generate-job", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var configPath string
	var sceneID string
	var endpoint string
	var apiKey string
	var jobsDir string
	var generatedDir string
	var preview bool
	var webp bool
	var screenshot bool
	var sprite bool
	var transcode bool
	var overwrite bool

	fs.StringVar(&configPath, "config", "", "path to worker config JSON file")
	fs.StringVar(&sceneID, "scene-id", "", "stash scene ID")
	fs.StringVar(&endpoint, "stash-url", "", "stash GraphQL endpoint URL")
	fs.StringVar(&apiKey, "stash-api-key", "", "stash API key")
	fs.StringVar(&jobsDir, "jobs-dir", "", "worker jobs root directory")
	fs.StringVar(&generatedDir, "generated", "", "generated output root to embed in the job")
	fs.BoolVar(&preview, "preview", false, "generate preview mp4")
	fs.BoolVar(&webp, "webp", false, "generate preview webp")
	fs.BoolVar(&screenshot, "screenshot", false, "generate screenshot jpg")
	fs.BoolVar(&sprite, "sprite", false, "generate sprite jpg and vtt")
	fs.BoolVar(&transcode, "transcode", false, "generate transcode mp4")
	fs.BoolVar(&overwrite, "overwrite", true, "overwrite existing files")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := worker.LoadConfig(configPath)
	if err != nil {
		return err
	}

	resolvedEndpoint := firstNonEmpty(endpoint, cfg.StashGraphQLEndpoint)
	resolvedAPIKey := firstNonEmpty(apiKey, cfg.StashAPIKey)
	resolvedJobsDir := firstNonEmpty(jobsDir, cfg.JobsDir, worker.DefaultConfig().JobsDir)
	resolvedGeneratedDir := firstNonEmpty(generatedDir, cfg.GeneratedDir, worker.DefaultConfig().GeneratedDir)

	scene, err := stashgraphql.Client{
		Endpoint: resolvedEndpoint,
		APIKey:   resolvedAPIKey,
	}.FetchScene(context.Background(), sceneID)
	if err != nil {
		return err
	}

	job := worker.DefaultJob()
	job.SceneID = scene.ID
	job.SceneTitle = scene.Title
	job.InputPath = scene.File.Path
	job.Checksum = scene.File.Checksum
	job.GeneratedDir = resolvedGeneratedDir
	job.Overwrite = overwrite

	if generationFlagsSet(fs) {
		job.Preview = preview
		job.WebP = webp
		job.Screenshot = screenshot
		job.Sprite = sprite
		job.Transcode = transcode
	} else {
		job.Preview = true
		job.WebP = true
		job.Screenshot = true
		job.Sprite = true
	}

	q := queue.New(resolvedJobsDir)
	path, err := q.Enqueue(job, fmt.Sprintf("scene-%s-%s", scene.ID, scene.Title))
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "queued job: %s\n", path)
	return nil
}

func runNextJob(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("run-next", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var configPath string
	var jobsDir string
	var generatedDir string
	var ffmpegPath string
	var ffprobePath string

	fs.StringVar(&configPath, "config", "", "path to worker config JSON file")
	fs.StringVar(&jobsDir, "jobs-dir", "", "worker jobs root directory")
	fs.StringVar(&generatedDir, "generated", "", "default generated output root")
	fs.StringVar(&ffmpegPath, "ffmpeg", "", "path to ffmpeg executable")
	fs.StringVar(&ffprobePath, "ffprobe", "", "path to ffprobe executable")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := worker.LoadConfig(configPath)
	if err != nil {
		return err
	}

	return processNextQueuedJob(
		firstNonEmpty(jobsDir, cfg.JobsDir, worker.DefaultConfig().JobsDir),
		firstNonEmpty(generatedDir, cfg.GeneratedDir, worker.DefaultConfig().GeneratedDir),
		firstNonEmpty(ffmpegPath, cfg.FFMpegPath),
		firstNonEmpty(ffprobePath, cfg.FFProbePath),
		stdout,
	)
}

func runQueue(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("run-queue", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var configPath string
	var jobsDir string
	var generatedDir string
	var ffmpegPath string
	var ffprobePath string

	fs.StringVar(&configPath, "config", "", "path to worker config JSON file")
	fs.StringVar(&jobsDir, "jobs-dir", "", "worker jobs root directory")
	fs.StringVar(&generatedDir, "generated", "", "default generated output root")
	fs.StringVar(&ffmpegPath, "ffmpeg", "", "path to ffmpeg executable")
	fs.StringVar(&ffprobePath, "ffprobe", "", "path to ffprobe executable")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := worker.LoadConfig(configPath)
	if err != nil {
		return err
	}

	resolvedJobsDir := firstNonEmpty(jobsDir, cfg.JobsDir, worker.DefaultConfig().JobsDir)
	resolvedGeneratedDir := firstNonEmpty(generatedDir, cfg.GeneratedDir, worker.DefaultConfig().GeneratedDir)
	resolvedFFMpeg := firstNonEmpty(ffmpegPath, cfg.FFMpegPath)
	resolvedFFProbe := firstNonEmpty(ffprobePath, cfg.FFProbePath)

	processed := 0
	for {
		ok, err := processNextQueuedJob(resolvedJobsDir, resolvedGeneratedDir, resolvedFFMpeg, resolvedFFProbe, stdout)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		processed++
	}

	_, _ = fmt.Fprintf(stdout, "processed jobs: %d\n", processed)
	return nil
}

func processNextQueuedJob(jobsDir string, defaultGeneratedDir string, ffmpegPath string, ffprobePath string, stdout io.Writer) (bool, error) {
	q := queue.New(jobsDir)
	queuedJob, err := q.AcquireNext()
	if err != nil {
		return false, err
	}
	if queuedJob == nil {
		_, _ = fmt.Fprintln(stdout, "no queued jobs")
		return false, nil
	}

	queuedJob.Job.ApplyDefaults(defaultGeneratedDir)
	if err := validateJob(queuedJob.Job); err != nil {
		if _, moveErr := queuedJob.MarkFailed(); moveErr != nil {
			return false, fmt.Errorf("%w (moving failed job: %v)", err, moveErr)
		}
		return false, err
	}

	if err := executeJob(context.Background(), queuedJob.Job, ffmpegPath, ffprobePath, stdout); err != nil {
		failedPath, moveErr := queuedJob.MarkFailed()
		if moveErr != nil {
			return false, fmt.Errorf("%w (moving failed job: %v)", err, moveErr)
		}
		return false, fmt.Errorf("%w (job moved to %s)", err, failedPath)
	}

	completedPath, err := queuedJob.MarkCompleted()
	if err != nil {
		return false, err
	}
	_, _ = fmt.Fprintf(stdout, "completed job: %s\n", completedPath)
	return true, nil
}

func runWorkerJob(ctx context.Context, job worker.Job, ffmpegPath string, ffprobePath string, stdout io.Writer) error {
	resolvedFFMpeg := stashffmpeg.ResolveFFMpeg(ffmpegPath, filepath.Dir(os.Args[0]))
	if resolvedFFMpeg == "" {
		return errors.New("ffmpeg not found")
	}

	resolvedFFProbe := stashffmpeg.ResolveFFProbe(ffprobePath, filepath.Dir(os.Args[0]))
	if resolvedFFProbe == "" {
		return errors.New("ffprobe not found")
	}

	probe := stashffmpeg.NewFFProbe(resolvedFFProbe)
	videoFile, err := probe.NewVideoFile(job.InputPath)
	if err != nil {
		return fmt.Errorf("probing video: %w", err)
	}

	if err := worker.Run(ctx, worker.RunRequest{
		Job:        job,
		VideoFile:  videoFile,
		FFMpegPath: resolvedFFMpeg,
	}); err != nil {
		return err
	}

	if job.Screenshot {
		_, _ = fmt.Fprintf(stdout, "screenshot: %s\n", filepath.Join(job.GeneratedDir, "screenshots", job.Checksum+".jpg"))
	}
	if job.Preview {
		_, _ = fmt.Fprintf(stdout, "preview: %s\n", filepath.Join(job.GeneratedDir, "screenshots", job.Checksum+".mp4"))
	}
	if job.WebP {
		_, _ = fmt.Fprintf(stdout, "webp: %s\n", filepath.Join(job.GeneratedDir, "screenshots", job.Checksum+".webp"))
	}
	if job.Sprite {
		_, _ = fmt.Fprintf(stdout, "sprite image: %s\n", filepath.Join(job.GeneratedDir, "vtt", job.Checksum+"_sprite.jpg"))
		_, _ = fmt.Fprintf(stdout, "sprite vtt: %s\n", filepath.Join(job.GeneratedDir, "vtt", job.Checksum+"_thumbs.vtt"))
	}
	if job.Transcode {
		_, _ = fmt.Fprintf(stdout, "transcode: %s\n", filepath.Join(job.GeneratedDir, "transcodes", job.Checksum+".mp4"))
	}

	return nil
}

func readJobFile(path string) (worker.Job, error) {
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

func validateJob(job worker.Job) error {
	if strings.TrimSpace(job.InputPath) == "" {
		return errors.New("input path is required")
	}
	if strings.TrimSpace(job.Checksum) == "" {
		return errors.New("checksum is required")
	}
	if strings.TrimSpace(job.GeneratedDir) == "" {
		return errors.New("generated dir is required")
	}
	return nil
}

func generationFlagsSet(fs *flag.FlagSet) bool {
	set := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "preview", "webp", "screenshot", "sprite", "transcode":
			set = true
		}
	})
	return set
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
