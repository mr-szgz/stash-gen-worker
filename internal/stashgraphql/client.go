package stashgraphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Scene struct {
	ID    string
	Title string
	File  SceneFile
}

type SceneFile struct {
	Path     string
	Checksum string
}

type Client struct {
	Endpoint string
	APIKey   string
	Client   *http.Client
}

func (c Client) FetchScene(ctx context.Context, sceneID string) (Scene, error) {
	endpoint := strings.TrimSpace(c.Endpoint)
	if endpoint == "" {
		return Scene{}, fmt.Errorf("stash GraphQL endpoint is required")
	}
	if strings.TrimSpace(sceneID) == "" {
		return Scene{}, fmt.Errorf("scene ID is required")
	}

	httpClient := c.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	body := map[string]any{
		"query": `
query FindScene($id: ID!) {
  findScene(id: $id) {
    id
    title
    files {
      path
      fingerprint(type: "MD5")
    }
  }
}`,
		"variables": map[string]any{
			"id": sceneID,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return Scene{}, fmt.Errorf("encoding GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Scene{}, fmt.Errorf("creating GraphQL request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey := strings.TrimSpace(c.APIKey); apiKey != "" {
		req.Header.Set("ApiKey", apiKey)
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return Scene{}, fmt.Errorf("calling GraphQL endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Scene{}, fmt.Errorf("GraphQL endpoint returned status %s", resp.Status)
	}

	var response struct {
		Data struct {
			FindScene *struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				Files []struct {
					Path        string `json:"path"`
					Fingerprint string `json:"fingerprint"`
				} `json:"files"`
			} `json:"findScene"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Scene{}, fmt.Errorf("decoding GraphQL response: %w", err)
	}

	if len(response.Errors) > 0 {
		return Scene{}, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}
	if response.Data.FindScene == nil {
		return Scene{}, fmt.Errorf("scene %s not found", sceneID)
	}

	scene := Scene{
		ID:    response.Data.FindScene.ID,
		Title: response.Data.FindScene.Title,
	}
	for _, file := range response.Data.FindScene.Files {
		if strings.TrimSpace(file.Path) == "" {
			continue
		}
		scene.File = SceneFile{
			Path:     file.Path,
			Checksum: file.Fingerprint,
		}
		if scene.File.Checksum != "" {
			break
		}
	}

	if scene.File.Path == "" {
		return Scene{}, fmt.Errorf("scene %s has no file path", sceneID)
	}
	if scene.File.Checksum == "" {
		return Scene{}, fmt.Errorf("scene %s has no MD5 fingerprint", sceneID)
	}

	return scene, nil
}
