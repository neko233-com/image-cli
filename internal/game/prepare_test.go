package game

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareUINoCompress(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "btn.png")
	f, err := os.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, 16, 16))); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()
	no := false
	res, err := Prepare(in, PrepareOptions{
		Profile: ProfileUI, Target: TargetMobile,
		Compress: &no, SkipCompress: true, WriteManifest: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Manifest.Meta["budget_overall"] == "" {
		t.Fatal("missing budget meta")
	}
	found := false
	for _, o := range res.Manifest.Outputs {
		if o.Kind == "prepared_png" {
			found = true
		}
	}
	if !found {
		t.Fatal("no prepared png")
	}
}
