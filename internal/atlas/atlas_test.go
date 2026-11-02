package atlas

import (
	"image"
	"image/color"
	"testing"
)

func solid(w, h int, c color.NRGBA) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.SetNRGBA(x, y, c)
		}
	}
	return im
}

func TestPackRoundTripEngines(t *testing.T) {
	sheet, err := Pack([]Sprite{
		{Name: "red", Image: solid(8, 8, color.NRGBA{R: 255, A: 255})},
		{Name: "blue", Image: solid(8, 4, color.NRGBA{B: 255, A: 255})},
	}, PackOptions{Padding: 0, POT: true, MaxSize: 256})
	if err != nil {
		t.Fatal(err)
	}
	if sheet.Width != 8 || sheet.Height != 16 {
		t.Fatalf("sheet %dx%d", sheet.Width, sheet.Height)
	}

	meta, err := WriteUnityMeta("atlas.png", sheet)
	if err != nil {
		t.Fatal(err)
	}
	unity, err := ParseUnityMeta(meta, sheet.Height)
	if err != nil {
		t.Fatal(err)
	}
	assertSamePixels(t, sheet, unity)

	plist, err := WriteCocosPlist("atlas.png", sheet)
	if err != nil {
		t.Fatal(err)
	}
	_, cocos, err := ParseCocosPlist([]byte(plist))
	if err != nil {
		t.Fatal(err)
	}
	assertSamePixels(t, sheet, cocos)

	raw, err := WriteUE("atlas.png", sheet)
	if err != nil {
		t.Fatal(err)
	}
	_, ue, err := ParseUE(raw)
	if err != nil {
		t.Fatal(err)
	}
	assertSamePixels(t, sheet, ue)
}

func TestCocosFormat2AndRotated(t *testing.T) {
	const plist = `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
<key>frames</key><dict>
<key>spin.png</key><dict>
<key>frame</key><string>{{1,2},{4,8}}</string>
<key>rotated</key><true/>
<key>sourceSize</key><string>{8,4}</string>
</dict>
</dict>
<key>metadata</key><dict>
<key>format</key><integer>2</integer>
<key>textureFileName</key><string>a.png</string>
</dict>
</dict></plist>`
	tex, sprites, err := ParseCocosPlist([]byte(plist))
	if err != nil {
		t.Fatal(err)
	}
	if tex != "a.png" || len(sprites) != 1 || !sprites[0].Rotated {
		t.Fatalf("%s %+v", tex, sprites)
	}
	if sprites[0].W != 8 || sprites[0].H != 4 || sprites[0].X != 1 {
		t.Fatalf("%+v", sprites[0])
	}
}

func assertSamePixels(t *testing.T, sheet Sheet, parsed []Sprite) {
	t.Helper()
	if len(parsed) != len(sheet.Sprites) {
		t.Fatalf("count %d want %d", len(parsed), len(sheet.Sprites))
	}
	byName := map[string]Sprite{}
	for _, sp := range parsed {
		byName[sp.Name] = sp
	}
	for _, sp := range sheet.Sprites {
		got, ok := byName[sp.Name]
		if !ok {
			t.Fatalf("missing %s", sp.Name)
		}
		crop, err := Crop(sheet.Image, got)
		if err != nil {
			t.Fatal(err)
		}
		want := sp.Image.At(sp.Image.Bounds().Min.X, sp.Image.Bounds().Min.Y)
		gotC := crop.At(0, 0)
		wr, wg, wb, wa := want.RGBA()
		gr, gg, gb, ga := gotC.RGBA()
		if wr != gr || wg != gg || wb != gb || wa != ga {
			t.Fatalf("%s color mismatch", sp.Name)
		}
	}
}
