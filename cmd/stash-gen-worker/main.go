package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	stashffmpeg "github.com/mr-szgz/stash-gen-worker/third_party/stash/pkg/ffmpeg"

	"github.com/mr-szgz/stash-gen-worker/internal/worker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
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

	flag.StringVar(&jobFile, "job", "", "path to JSON job file")
	flag.StringVar(&inputPath, "input", "", "input scene path")
	flag.StringVar(&checksum, "checksum", "", "scene checksum")
	flag.StringVar(&generatedDir, "generated", "./generated", "generated output root")
	flag.StringVar(&ffmpegPath, "ffmpeg", "", "path to ffmpeg executable")
	flag.StringVar(&ffprobePath, "ffprobe", "", "path to ffprobe executable")
	flag.BoolVar(&preview, "preview", false, "generate preview mp4")
	flag.BoolVar(&webp, "webp", false, "generate preview webp")
	flag.BoolVar(&screenshot, "screenshot", false, "generate screenshot jpg")
	flag.BoolVar(&sprite, "sprite", false, "generate sprite jpg and vtt")
	flag.BoolVar(&transcode, "transcode", false, "generate transcode mp4")
	flag.BoolVar(&overwrite, "overwrite", true, "overwrite existing files")
	flag.Parse()

	job := worker.Job{}
	if jobFile != "" {
		b, err := os.ReadFile(jobFile)
		if err != nil {
			return fmt.Errorf("reading job file: %w", err)
		}
		if err := json.Unmarshal(b, &job); err != nil {
			return fmt.Errorf("parsing job file: %w", err)
		}
	} else {
		job = worker.Job{
			InputPath:    inputPath,
			Checksum:     checksum,
			GeneratedDir: generatedDir,
			Preview:      preview,
			WebP:         webp,
			Screenshot:   screenshot,
			Sprite:       sprite,
			Transcode:    transcode,
			Overwrite:    overwrite,
			PreviewOptions: worker.PreviewOptions{
				Segments:        12,
				SegmentDuration: 0.5,
				ExcludeStart:    "0",
				ExcludeEnd:      "0",
				Preset:          "veryfast",
				Audio:           false,
			},
			SpriteOptions: worker.SpriteOptions{
				Count: 25,
				Size:  320,
			},
		}
	}

	if job.InputPath == "" {
		return errors.New("input path is required")
	}
	if job.Checksum == "" {
		return errors.New("checksum is required")
	}
	if job.GeneratedDir == "" {
		job.GeneratedDir = "./generated"
	}

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

	if err := worker.Run(context.Background(), worker.RunRequest{
		Job:        job,
		VideoFile:  videoFile,
		FFMpegPath: resolvedFFMpeg,
	}); err != nil {
		return err
	}

	paths := worker.NewOutputPaths(job.GeneratedDir)
	if job.Screenshot {
		fmt.Println("screenshot:", paths.Screenshot(job.Checksum))
	}
	if job.Preview {
		fmt.Println("preview:", paths.Preview(job.Checksum))
	}
	if job.WebP {
		fmt.Println("webp:", paths.WebP(job.Checksum))
	}
	if job.Sprite {
		fmt.Println("sprite image:", paths.SpriteImage(job.Checksum))
		fmt.Println("sprite vtt:", paths.SpriteVTT(job.Checksum))
	}
	if job.Transcode {
		fmt.Println("transcode:", paths.Transcode(job.Checksum))
	}

	return nil
}
