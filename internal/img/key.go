package img

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"
)

// KeyOptions 控制色度抠像。Similarity 越大，越接近键色的像素越透明。
type KeyOptions struct {
	Key         color.NRGBA
	Similarity  float64
	Smoothness  float64
	Despill     bool
}

func (o KeyOptions) withDefaults() KeyOptions {
	if o.Key == (color.NRGBA{}) {
		o.Key = color.NRGBA{G: 255, A: 255}
	}
	if o.Similarity <= 0 {
		o.Similarity = 0.28
	}
	if o.Smoothness <= 0 {
		o.Smoothness = 0.15
	}
	return o
}

// ParseKeyColor 接受 green、blue、screen（棚拍绿）或 #RRGGBB。
func ParseKeyColor(s string) (color.NRGBA, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "green":
		return color.NRGBA{G: 255, A: 255}, nil
	case "screen":
		return color.NRGBA{G: 0xB1, B: 0x40, A: 255}, nil
	case "blue":
		return color.NRGBA{B: 255, A: 255}, nil
	case "magenta":
		return color.NRGBA{R: 255, B: 255, A: 255}, nil
	case "black":
		return color.NRGBA{A: 255}, nil
	case "white":
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}, nil
	case "none":
		return color.NRGBA{}, fmt.Errorf("color none")
	}
	hex := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "#"), "0x")
	if len(hex) != 6 {
		return color.NRGBA{}, fmt.Errorf("unknown key color %q", s)
	}
	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("unknown key color %q", s)
	}
	return color.NRGBA{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n), A: 255}, nil
}

// ChromaKey 按键色生成透明通道，并可选去掉溢色。
func ChromaKey(src image.Image, opt KeyOptions) *image.NRGBA {
	opt = opt.withDefaults()
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	kr := float64(opt.Key.R) / 255
	kg := float64(opt.Key.G) / 255
	kb := float64(opt.Key.B) / 255
	dom := dominant(kr, kg, kb)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r0, g0, b0, a0 := src.At(x, y).RGBA()
			if a0 == 0 {
				continue
			}
			r := float64(r0) / 65535
			g := float64(g0) / 65535
			bl := float64(b0) / 65535
			a := float64(a0) / 65535
			a *= screenAlpha(r, g, bl, kr, kg, kb, dom, opt.Similarity, opt.Smoothness)
			if opt.Despill {
				r, g, bl = despill(r, g, bl, dom)
			}
			dst.SetNRGBA(x-b.Min.X, y-b.Min.Y, color.NRGBA{
				R: to8(r), G: to8(g), B: to8(bl), A: to8(a),
			})
		}
	}
	return dst
}

// Despill 只压掉键色通道，不改透明度。
func Despill(src image.Image, key color.NRGBA) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	dom := dominant(float64(key.R), float64(key.G), float64(key.B))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r0, g0, b0, a0 := src.At(x, y).RGBA()
			r, g, bl := despill(float64(r0)/65535, float64(g0)/65535, float64(b0)/65535, dom)
			dst.SetNRGBA(x-b.Min.X, y-b.Min.Y, color.NRGBA{R: to8(r), G: to8(g), B: to8(bl), A: uint8(a0 >> 8)})
		}
	}
	return dst
}

func dominant(r, g, b float64) int {
	if g >= r && g >= b {
		return 1
	}
	if b >= r && b >= g {
		return 2
	}
	return 0
}

func screenAlpha(r, g, b, kr, kg, kb float64, dom int, similarity, smoothness float64) float64 {
	screen, other := screenPair(r, g, b, dom)
	if screen <= other {
		return 1
	}
	dist := math.Sqrt((r-kr)*(r-kr)+(g-kg)*(g-kg)+(b-kb)*(b-kb)) / math.Sqrt(3)
	if dist <= similarity {
		return 0
	}
	if dist >= similarity+smoothness {
		return 1
	}
	return (dist - similarity) / smoothness
}

func screenPair(r, g, b float64, dom int) (float64, float64) {
	switch dom {
	case 1:
		return g, math.Max(r, b)
	case 2:
		return b, math.Max(r, g)
	default:
		return r, math.Max(g, b)
	}
}

func despill(r, g, b float64, dom int) (float64, float64, float64) {
	switch dom {
	case 1:
		if limit := math.Max(r, b); g > limit {
			g = limit
		}
	case 2:
		if limit := math.Max(r, g); b > limit {
			b = limit
		}
	default:
		if limit := math.Max(g, b); r > limit {
			r = limit
		}
	}
	return r, g, b
}

func to8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

// Matte 收缩、扩张或羽化已有透明通道。
func Matte(src image.Image, choke, expand int, feather float64) *image.NRGBA {
	dst := cloneNRGBA(src)
	if choke > 0 {
		dst = filterAlpha(dst, choke, true)
	}
	if expand > 0 {
		dst = filterAlpha(dst, expand, false)
	}
	if feather > 0 {
		dst = blurAlpha(dst, int(feather+0.5))
	}
	return dst
}

func filterAlpha(src *image.NRGBA, radius int, erode bool) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			a := c.A
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					xx, yy := x+dx, y+dy
					var na uint8
					if xx < b.Min.X || yy < b.Min.Y || xx >= b.Max.X || yy >= b.Max.Y {
						na = 0
					} else {
						na = src.NRGBAAt(xx, yy).A
					}
					if erode {
						if na < a {
							a = na
						}
					} else if na > a {
						a = na
					}
				}
			}
			c.A = a
			if a == 0 {
				c = color.NRGBA{}
			}
			dst.SetNRGBA(x, y, c)
		}
	}
	return dst
}

func blurAlpha(src *image.NRGBA, radius int) *image.NRGBA {
	if radius < 1 {
		return src
	}
	b := src.Bounds()
	tmp := image.NewNRGBA(b)
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sum, n := 0, 0
			for dx := -radius; dx <= radius; dx++ {
				xx := x + dx
				if xx < b.Min.X || xx >= b.Max.X {
					continue
				}
				sum += int(src.NRGBAAt(xx, y).A)
				n++
			}
			c := src.NRGBAAt(x, y)
			c.A = uint8(sum / n)
			tmp.SetNRGBA(x, y, c)
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sum, n := 0, 0
			for dy := -radius; dy <= radius; dy++ {
				yy := y + dy
				if yy < b.Min.Y || yy >= b.Max.Y {
					continue
				}
				sum += int(tmp.NRGBAAt(x, yy).A)
				n++
			}
			c := tmp.NRGBAAt(x, y)
			c.A = uint8(sum / n)
			if c.A == 0 {
				c = color.NRGBA{}
			}
			dst.SetNRGBA(x, y, c)
		}
	}
	return dst
}

// Composite 把前景按透明通道叠到背景上。
func Composite(bg, fg image.Image, x, y int) *image.NRGBA {
	dst := cloneNRGBA(bg)
	fb := fg.Bounds()
	draw.Draw(dst, image.Rect(x, y, x+fb.Dx(), y+fb.Dy()), fg, fb.Min, draw.Over)
	return dst
}

// Letterbox 按目标宽高比加遮幅。aspect 是宽/高，例如 2.39。
func Letterbox(src image.Image, aspect float64, bar color.Color) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if aspect <= 0 {
		aspect = 2.39
	}
	cur := float64(w) / float64(h)
	outW, outH := w, h
	if cur > aspect {
		outH = int(float64(w)/aspect + 0.5)
	} else if cur < aspect {
		outW = int(float64(h)*aspect + 0.5)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, outW, outH))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: bar}, image.Point{}, draw.Src)
	ox := (outW - w) / 2
	oy := (outH - h) / 2
	draw.Draw(dst, image.Rect(ox, oy, ox+w, oy+h), src, b.Min, draw.Src)
	return dst
}

// Vignette 从画面边缘压暗，strength 取 0 到 1。
func Vignette(src image.Image, strength float64) *image.NRGBA {
	if strength < 0 {
		strength = 0
	}
	if strength > 1 {
		strength = 1
	}
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	cx := float64(b.Dx()) / 2
	cy := float64(b.Dy()) / 2
	maxD := math.Hypot(cx, cy)
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r0, g0, b0, a0 := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			d := math.Hypot(float64(x)-cx, float64(y)-cy) / maxD
			shade := 1 - strength*d*d
			dst.SetNRGBA(x, y, color.NRGBA{
				R: to8(float64(r0) / 65535 * shade),
				G: to8(float64(g0) / 65535 * shade),
				B: to8(float64(b0) / 65535 * shade),
				A: uint8(a0 >> 8),
			})
		}
	}
	return dst
}

func cloneNRGBA(src image.Image) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}
