package img

import (
	"fmt"
	"image"
	"image/color"
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
	rect := gifCanvas(g)
	info.Width, info.Height = rect.Dx(), rect.Dy()
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
	for i, frame := range FlattenGIF(g) {
		name := fmt.Sprintf("%s_%03d.png", prefix, i)
		dest := filepath.Join(outDir, name)
		if err := save(dest, frame); err != nil {
			return nil, err
		}
		out = append(out, dest)
	}
	return out, nil
}

// FlattenGIF 按 GIF 的帧处置规则合成完整画面，避免只拿到差分小图。
func FlattenGIF(g *gif.GIF) []*image.NRGBA {
	canvasRect := gifCanvas(g)
	canvas := image.NewNRGBA(canvasRect)
	var backup *image.NRGBA
	var prevDisposal byte = gif.DisposalNone
	prevBounds := canvasRect
	out := make([]*image.NRGBA, 0, len(g.Image))
	for i, frame := range g.Image {
		switch prevDisposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, prevBounds, image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			if backup != nil {
				canvas = cloneNRGBA(backup)
			}
		}
		if disposalAt(g, i) == gif.DisposalPrevious {
			backup = cloneNRGBA(canvas)
		}
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		out = append(out, cloneNRGBA(canvas))
		prevDisposal = disposalAt(g, i)
		prevBounds = frame.Bounds()
	}
	return out
}

func disposalAt(g *gif.GIF, i int) byte {
	if i < len(g.Disposal) {
		return g.Disposal[i]
	}
	return gif.DisposalNone
}

func gifCanvas(g *gif.GIF) image.Rectangle {
	if g.Config.Width > 0 && g.Config.Height > 0 {
		return image.Rect(0, 0, g.Config.Width, g.Config.Height)
	}
	var r image.Rectangle
	for _, im := range g.Image {
		r = r.Union(im.Bounds())
	}
	if r.Empty() {
		return image.Rect(0, 0, 1, 1)
	}
	return r
}

// SaveGIF 把一组图写成循环动画。delayCS 是每帧百分之一秒。透明像素会保留。
func SaveGIF(path string, frames []image.Image, delayCS int) error {
	if delayCS <= 0 {
		delayCS = 10
	}
	delays := make([]int, len(frames))
	for i := range delays {
		delays[i] = delayCS
	}
	return SaveGIFDelays(path, frames, delays)
}

// SaveGIFDelays 按每帧自己的延迟写出 GIF。
func SaveGIFDelays(path string, frames []image.Image, delays []int) error {
	if len(frames) == 0 {
		return fmt.Errorf("gif: no frames")
	}
	g := &gif.GIF{LoopCount: 0}
	for i, frame := range frames {
		g.Image = append(g.Image, palettize(frame))
		delay := 10
		if i < len(delays) && delays[i] > 0 {
			delay = delays[i]
		}
		g.Delay = append(g.Delay, delay)
		g.Disposal = append(g.Disposal, gif.DisposalBackground)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}

// KeyGIF 对动画每一帧做色度抠像，并保留透明通道。
func KeyGIF(src, dest string, opt KeyOptions, choke int, feather float64) error {
	g, err := openGIF(src)
	if err != nil {
		return err
	}
	frames := FlattenGIF(g)
	images := make([]image.Image, len(frames))
	for i, frame := range frames {
		cut := ChromaKey(frame, opt)
		var im image.Image = cut
		if choke > 0 || feather > 0 {
			im = Matte(cut, choke, 0, feather)
		}
		images[i] = im
	}
	return SaveGIFDelays(dest, images, g.Delay)
}

func palettize(src image.Image) *image.Paletted {
	b := src.Bounds()
	pal := make(color.Palette, 0, 256)
	pal = append(pal, color.NRGBA{A: 0})
	pal = append(pal, palette.Plan9[:255]...)
	pm := image.NewPaletted(image.Rect(0, 0, b.Dx(), b.Dy()), pal)
	transparent := false
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := src.At(b.Min.X+x, b.Min.Y+y)
			_, _, _, a := c.RGBA()
			if a < 0x8000 {
				pm.SetColorIndex(x, y, 0)
				transparent = true
				continue
			}
			pm.SetColorIndex(x, y, nearestOpaque(pal, c))
		}
	}
	if !transparent {
		dst := image.NewPaletted(pm.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(dst, dst.Bounds(), src, b.Min)
		return dst
	}
	return pm
}

func nearestOpaque(pal color.Palette, c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	bestI := 1
	best := uint32(1 << 31)
	for i := 1; i < len(pal); i++ {
		cr, cg, cb, _ := pal[i].RGBA()
		dr := int32(r) - int32(cr)
		dg := int32(g) - int32(cg)
		db := int32(b) - int32(cb)
		d := uint32(dr*dr + dg*dg + db*db)
		if d < best {
			best = d
			bestI = i
		}
	}
	return uint8(bestI)
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
