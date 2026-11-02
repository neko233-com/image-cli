package game

import (
	"fmt"
	"strings"
)

// Profile 描述游戏资源使用场景；默认均不生成 mipmap（UI / 2D 优先）。
type Profile string

const (
	ProfileUI       Profile = "ui"
	ProfileSprite   Profile = "sprite"
	ProfileAtlas    Profile = "atlas"
	ProfileIcon     Profile = "icon"
	ProfileVFX      Profile = "vfx"
	ProfileAlbedo   Profile = "albedo"
	ProfileNormal   Profile = "normal"
	ProfileEmissive Profile = "emissive"
	ProfileORM      Profile = "orm"
	ProfileLightmap Profile = "lightmap"
	ProfileCubemap  Profile = "cubemap"
)

func ParseProfile(s string) (Profile, error) {
	p := Profile(strings.ToLower(strings.TrimSpace(s)))
	switch p {
	case ProfileUI, ProfileSprite, ProfileAtlas, ProfileIcon, ProfileVFX,
		ProfileAlbedo, ProfileNormal, ProfileEmissive, ProfileORM, ProfileLightmap, ProfileCubemap:
		return p, nil
	default:
		return "", fmt.Errorf("unknown profile %q (ui|sprite|atlas|icon|vfx|albedo|normal|emissive|orm|lightmap|cubemap)", s)
	}
}

// Target 平台目标：移动端默认 ASTC，PC 走 BC 系列。
type Target string

const (
	TargetMobile Target = "mobile"
	TargetPC     Target = "pc"
)

func ParseTarget(s string) (Target, error) {
	t := Target(strings.ToLower(strings.TrimSpace(s)))
	switch t {
	case TargetMobile, TargetPC:
		return t, nil
	default:
		return "", fmt.Errorf("unknown platform %q (mobile|pc)", s)
	}
}

// ProfileDefaults 返回场景默认策略（不含用户显式 flag 覆盖）。
type ProfileDefaults struct {
	Mipmap         bool
	Pow2           bool
	Square         bool
	MaxEdge        int
	Compress       bool
	CompressFormat string // astc|bc7|bc5|bc6h|none
	ASTCMode       string // L|H|norm
	SRGB           bool
	Premultiply    bool
}

func Defaults(profile Profile, target Target) ProfileDefaults {
	d := ProfileDefaults{
		Mipmap:         false,
		Pow2:           false,
		Square:         false,
		MaxEdge:        0,
		Compress:       false,
		CompressFormat: "none",
		ASTCMode:       "L",
		SRGB:           true,
		Premultiply:    false,
	}
	switch profile {
	case ProfileUI:
		d.MaxEdge = 2048
		d.Premultiply = true
	case ProfileSprite, ProfileAtlas, ProfileVFX:
		d.MaxEdge = 4096
	case ProfileIcon:
		d.MaxEdge = 1024
	case ProfileAlbedo:
		d.Pow2 = true
		d.MaxEdge = 4096
		d.Compress = true
		d.SRGB = true
		if target == TargetPC {
			d.CompressFormat = "bc7"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "L"
		}
	case ProfileNormal:
		d.Pow2 = true
		d.MaxEdge = 4096
		d.Compress = true
		d.SRGB = false
		if target == TargetPC {
			d.CompressFormat = "bc5"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "norm"
		}
	case ProfileEmissive:
		d.Pow2 = true
		d.MaxEdge = 2048
		d.Compress = true
		if target == TargetPC {
			d.CompressFormat = "bc7"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "H"
		}
	case ProfileORM:
		d.Pow2 = true
		d.MaxEdge = 4096
		d.Compress = true
		d.SRGB = false
		if target == TargetPC {
			d.CompressFormat = "bc7"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "L"
		}
	case ProfileLightmap:
		d.Pow2 = true
		d.MaxEdge = 2048
		d.Compress = true
		d.SRGB = false
		if target == TargetPC {
			d.CompressFormat = "bc6h"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "H"
		}
	case ProfileCubemap:
		d.Pow2 = true
		d.Square = true
		d.MaxEdge = 2048
		d.Compress = true
		if target == TargetPC {
			d.CompressFormat = "bc6h"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "H"
		}
	}
	return d
}
