package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const UESchema = "image-cli.ue-atlas.v1"

type ueFile struct {
	Schema  string     `json:"schema"`
	Texture string     `json:"texture"`
	Origin  string     `json:"origin"`
	Sprites []ueSprite `json:"sprites"`
}

type ueSprite struct {
	Name         string `json:"name"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	W            int    `json:"w"`
	H            int    `json:"h"`
	Rotated      bool   `json:"rotated"`
	SourceWidth  int    `json:"sourceWidth"`
	SourceHeight int    `json:"sourceHeight"`
	TrimX        int    `json:"trimX"`
	TrimY        int    `json:"trimY"`
}

// WriteUE 写出 Unreal Paper2D 可用的左上角像素坐标 JSON。
func WriteUE(textureName string, sheet Sheet) ([]byte, error) {
	file := ueFile{Schema: UESchema, Texture: textureName, Origin: "top-left"}
	for _, sp := range sheet.Sprites {
		file.Sprites = append(file.Sprites, ueSprite{
			Name: sp.Name, X: sp.X, Y: sp.Y, W: sp.W, H: sp.H, Rotated: sp.Rotated,
			SourceWidth: sp.SourceW, SourceHeight: sp.SourceH, TrimX: sp.TrimX, TrimY: sp.TrimY,
		})
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(file); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ParseUE(data []byte) (string, []Sprite, error) {
	var file ueFile
	if err := json.Unmarshal(data, &file); err != nil {
		return "", nil, err
	}
	if len(file.Sprites) == 0 {
		return "", nil, fmt.Errorf("ue atlas: no sprites")
	}
	if file.Origin != "" && file.Origin != "top-left" {
		return "", nil, fmt.Errorf("ue atlas: unsupported origin %q", file.Origin)
	}
	out := make([]Sprite, 0, len(file.Sprites))
	for _, sp := range file.Sprites {
		s := Sprite{
			Name: sp.Name, X: sp.X, Y: sp.Y, W: sp.W, H: sp.H, Rotated: sp.Rotated,
			SourceW: sp.SourceWidth, SourceH: sp.SourceHeight, TrimX: sp.TrimX, TrimY: sp.TrimY,
		}
		if s.SourceW == 0 {
			s.SourceW = s.W
		}
		if s.SourceH == 0 {
			s.SourceH = s.H
		}
		out = append(out, s)
	}
	return file.Texture, out, nil
}
