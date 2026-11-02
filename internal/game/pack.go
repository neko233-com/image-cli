package game

import (
	"fmt"
	"image"
	"image/color"
)

// PackORM 将 AO/Roughness/Metallic 灰度图打包到 RGB（Unity URP 常用：R=AO G=Rough B=Metal）。
func PackORM(ao, roughness, metallic image.Image) (image.Image, error) {
	b := ao.Bounds()
	if roughness.Bounds() != b || metallic.Bounds() != b {
		return nil, fmt.Errorf("orm inputs must share identical dimensions")
	}
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			ar, _, _, _ := ao.At(x, y).RGBA()
			rr, _, _, _ := roughness.At(x, y).RGBA()
			mr, _, _, _ := metallic.At(x, y).RGBA()
			out.Set(x, y, color.RGBA{
				R: uint8(ar >> 8),
				G: uint8(rr >> 8),
				B: uint8(mr >> 8),
				A: 255,
			})
		}
	}
	return out, nil
}
