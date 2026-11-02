package game

import (
	"image"
	"testing"
)

func TestGenerateMipmapsCount(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 256, 128))
	levels, err := GenerateMipmaps(src, 0, "box")
	if err != nil {
		t.Fatal(err)
	}
	if len(levels) < 8 {
		t.Fatalf("expected long chain, got %d levels", len(levels))
	}
	if levels[0].Width != 256 || levels[0].Height != 128 {
		t.Fatalf("level0 size mismatch")
	}
	last := levels[len(levels)-1]
	if last.Width != 1 || last.Height != 1 {
		t.Fatalf("expected 1x1 last level, got %dx%d", last.Width, last.Height)
	}
}

func TestDefaultsNoMipmap(t *testing.T) {
	for _, p := range []Profile{ProfileUI, ProfileAlbedo, ProfileNormal, ProfileSprite} {
		d := Defaults(p, TargetMobile)
		if d.Mipmap {
			t.Fatalf("profile %s should default mipmap=false", p)
		}
	}
}
