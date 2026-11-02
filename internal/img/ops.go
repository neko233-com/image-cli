package img

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
)

type ResizeMode string

const (
	ResizeFit    ResizeMode = "fit"
	ResizeFill   ResizeMode = "fill"
	ResizeExact  ResizeMode = "exact"
	ResizeWidth  ResizeMode = "width"
	ResizeHeight ResizeMode = "height"
)

func Resize(src image.Image, width, height int, mode ResizeMode) image.Image {
	switch mode {
	case ResizeWidth:
		if width <= 0 {
			return src
		}
		return imaging.Resize(src, width, 0, imaging.Lanczos)
	case ResizeHeight:
		if height <= 0 {
			return src
		}
		return imaging.Resize(src, 0, height, imaging.Lanczos)
	case ResizeExact:
		if width <= 0 || height <= 0 {
			return src
		}
		return imaging.Resize(src, width, height, imaging.Lanczos)
	case ResizeFill:
		if width <= 0 || height <= 0 {
			return src
		}
		return imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)
	default:
		if width <= 0 && height <= 0 {
			return src
		}
		if width <= 0 {
			return imaging.Resize(src, 0, height, imaging.Lanczos)
		}
		if height <= 0 {
			return imaging.Resize(src, width, 0, imaging.Lanczos)
		}
		return imaging.Fit(src, width, height, imaging.Lanczos)
	}
}

func Crop(src image.Image, x, y, w, h int) image.Image {
	b := src.Bounds()
	rect := image.Rect(x, y, x+w, y+h).Intersect(b)
	if rect.Empty() {
		return src
	}
	return imaging.Crop(src, rect)
}

func Rotate(src image.Image, angle float64) image.Image {
	return imaging.Rotate(src, angle, color.Transparent)
}

func Flip(src image.Image, horizontal, vertical bool) image.Image {
	out := src
	if horizontal {
		out = imaging.FlipH(out)
	}
	if vertical {
		out = imaging.FlipV(out)
	}
	return out
}

func Pad(src image.Image, top, right, bottom, left int, bg color.Color) image.Image {
	b := src.Bounds()
	w := b.Dx() + left + right
	h := b.Dy() + top + bottom
	dst := imaging.New(w, h, bg)
	draw.Draw(dst, image.Rect(left, top, left+b.Dx(), top+b.Dy()), src, b.Min, draw.Over)
	return dst
}

func Grayscale(src image.Image) image.Image {
	return imaging.Grayscale(src)
}

func Blur(src image.Image, sigma float64) image.Image {
	if sigma <= 0 {
		sigma = 1
	}
	return imaging.Blur(src, sigma)
}

func Sharpen(src image.Image, sigma float64) image.Image {
	if sigma <= 0 {
		sigma = 1
	}
	return imaging.Sharpen(src, sigma)
}

func AdjustGamma(src image.Image, gamma float64) image.Image {
	if gamma <= 0 {
		gamma = 1
	}
	return imaging.AdjustGamma(src, gamma)
}

func AdjustContrast(src image.Image, percent float64) image.Image {
	return imaging.AdjustContrast(src, percent)
}

func AdjustBrightness(src image.Image, percent float64) image.Image {
	return imaging.AdjustBrightness(src, percent)
}

func AdjustSaturation(src image.Image, percent float64) image.Image {
	return imaging.AdjustSaturation(src, percent)
}

func NextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

func PowerOfTwoCanvas(src image.Image, square bool) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	nw, nh := NextPowerOfTwo(w), NextPowerOfTwo(h)
	if square {
		if nw < nh {
			nw = nh
		} else {
			nh = nw
		}
	}
	if nw == w && nh == h {
		return src
	}
	return Pad(src, 0, nw-w, nh-h, 0, color.Transparent)
}

type RGBExtractOptions struct {
	OutDir   string
	Prefix   string
	RawRGB   bool
	Separate bool
}

// ExtractRGB 导出 RGB/RGBA 平面或合并 raw 数据。
func ExtractRGB(src image.Image, opts RGBExtractOptions) ([]string, error) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, err
	}
	prefix := opts.Prefix
	if prefix == "" {
		prefix = "channel"
	}
	var written []string
	if opts.RawRGB {
		path := filepath.Join(opts.OutDir, prefix+".rgb")
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, w*h*3)
		i := 0
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				r, g, bl, _ := src.At(x, y).RGBA()
				buf[i] = byte(r >> 8)
				buf[i+1] = byte(g >> 8)
				buf[i+2] = byte(bl >> 8)
				i += 3
			}
		}
		if _, err := f.Write(buf); err != nil {
			_ = f.Close()
			return nil, err
		}
		_ = f.Close()
		written = append(written, path)
	}
	if !opts.Separate && !opts.RawRGB {
		opts.Separate = true
	}
	if opts.Separate {
		for _, ch := range []struct {
			name string
			idx  int
		}{
			{"r", 0}, {"g", 1}, {"b", 2}, {"a", 3},
		} {
			if ch.name == "a" && !imageHasAlpha(src) {
				continue
			}
			dst := imaging.New(w, h, color.Black)
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					r, g, bl, a := src.At(x, y).RGBA()
					var v uint32
					switch ch.idx {
					case 0:
						v = r
					case 1:
						v = g
					case 2:
						v = bl
					case 3:
						v = a
					}
					c := color.Gray{Y: uint8(v >> 8)}
					if ch.name == "a" {
						dst.Set(x, y, color.Alpha{A: uint8(v >> 8)})
					} else {
						dst.Set(x, y, c)
					}
				}
			}
			path := filepath.Join(opts.OutDir, fmt.Sprintf("%s_%s.png", prefix, ch.name))
			if err := Save(path, dst, EncodeOptions{Format: "png"}); err != nil {
				return written, err
			}
			written = append(written, path)
		}
	}
	return written, nil
}

type PaletteEntry struct {
	Hex   string `json:"hex"`
	Count int    `json:"count"`
}

func ExtractPalette(src image.Image, maxColors int) ([]PaletteEntry, error) {
	if maxColors <= 0 {
		maxColors = 16
	}
	if maxColors > 256 {
		maxColors = 256
	}
	counts := map[color.RGBA]int{}
	b := src.Bounds()
	step := 1
	pixels := b.Dx() * b.Dy()
	if pixels > 512*512 {
		step = int(math.Sqrt(float64(pixels) / (512 * 512)))
		if step < 1 {
			step = 1
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			r, g, bl, a := src.At(x, y).RGBA()
			key := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(a >> 8)}
			counts[key]++
		}
	}
	type pair struct {
		c color.RGBA
		n int
	}
	list := make([]pair, 0, len(counts))
	for c, n := range counts {
		list = append(list, pair{c: c, n: n})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })
	if len(list) > maxColors {
		list = list[:maxColors]
	}
	out := make([]PaletteEntry, 0, len(list))
	for _, p := range list {
		out = append(out, PaletteEntry{
			Hex:   fmt.Sprintf("#%02x%02x%02x", p.c.R, p.c.G, p.c.B),
			Count: p.n,
		})
	}
	return out, nil
}

func Watermark(base, mark image.Image, x, y int, opacity float64) image.Image {
	if opacity <= 0 {
		opacity = 0.5
	}
	if opacity > 1 {
		opacity = 1
	}
	dst := imaging.Clone(base)
	mb := mark.Bounds()
	layer := applyOpacity(mark, opacity)
	draw.Draw(dst, image.Rect(x, y, x+mb.Dx(), y+mb.Dy()), layer, mb.Min, draw.Over)
	return dst
}

func SplitGrid(src image.Image, cols, rows int) []image.Image {
	if cols <= 0 || rows <= 0 {
		return nil
	}
	b := src.Bounds()
	cellW := b.Dx() / cols
	cellH := b.Dy() / rows
	if cellW <= 0 || cellH <= 0 {
		return nil
	}
	out := make([]image.Image, 0, cols*rows)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := b.Min.X + col*cellW
			y := b.Min.Y + row*cellH
			out = append(out, imaging.Crop(src, image.Rect(x, y, x+cellW, y+cellH)))
		}
	}
	return out
}

func applyOpacity(src image.Image, opacity float64) image.Image {
	b := src.Bounds()
	dst := imaging.New(b.Dx(), b.Dy(), color.Transparent)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			na := uint32(float64(a) * opacity)
			dst.Set(x, y, color.NRGBA{
				R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(na >> 8),
			})
		}
	}
	return dst
}

func OutputPath(input, output, ext string) string {
	if output != "" {
		return output
	}
	ext = strings.TrimPrefix(ext, ".")
	if ext == "" {
		ext = "png"
	}
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	return filepath.Join(filepath.Dir(input), base+"."+ext)
}
