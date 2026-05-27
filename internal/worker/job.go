package worker

import scenegenerate "github.com/stashapp/stash/pkg/scene/generate"

type Job struct {
	InputPath      string         `json:"input_path"`
	Checksum       string         `json:"checksum"`
	GeneratedDir   string         `json:"generated_dir"`
	Preview        bool           `json:"preview"`
	WebP           bool           `json:"webp"`
	Screenshot     bool           `json:"screenshot"`
	Sprite         bool           `json:"sprite"`
	Transcode      bool           `json:"transcode"`
	Overwrite      bool           `json:"overwrite"`
	PreviewOptions PreviewOptions `json:"preview_options"`
	SpriteOptions  SpriteOptions  `json:"sprite_options"`
}

type PreviewOptions struct {
	Segments        int     `json:"segments"`
	SegmentDuration float64 `json:"segment_duration"`
	ExcludeStart    string  `json:"exclude_start"`
	ExcludeEnd      string  `json:"exclude_end"`
	Preset          string  `json:"preset"`
	Audio           bool    `json:"audio"`
	Fallback        bool    `json:"fallback"`
	UseVsync2       bool    `json:"use_vsync_2"`
}

type SpriteOptions struct {
	Count int `json:"count"`
	Size  int `json:"size"`
}

func (p PreviewOptions) ToStash() scenegenerate.PreviewOptions {
	segments := p.Segments
	if segments <= 0 {
		segments = 12
	}
	segmentDuration := p.SegmentDuration
	if segmentDuration <= 0 {
		segmentDuration = 0.5
	}
	preset := p.Preset
	if preset == "" {
		preset = "veryfast"
	}
	return scenegenerate.PreviewOptions{
		Segments:        segments,
		SegmentDuration: segmentDuration,
		ExcludeStart:    p.ExcludeStart,
		ExcludeEnd:      p.ExcludeEnd,
		Preset:          preset,
		Audio:           p.Audio,
	}
}
