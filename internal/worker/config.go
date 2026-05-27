package worker

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	JobsDir               string `json:"jobs_dir"`
	GeneratedDir          string `json:"generated_dir"`
	FFMpegPath            string `json:"ffmpeg_path"`
	FFProbePath           string `json:"ffprobe_path"`
	StashGraphQLEndpoint  string `json:"stash_graphql_endpoint"`
	StashAPIKey           string `json:"stash_api_key"`
}

func DefaultConfig() Config {
	return Config{
		JobsDir:      "./jobs",
		GeneratedDir: "./generated",
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.JobsDir == "" {
		cfg.JobsDir = DefaultConfig().JobsDir
	}
	if cfg.GeneratedDir == "" {
		cfg.GeneratedDir = DefaultConfig().GeneratedDir
	}

	return cfg, nil
}
