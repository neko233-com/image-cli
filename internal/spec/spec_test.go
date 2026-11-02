package spec

import "testing"

func TestFilmRejectsOddAndSmall(t *testing.T) {
	rep := Evaluate([]Domain{DomainFilm}, Input{Width: 8, Height: 8, Format: "png", Bits: 8})
	if rep.Overall != SeverityWarn {
		t.Fatalf("%+v", rep)
	}
	rep = Evaluate([]Domain{DomainFilm}, Input{Width: 1921, Height: 1080, Format: "png", Bits: 16})
	if rep.Overall != SeverityFail {
		t.Fatalf("odd film: %+v", rep)
	}
}

func TestVFXRejectsJPEG(t *testing.T) {
	rep := Evaluate([]Domain{DomainVFX}, Input{Width: 64, Height: 64, Format: "jpeg", HasAlpha: false})
	if rep.Overall != SeverityFail {
		t.Fatalf("%+v", rep)
	}
}

func TestAutoDomain(t *testing.T) {
	d, err := DomainsFor("auto", "shot_plate.png", "")
	if err != nil || d[0] != DomainFilm {
		t.Fatal(d, err)
	}
	d, err = DomainsFor("auto", "fx.png", "vfx")
	if err != nil || d[0] != DomainVFX {
		t.Fatal(d, err)
	}
}

func TestDesignFlagsBMP(t *testing.T) {
	rep := Evaluate([]Domain{DomainDesign}, Input{Width: 100, Height: 100, Format: "bmp"})
	if rep.Overall != SeverityWarn {
		t.Fatalf("%+v", rep)
	}
}
