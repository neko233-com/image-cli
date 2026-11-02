package game

import "testing"

func TestParseProfileAliases(t *testing.T) {
	p, err := ParseProfile("base-color")
	if err != nil || p != ProfileAlbedo {
		t.Fatalf("base-color: %v %s", err, p)
	}
	p, err = ParseProfile("particle")
	if err != nil || p != ProfileVFX {
		t.Fatalf("particle: %v %s", err, p)
	}
	p, err = ParseProfile("auto")
	if err != nil || p != ProfileAuto {
		t.Fatalf("auto: %v %s", err, p)
	}
}

func TestDetectProfileSuffixes(t *testing.T) {
	cases := map[string]Profile{
		"hero_albedo.png":     ProfileAlbedo,
		"char_body_n.png":     ProfileNormal,
		"T_Weapon_ORM.tga":    ProfileORM,
		"fx_fire_vfx.png":     ProfileVFX,
		"ground_roughness.png": ProfileRoughness,
		"env_cubemap.png":     ProfileCubemap,
		"prop_detailnormal.png": ProfileNormal,
	}
	for name, want := range cases {
		got, ok := DetectProfile(name)
		if !ok || got != want {
			t.Fatalf("%s: got %s ok=%v want %s", name, got, ok, want)
		}
	}
	if _, ok := DetectProfile("readme.png"); ok {
		t.Fatal("unexpected match")
	}
}

func TestDefaultsDataAndVFX(t *testing.T) {
	v := Defaults(ProfileVFX, TargetMobile)
	if !v.Premultiply || v.Compress || v.Mipmap {
		t.Fatalf("vfx: %+v", v)
	}
	r := Defaults(ProfileRoughness, TargetPC)
	if r.CompressFormat != "bc4" || r.SRGB {
		t.Fatalf("roughness pc: %+v", r)
	}
	a := Defaults(ProfileAlbedo, TargetMobile)
	if a.ASTCMode != "srgb" || a.CompressFormat != "astc" {
		t.Fatalf("albedo: %+v", a)
	}
}
