package img

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"
)

func TestGIFPackExtract(t *testing.T) {
	dir := t.TempDir()
	a := image.NewRGBA(image.Rect(0, 0, 4, 4))
	b := image.NewRGBA(image.Rect(0, 0, 4, 4))
	a.Set(0, 0, color.RGBA{R: 255, A: 255})
	b.Set(0, 0, color.RGBA{B: 255, A: 255})
	path := filepath.Join(dir, "anim.gif")
	if err := SaveGIF(path, []image.Image{a, b}, 8); err != nil {
		t.Fatal(err)
	}
	info, err := InspectGIF(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Frames != 2 || info.Delays[0] != 8 {
		t.Fatalf("%+v", info)
	}
	paths, err := ExtractGIF(path, dir, "f", func(p string, im image.Image) error {
		return Save(p, im, EncodeOptions{Format: "png"})
	})
	if err != nil || len(paths) != 2 {
		t.Fatalf("%v %v", paths, err)
	}
}
