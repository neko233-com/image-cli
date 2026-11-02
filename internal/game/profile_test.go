package game

import "testing"

func TestParseProfileTarget(t *testing.T) {
	p, err := ParseProfile("UI")
	if err != nil || p != ProfileUI {
		t.Fatal(err)
	}
	if _, err := ParseProfile("nope"); err == nil {
		t.Fatal("expected error")
	}
	tg, err := ParseTarget("mobile")
	if err != nil || tg != TargetMobile {
		t.Fatal(err)
	}
}

func TestDefaultsUICompressOff(t *testing.T) {
	d := Defaults(ProfileUI, TargetMobile)
	if d.Compress || d.Mipmap {
		t.Fatalf("ui defaults: %+v", d)
	}
}

func TestDefaultsAlbedoMobileASTC(t *testing.T) {
	d := Defaults(ProfileAlbedo, TargetMobile)
	if !d.Compress || d.CompressFormat != "astc" || d.Mipmap {
		t.Fatalf("albedo mobile: %+v", d)
	}
}

func TestParsePlatforms(t *testing.T) {
	all, err := ParsePlatforms("all")
	if err != nil || len(all) != len(AllPlatforms()) {
		t.Fatal(err)
	}
	sub, err := ParsePlatforms("ios,android")
	if err != nil || len(sub) != 2 {
		t.Fatal(err)
	}
}
