package game

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

// MipmapLevel 单级 mip 及其逻辑尺寸。
type MipmapLevel struct {
	Level int
	Width int
	Height int
	Image  image.Image
}

// GenerateMipmaps 生成完整或受限的 mip 链（level0 为原图）。
// filter: box（游戏常用）| lanczos（UI 缩放更清晰时可单独 resize，不建议用于 3D mip）。
func GenerateMipmaps(src image.Image, maxLevels int, filter string) ([]MipmapLevel, error) {
	if src == nil {
		return nil, fmt.Errorf("nil image")
	}
	b := src.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		return nil, fmt.Errorf("empty image")
	}
	levels := []MipmapLevel{{Level: 0, Width: b.Dx(), Height: b.Dy(), Image: src}}
	current := src
	for level := 1; ; level++ {
		if maxLevels > 0 && level >= maxLevels {
			break
		}
		cb := current.Bounds()
		w, h := cb.Dx(), cb.Dy()
		if w <= 1 && h <= 1 {
			break
		}
		nw, nh := w/2, h/2
		if nw < 1 {
			nw = 1
		}
		if nh < 1 {
			nh = 1
		}
		if nw == w && nh == h {
			break
		}
		var next image.Image
		switch filter {
		case "lanczos":
			next = imaging.Resize(current, nw, nh, imaging.Lanczos)
		default:
			next = imaging.Resize(current, nw, nh, imaging.Box)
		}
		levels = append(levels, MipmapLevel{Level: level, Width: nw, Height: nh, Image: next})
		current = next
	}
	return levels, nil
}
