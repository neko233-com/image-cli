package img

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSavePNG(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	writePNG(t, in, 64, 32)
	im, format, err := Load(in)
	if err != nil || format != "png" {
		t.Fatalf("load: format=%q err=%v", format, err)
	}
	out := filepath.Join(dir, "out.webp")
	if err := Save(out, im, EncodeOptions{Format: "webp", Quality: 90}); err != nil {
		t.Fatal(err)
	}
	im2, format2, err := Load(out)
	if err != nil || format2 != "webp" || im2.Bounds().Dx() != 64 {
		t.Fatalf("roundtrip webp: %q err=%v", format2, err)
	}
}

func TestInspectAndResize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pot.png")
	writePNG(t, path, 100, 50)
	info, err := Inspect(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if !info.NPOT || info.Width != 100 {
		t.Fatalf("inspect: %+v", info)
	}
	im, _, _ := Load(path)
	resized := Resize(im, 50, 50, ResizeFit)
	if resized.Bounds().Dx() != 50 {
		t.Fatal("resize fit failed")
	}
}

func TestExtractRGB(t *testing.T) {
	dir := t.TempDir()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	paths, err := ExtractRGB(src, RGBExtractOptions{OutDir: dir, Prefix: "t", Separate: true})
	if err != nil || len(paths) < 3 {
		t.Fatalf("extract rgb: %v %v", paths, err)
	}
}

func TestEachRaster(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "a.png"), 4, 4)
	writePNG(t, filepath.Join(dir, "sub", "b.png"), 4, 4)
	count := 0
	if err := EachRaster(dir, false, func(string) error { count++; return nil }); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("non-recursive count=%d", count)
	}
	count = 0
	if err := EachRaster(dir, true, func(string) error { count++; return nil }); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("recursive count=%d", count)
	}
}
