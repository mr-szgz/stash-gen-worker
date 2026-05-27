package worker

import (
	"fmt"
	"os"
	"path/filepath"
)

type scenePaths struct {
	root        string
	screenshots string
	vtt         string
	markers     string
	transcodes  string
	tmp         string
}

func newScenePaths(root string) *scenePaths {
	return &scenePaths{
		root:        root,
		screenshots: filepath.Join(root, "screenshots"),
		vtt:         filepath.Join(root, "vtt"),
		markers:     filepath.Join(root, "markers"),
		transcodes:  filepath.Join(root, "transcodes"),
		tmp:         filepath.Join(root, "tmp"),
	}
}

func (p *scenePaths) ensure() error {
	for _, dir := range []string{p.root, p.screenshots, p.vtt, p.markers, p.transcodes, p.tmp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}
	return nil
}

func (p *scenePaths) TempFile(pattern string) (*os.File, error) {
	if err := p.ensure(); err != nil {
		return nil, err
	}
	return os.CreateTemp(p.tmp, pattern)
}

func (p *scenePaths) GetVideoPreviewPath(checksum string) string {
	return filepath.Join(p.screenshots, checksum+".mp4")
}

func (p *scenePaths) GetWebpPreviewPath(checksum string) string {
	return filepath.Join(p.screenshots, checksum+".webp")
}

func (p *scenePaths) GetScreenshotPath(checksum string) string {
	return filepath.Join(p.screenshots, checksum+".jpg")
}

func (p *scenePaths) GetSpriteImageFilePath(checksum string) string {
	return filepath.Join(p.vtt, checksum+"_sprite.jpg")
}

func (p *scenePaths) GetSpriteVttFilePath(checksum string) string {
	return filepath.Join(p.vtt, checksum+"_thumbs.vtt")
}

func (p *scenePaths) GetTranscodePath(checksum string) string {
	return filepath.Join(p.transcodes, checksum+".mp4")
}
