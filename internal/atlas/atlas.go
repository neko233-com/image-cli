package atlas

import (
	"fmt"
	"image"
	"image/draw"
	"sort"
)

// Sprite 是图集中的一块。X/Y 是 PNG 左上角为原点的像素坐标。
type Sprite struct {
	Name     string
	Image    image.Image
	X        int
	Y        int
	W        int
	H        int
	Rotated  bool
	SourceW  int
	SourceH  int
	TrimX    int
	TrimY    int
}

// Sheet 是一张合图及其切片。
type Sheet struct {
	Name    string
	Width   int
	Height  int
	Image   *image.NRGBA
	Sprites []Sprite
}

type PackOptions struct {
	Padding int
	MaxSize int
	POT     bool
}

func (o PackOptions) withDefaults() PackOptions {
	if o.MaxSize <= 0 {
		o.MaxSize = 4096
	}
	if o.Padding < 0 {
		o.Padding = 0
	}
	return o
}

// Pack 用货架算法把多张小图合成一张图集。不旋转，便于 Unity / UE / Cocos 直接使用。
func Pack(sprites []Sprite, opt PackOptions) (Sheet, error) {
	opt = opt.withDefaults()
	if len(sprites) == 0 {
		return Sheet{}, fmt.Errorf("no sprites")
	}
	items := make([]Sprite, len(sprites))
	copy(items, sprites)
	for i := range items {
		if items[i].Image == nil {
			return Sheet{}, fmt.Errorf("%s: missing image", items[i].Name)
		}
		b := items[i].Image.Bounds()
		items[i].W = b.Dx()
		items[i].H = b.Dy()
		if items[i].SourceW == 0 {
			items[i].SourceW = items[i].W
		}
		if items[i].SourceH == 0 {
			items[i].SourceH = items[i].H
		}
		if items[i].W <= 0 || items[i].H <= 0 {
			return Sheet{}, fmt.Errorf("%s: empty sprite", items[i].Name)
		}
		if items[i].W > opt.MaxSize || items[i].H > opt.MaxSize {
			return Sheet{}, fmt.Errorf("%s: %dx%d exceeds max size %d", items[i].Name, items[i].W, items[i].H, opt.MaxSize)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].H != items[j].H {
			return items[i].H > items[j].H
		}
		if items[i].W != items[j].W {
			return items[i].W > items[j].W
		}
		return items[i].Name < items[j].Name
	})

	width := nextPOT(items[0].W)
	if width < 1 {
		width = 1
	}
	var placed []Sprite
	var usedW, usedH int
	for {
		var ok bool
		placed, usedW, usedH, ok = shelf(items, width, opt.Padding, opt.MaxSize)
		if ok {
			break
		}
		if width >= opt.MaxSize {
			return Sheet{}, fmt.Errorf("sprites do not fit within %d", opt.MaxSize)
		}
		next := width * 2
		if next > opt.MaxSize || next < width {
			next = opt.MaxSize
		}
		if next == width {
			return Sheet{}, fmt.Errorf("sprites do not fit within %d", opt.MaxSize)
		}
		width = next
	}
	if opt.POT {
		usedW = nextPOT(usedW)
		usedH = nextPOT(usedH)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, usedW, usedH))
	for _, sp := range placed {
		draw.Draw(dst, image.Rect(sp.X, sp.Y, sp.X+sp.W, sp.Y+sp.H), sp.Image, sp.Image.Bounds().Min, draw.Src)
	}
	return Sheet{Width: usedW, Height: usedH, Image: dst, Sprites: placed}, nil
}

func shelf(items []Sprite, width, padding, maxSize int) ([]Sprite, int, int, bool) {
	placed := make([]Sprite, len(items))
	x, y := 0, 0
	shelfH := 0
	usedW := 0
	for i, sp := range items {
		if sp.W > width || sp.H > maxSize {
			return nil, 0, 0, false
		}
		if x > 0 && x+sp.W > width {
			y += shelfH + padding
			x = 0
			shelfH = 0
		}
		if y+sp.H > maxSize {
			return nil, 0, 0, false
		}
		sp.X = x
		sp.Y = y
		placed[i] = sp
		if sp.H > shelfH {
			shelfH = sp.H
		}
		right := x + sp.W
		if right > usedW {
			usedW = right
		}
		x += sp.W + padding
	}
	return placed, usedW, y + shelfH, true
}

func nextPOT(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

// Crop 从图集中取出一块。Rotated 为 true 时按 Cocos 约定：图集里是顺时针 90°，这里再逆时针转回。
func Crop(src image.Image, sp Sprite) (image.Image, error) {
	tw, th := sp.W, sp.H
	if sp.Rotated {
		tw, th = sp.H, sp.W
	}
	if tw <= 0 || th <= 0 {
		return nil, fmt.Errorf("%s: invalid size", sp.Name)
	}
	b := src.Bounds()
	r := image.Rect(b.Min.X+sp.X, b.Min.Y+sp.Y, b.Min.X+sp.X+tw, b.Min.Y+sp.Y+th)
	if !r.In(b) {
		return nil, fmt.Errorf("%s: rect %v outside %v", sp.Name, r, b)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, tw, th))
	draw.Draw(dst, dst.Bounds(), src, r.Min, draw.Src)
	if sp.Rotated {
		return rotateCCW(dst), nil
	}
	return dst, nil
}

func rotateCCW(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(y, w-1-x, src.NRGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}
