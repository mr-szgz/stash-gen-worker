package worker

import (
	"context"
	"fmt"
	"image"
	"math"
	"os"

	"github.com/disintegration/imaging"
	stashffmpeg "github.com/stashapp/stash/pkg/ffmpeg"
	"github.com/stashapp/stash/pkg/fsutil"
	scenegenerate "github.com/stashapp/stash/pkg/scene/generate"
)

type RunRequest struct {
	Job        Job
	VideoFile  *stashffmpeg.VideoFile
	FFMpegPath string
}

func Run(ctx context.Context, req RunRequest) error {
	paths := newScenePaths(req.Job.GeneratedDir)
	if err := paths.ensure(); err != nil {
		return err
	}

	encoder := stashffmpeg.NewEncoder(req.FFMpegPath)
	gen := scenegenerate.Generator{
		Encoder:      encoder,
		FFMpegConfig: &ffmpegConfig{},
		LockManager:  fsutil.NewReadLockManager(),
		ScenePaths:   paths,
		Overwrite:    req.Job.Overwrite,
	}

	input := req.Job.InputPath
	duration := req.VideoFile.VideoStreamDuration
	if duration <= 0 {
		duration = req.VideoFile.FileDuration
	}

	if req.Job.Preview {
		if err := gen.PreviewVideo(ctx, input, duration, req.Job.Checksum, req.Job.PreviewOptions.ToStash(), req.Job.PreviewOptions.Fallback, req.Job.PreviewOptions.UseVsync2); err != nil {
			return fmt.Errorf("generate preview: %w", err)
		}
	}

	if req.Job.WebP {
		if err := gen.PreviewWebp(ctx, input, req.Job.Checksum); err != nil {
			return fmt.Errorf("generate webp: %w", err)
		}
	}

	if req.Job.Screenshot {
		imgBytes, err := gen.Screenshot(ctx, input, req.VideoFile.Width, duration, scenegenerate.ScreenshotOptions{})
		if err != nil {
			return fmt.Errorf("generate screenshot: %w", err)
		}
		if err := os.WriteFile(paths.GetScreenshotPath(req.Job.Checksum), imgBytes, 0o644); err != nil {
			return fmt.Errorf("write screenshot: %w", err)
		}
	}

	if req.Job.Sprite {
		if err := generateSprite(ctx, gen, paths, input, req.Job.Checksum, req.VideoFile, req.Job.SpriteOptions); err != nil {
			return fmt.Errorf("generate sprite: %w", err)
		}
	}

	if req.Job.Transcode {
		width, height := req.VideoFile.TranscodeScale(720)
		if err := gen.Transcode(ctx, input, req.Job.Checksum, scenegenerate.TranscodeOptions{Width: width, Height: height}); err != nil {
			return fmt.Errorf("generate transcode: %w", err)
		}
	}

	return nil
}

func generateSprite(ctx context.Context, gen scenegenerate.Generator, paths *scenePaths, input string, checksum string, vf *stashffmpeg.VideoFile, options SpriteOptions) error {
	count := options.Count
	if count <= 0 {
		count = 25
	}
	size := options.Size
	if size <= 0 {
		size = 320
	}
	duration := vf.VideoStreamDuration
	if duration <= 0 {
		duration = vf.FileDuration
	}
	if duration <= 0 {
		return fmt.Errorf("invalid video duration")
	}

	isPortrait := vf.Height > vf.Width
	stepSize := duration / float64(count)
	images := make([]image.Image, 0, count)
	for i := 0; i < count; i++ {
		seconds := float64(i) * stepSize
		img, err := gen.SpriteScreenshot(ctx, input, seconds, size, isPortrait)
		if err != nil {
			return err
		}
		images = append(images, img)
	}
	if len(images) == 0 {
		return fmt.Errorf("no sprite images generated")
	}

	montage := gen.CombineSpriteImages(images)
	if err := imaging.Save(montage, paths.GetSpriteImageFilePath(checksum), imaging.JPEGQuality(90)); err != nil {
		return err
	}

	chunks := int(math.Min(float64(count), float64(len(images))))
	return gen.SpriteVTT(ctx, paths.GetSpriteVttFilePath(checksum), paths.GetSpriteImageFilePath(checksum), stepSize, chunks)
}
