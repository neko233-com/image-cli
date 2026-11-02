package game

import (
	"encoding/json"
	"os"
	"time"
)

type Manifest struct {
	GeneratedAt string            `json:"generated_at"`
	Input       string            `json:"input"`
	Profile     Profile           `json:"profile"`
	Target      Target            `json:"target"`
	Mipmap      bool              `json:"mipmap"`
	MipLevels   int               `json:"mip_levels,omitempty"`
	Outputs     []ManifestOutput  `json:"outputs"`
	Warnings    []string          `json:"warnings,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type ManifestOutput struct {
	Kind   string `json:"kind"` // prepared_png|mip_png|astc|dds
	Path   string `json:"path"`
	Level  int    `json:"level,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

func WriteManifest(path string, m Manifest) error {
	m.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}
