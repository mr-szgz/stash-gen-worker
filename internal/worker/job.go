package worker

import scenegenerate "github.com/stashapp/stash/pkg/scene/generate"

type Job struct {
	SceneID        string         `json:"scene_id,omitempty"`
	SceneTitle     string         `json:"scene_title,omitempty"`
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

func DefaultJob() Job {
	return Job{
		GeneratedDir: "./generated",
		Overwrite:    true,
		PreviewOptions: PreviewOptions{
			Segments:        12,
			SegmentDuration: 0.5,
			ExcludeStart:    "0",
			ExcludeEnd:      "0",
			Preset:          "veryfast",
			Audio:           false,
		},
		SpriteOptions: SpriteOptions{
			Count: 25,
			Size:  320,
		},
	}
}

func (j *Job) ApplyDefaults(defaultGeneratedDir string) {
	defaults := DefaultJob()

	if j.GeneratedDir == "" {
		if defaultGeneratedDir != "" {
			j.GeneratedDir = defaultGeneratedDir
		} else {
			j.GeneratedDir = defaults.GeneratedDir
		}
	}

	if j.PreviewOptions.Segments <= 0 {
		j.PreviewOptions.Segments = defaults.PreviewOptions.Segments
	}
	if j.PreviewOptions.SegmentDuration <= 0 {
		j.PreviewOptions.SegmentDuration = defaults.PreviewOptions.SegmentDuration
	}
	if j.PreviewOptions.ExcludeStart == "" {
		j.PreviewOptions.ExcludeStart = defaults.PreviewOptions.ExcludeStart
	}
	if j.PreviewOptions.ExcludeEnd == "" {
		j.PreviewOptions.ExcludeEnd = defaults.PreviewOptions.ExcludeEnd
	}
	if j.PreviewOptions.Preset == "" {
		j.PreviewOptions.Preset = defaults.PreviewOptions.Preset
	}

	if j.SpriteOptions.Count <= 0 {
		j.SpriteOptions.Count = defaults.SpriteOptions.Count
	}
	if j.SpriteOptions.Size <= 0 {
		j.SpriteOptions.Size = defaults.SpriteOptions.Size
	}
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
