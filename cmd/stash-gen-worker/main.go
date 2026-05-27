package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	stashffmpeg "github.com/stashapp/stash/pkg/ffmpeg"
	scenegenerate "github.com/stashapp/stash/pkg/scene/generate"

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

	if job.Screenshot {
		out := filepath.Join(job.GeneratedDir, "screenshots", job.Checksum+".jpg")
		fmt.Println("screenshot:", out)
	}
	if job.Preview {
		out := filepath.Join(job.GeneratedDir, "screenshots", job.Checksum+".mp4")
		fmt.Println("preview:", out)
	}
	if job.WebP {
		out := filepath.Join(job.GeneratedDir, "screenshots", job.Checksum+".webp")
		fmt.Println("webp:", out)
	}
	if job.Sprite {
		fmt.Println("sprite image:", filepath.Join(job.GeneratedDir, "vtt", job.Checksum+"_sprite.jpg"))
		fmt.Println("sprite vtt:", filepath.Join(job.GeneratedDir, "vtt", job.Checksum+"_thumbs.vtt"))
	}
	if job.Transcode {
		fmt.Println("transcode:", filepath.Join(job.GeneratedDir, "transcodes", job.Checksum+".mp4"))
	}

	_ = scenegenerate.PreviewOptions{}
	_ = imaging.JPEGQuality(90)
	_ = image.Point{}
	return nil
}
