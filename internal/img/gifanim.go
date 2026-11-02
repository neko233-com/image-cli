package img

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
	"path/filepath"
)

// GIFInfo 是动画 GIF 的轻量信息。
type GIFInfo struct {
	Frames int
	Width  int
	Height int
	Delays []int
	Loop   int
}

// InspectGIF 读取全部帧，不把它压成第一帧。
func InspectGIF(path string) (GIFInfo, error) {
	g, err := openGIF(path)
	if err != nil {
		return GIFInfo{}, err
	}
	info := GIFInfo{Frames: len(g.Image), Delays: append([]int(nil), g.Delay...), Loop: g.LoopCount}
	if len(g.Image) > 0 {
		b := g.Image[0].Bounds()
		info.Width, info.Height = b.Dx(), b.Dy()
	}
	return info, nil
}

// ExtractGIF 把每一帧写成 PNG。
func ExtractGIF(path, outDir, prefix string, save func(path string, im image.Image) error) ([]string, error) {
	g, err := openGIF(path)
	if err != nil {
		return nil, err
	}
	if prefix == "" {
		prefix = "frame"
	}
	var out []string
	for i, frame := range g.Image {
		name := fmt.Sprintf("%s_%03d.png", prefix, i)
		dest := filepath.Join(outDir, name)
		if err := save(dest, frame); err != nil {
			return nil, err
		}
		out = append(out, dest)
	}
	return out, nil
}

// SaveGIF 把一组图写成循环动画。delayCS 是每帧百分之一秒。
func SaveGIF(path string, frames []image.Image, delayCS int) error {
	if len(frames) == 0 {
		return fmt.Errorf("gif: no frames")
	}
	if delayCS <= 0 {
		delayCS = 10
	}
	g := &gif.GIF{LoopCount: 0}
	for _, frame := range frames {
		pm := image.NewPaletted(image.Rect(0, 0, frame.Bounds().Dx(), frame.Bounds().Dy()), palette.Plan9)
		draw.FloydSteinberg.Draw(pm, pm.Bounds(), frame, frame.Bounds().Min)
		g.Image = append(g.Image, pm)
		g.Delay = append(g.Delay, delayCS)
		g.Disposal = append(g.Disposal, gif.DisposalBackground)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}

func openGIF(path string) (*gif.GIF, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}
	if len(g.Image) == 0 {
		return nil, fmt.Errorf("gif: no frames")
	}
	return g, nil
}
