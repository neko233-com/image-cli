package img

import (
	"image"
	"image/color"
	"image/gif"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func solidRow() *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	im.SetNRGBA(0, 0, color.NRGBA{G: 255, A: 255})
	im.SetNRGBA(1, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	im.SetNRGBA(2, 0, color.NRGBA{R: 220, G: 40, B: 40, A: 255})
	return im
}

func TestChromaKeyKeepsSubject(t *testing.T) {
	out := ChromaKey(solidRow(), KeyOptions{Despill: true})
	if out.NRGBAAt(0, 0).A != 0 {
		t.Fatalf("green alpha %d", out.NRGBAAt(0, 0).A)
	}
	if out.NRGBAAt(1, 0).A < 250 {
		t.Fatalf("white alpha %d", out.NRGBAAt(1, 0).A)
	}
	if out.NRGBAAt(2, 0).A < 250 {
		t.Fatalf("red alpha %d", out.NRGBAAt(2, 0).A)
	}
}

func TestDespillLimitsGreen(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	im.SetNRGBA(0, 0, color.NRGBA{R: 80, G: 255, B: 90, A: 255})
	out := Despill(im, color.NRGBA{G: 255, A: 255})
	c := out.NRGBAAt(0, 0)
	if c.G > c.R && c.G > c.B {
		t.Fatalf("still spilling %+v", c)
	}
}

func TestStickerStroke(t *testing.T) {
	keyed := ChromaKey(solidRow(), KeyOptions{})
	out := Stroke(keyed, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	left := out.NRGBAAt(0, 0)
	if left.A == 0 || left.R < 250 {
		t.Fatalf("expected white stroke, got %+v", left)
	}
}

func TestLetterboxWider(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	out := Letterbox(im, 2, color.Black)
	if out.Bounds().Dx() != 200 || out.Bounds().Dy() != 100 {
		t.Fatalf("%v", out.Bounds())
	}
}

func TestMemeCaptionASCII(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 80, 40))
	out := MemeCaption(im, "HI", "", basicfont.Face7x13)
	ink := 0
	for y := 0; y < out.Bounds().Dy(); y++ {
		for x := 0; x < out.Bounds().Dx(); x++ {
			if out.NRGBAAt(x, y).A > 0 {
				ink++
			}
		}
	}
	if ink == 0 {
		t.Fatal("expected caption pixels")
	}
}

func TestTransparentGIFRoundTrip(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	im.SetNRGBA(0, 0, color.NRGBA{A: 0})
	im.SetNRGBA(1, 0, color.NRGBA{R: 255, A: 255})
	path := t.TempDir() + "/a.gif"
	if err := SaveGIF(path, []image.Image{im}, 5); err != nil {
		t.Fatal(err)
	}
	f, err := openGIF(path)
	if err != nil {
		t.Fatal(err)
	}
	flat := FlattenGIF(f)
	if flat[0].NRGBAAt(0, 0).A > 20 {
		t.Fatalf("transparent pixel alpha %d", flat[0].NRGBAAt(0, 0).A)
	}
	if flat[0].NRGBAAt(1, 0).R < 200 {
		t.Fatalf("red pixel %+v", flat[0].NRGBAAt(1, 0))
	}
	_ = gif.DisposalBackground
}
