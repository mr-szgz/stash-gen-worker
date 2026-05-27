package worker

import "path/filepath"

type OutputPaths struct {
	root string
}

func NewOutputPaths(root string) OutputPaths {
	return OutputPaths{root: root}
}

func (p OutputPaths) Screenshot(checksum string) string {
	return filepath.Join(p.root, "screenshots", checksum+".jpg")
}

func (p OutputPaths) Preview(checksum string) string {
	return filepath.Join(p.root, "screenshots", checksum+".mp4")
}

func (p OutputPaths) WebP(checksum string) string {
	return filepath.Join(p.root, "screenshots", checksum+".webp")
}

func (p OutputPaths) SpriteImage(checksum string) string {
	return filepath.Join(p.root, "vtt", checksum+"_sprite.jpg")
}

func (p OutputPaths) SpriteVTT(checksum string) string {
	return filepath.Join(p.root, "vtt", checksum+"_thumbs.vtt")
}

func (p OutputPaths) Transcode(checksum string) string {
	return filepath.Join(p.root, "transcodes", checksum+".mp4")
}
