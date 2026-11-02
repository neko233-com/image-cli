package img

import (
	"image"
	"image/color"
	"image/draw"
)

// Stroke 在不透明区域外画一圈描边，适合表情包和贴纸。
func Stroke(src image.Image, radius int, col color.NRGBA) *image.NRGBA {
	base := cloneNRGBA(src)
	if radius <= 0 {
		return base
	}
	mask := filterAlpha(base, radius, false)
	b := base.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if base.NRGBAAt(x, y).A > 0 {
				continue
			}
			if mask.NRGBAAt(x, y).A == 0 {
				continue
			}
			c := col
			c.A = mask.NRGBAAt(x, y).A
			dst.SetNRGBA(x, y, c)
		}
	}
	draw.Draw(dst, b, base, b.Min, draw.Over)
	return dst
}

// DropShadow 在主体下方加一层偏移阴影。
func DropShadow(src image.Image, dx, dy, radius int, opacity float64) *image.NRGBA {
	base := cloneNRGBA(src)
	if opacity <= 0 {
		opacity = 0.45
	}
	if opacity > 1 {
		opacity = 1
	}
	blurred := base
	if radius > 0 {
		blurred = blurAlpha(base, radius)
	}
	b := base.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx()+abs(dx)+radius*2, b.Dy()+abs(dy)+radius*2))
	ox := radius
	oy := radius
	if dx < 0 {
		ox -= dx
	}
	if dy < 0 {
		oy -= dy
	}
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			a := blurred.NRGBAAt(b.Min.X+x, b.Min.Y+y).A
			if a == 0 {
				continue
			}
			dst.SetNRGBA(ox+dx+x, oy+dy+y, color.NRGBA{A: uint8(float64(a) * opacity)})
		}
	}
	draw.Draw(dst, image.Rect(ox, oy, ox+b.Dx(), oy+b.Dy()), base, b.Min, draw.Over)
	return dst
}

// TrimAlpha 裁掉四周完全透明的边。
func TrimAlpha(src image.Image) *image.NRGBA {
	base := cloneNRGBA(src)
	b := base.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if base.NRGBAAt(x, y).A == 0 {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+1 > maxX {
				maxX = x + 1
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}
	if !found {
		return base
	}
	dst := image.NewNRGBA(image.Rect(0, 0, maxX-minX, maxY-minY))
	draw.Draw(dst, dst.Bounds(), base, image.Point{X: minX, Y: minY}, draw.Src)
	return dst
}

// CircleMask 把画面裁成圆形透明贴纸。
func CircleMask(src image.Image) *image.NRGBA {
	base := cloneNRGBA(src)
	b := base.Bounds()
	cx := float64(b.Dx()) / 2
	cy := float64(b.Dy()) / 2
	rad := cx
	if cy < rad {
		rad = cy
	}
	rad *= rad
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy > rad {
				base.SetNRGBA(x, y, color.NRGBA{})
			}
		}
	}
	return base
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
