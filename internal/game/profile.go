package game

import (
	"fmt"
	"strings"
)

// Profile 描述游戏资源使用场景；默认均不生成 mipmap（UI / 2D 优先）。
type Profile string

const (
	ProfileAuto     Profile = "auto"
	ProfileUI       Profile = "ui"
	ProfileSprite   Profile = "sprite"
	ProfileAtlas    Profile = "atlas"
	ProfileIcon     Profile = "icon"
	ProfileVFX      Profile = "vfx"
	ProfileAlbedo   Profile = "albedo"
	ProfileNormal   Profile = "normal"
	ProfileEmissive Profile = "emissive"
	ProfileORM      Profile = "orm"
	ProfileRoughness Profile = "roughness"
	ProfileMetallic  Profile = "metallic"
	ProfileAO        Profile = "ao"
	ProfileHeight    Profile = "height"
	ProfileMask      Profile = "mask"
	ProfileDetail    Profile = "detail"
	ProfileSpecular  Profile = "specular"
	ProfileOpacity   Profile = "opacity"
	ProfileFlow      Profile = "flow"
	ProfileNoise     Profile = "noise"
	ProfileHDR       Profile = "hdr"
	ProfileLightmap Profile = "lightmap"
	ProfileCubemap  Profile = "cubemap"
)

// AllProfiles 返回可执行的场景（不含 auto）。
func AllProfiles() []Profile {
	return []Profile{
		ProfileUI, ProfileSprite, ProfileAtlas, ProfileIcon, ProfileVFX,
		ProfileAlbedo, ProfileNormal, ProfileEmissive, ProfileORM,
		ProfileRoughness, ProfileMetallic, ProfileAO, ProfileHeight,
		ProfileMask, ProfileDetail, ProfileSpecular, ProfileOpacity,
		ProfileFlow, ProfileNoise, ProfileHDR,
		ProfileLightmap, ProfileCubemap,
	}
}

func ParseProfile(s string) (Profile, error) {
	raw := strings.ToLower(strings.TrimSpace(s))
	raw = strings.ReplaceAll(raw, "-", "")
	raw = strings.ReplaceAll(raw, "_", "")
	raw = strings.ReplaceAll(raw, " ", "")
	if p, ok := profileAliases[raw]; ok {
		return p, nil
	}
	return "", fmt.Errorf("unknown profile %q (see: image-cli game profiles)", s)
}

var profileAliases = map[string]Profile{
	"auto": ProfileAuto,
	"ui": ProfileUI, "hud": ProfileUI, "gui": ProfileUI,
	"sprite": ProfileSprite, "sprites": ProfileSprite,
	"atlas": ProfileAtlas, "sheet": ProfileAtlas, "spritesheet": ProfileAtlas,
	"icon": ProfileIcon,
	"vfx": ProfileVFX, "fx": ProfileVFX, "effect": ProfileVFX, "particle": ProfileVFX, "particles": ProfileVFX,
	"albedo": ProfileAlbedo, "basecolor": ProfileAlbedo, "basecolour": ProfileAlbedo,
	"diffuse": ProfileAlbedo, "diff": ProfileAlbedo, "color": ProfileAlbedo, "colour": ProfileAlbedo,
	"base": ProfileAlbedo, "alb": ProfileAlbedo,
	"normal": ProfileNormal, "nrm": ProfileNormal, "nrml": ProfileNormal, "normalmap": ProfileNormal, "nor": ProfileNormal,
	"emissive": ProfileEmissive, "emission": ProfileEmissive, "emit": ProfileEmissive,
	"orm": ProfileORM, "rma": ProfileORM, "arm": ProfileORM, "occlusionroughnessmetallic": ProfileORM,
	"roughness": ProfileRoughness, "rough": ProfileRoughness, "smoothness": ProfileRoughness,
	"metallic": ProfileMetallic, "metal": ProfileMetallic, "metalness": ProfileMetallic,
	"ao": ProfileAO, "occlusion": ProfileAO, "ambientocclusion": ProfileAO, "ambient": ProfileAO,
	"height": ProfileHeight, "disp": ProfileHeight, "displacement": ProfileHeight, "bump": ProfileHeight,
	"mask": ProfileMask,
	"detail": ProfileDetail, "detailmask": ProfileDetail, "detailnormal": ProfileNormal,
	"specular": ProfileSpecular, "spec": ProfileSpecular,
	"opacity": ProfileOpacity, "alpha": ProfileOpacity,
	"flow": ProfileFlow, "flowmap": ProfileFlow,
	"noise": ProfileNoise,
	"hdr": ProfileHDR, "exr": ProfileHDR,
	"lightmap": ProfileLightmap, "light": ProfileLightmap, "gi": ProfileLightmap,
	"cubemap": ProfileCubemap, "cube": ProfileCubemap, "env": ProfileCubemap, "sky": ProfileCubemap,
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
	case ProfileSprite, ProfileAtlas:
		d.MaxEdge = 4096
		d.Premultiply = true
	case ProfileVFX:
		d.MaxEdge = 2048
		d.Premultiply = true
	case ProfileIcon:
		d.MaxEdge = 1024
	case ProfileAlbedo, ProfileDetail:
		return colorMap(target, 4096, true)
	case ProfileNormal:
		return dataMap(target, 4096, "bc5", "norm")
	case ProfileEmissive:
		return colorMap(target, 2048, false)
	case ProfileORM, ProfileMask:
		return dataMap(target, 4096, "bc7", "linear")
	case ProfileRoughness, ProfileMetallic, ProfileAO, ProfileHeight, ProfileSpecular, ProfileOpacity:
		return dataMap(target, 4096, "bc4", "linear")
	case ProfileFlow, ProfileNoise:
		d.MaxEdge = 1024
		d.Pow2 = true
		d.SRGB = false
		d.Compress = true
		if target == TargetPC {
			d.CompressFormat = "bc7"
		} else {
			d.CompressFormat = "astc"
			d.ASTCMode = "linear"
		}
	case ProfileHDR, ProfileLightmap:
		return hdrMap(target, 2048, false)
	case ProfileCubemap:
		out := hdrMap(target, 2048, true)
		return out
	}
	return d
}

func colorMap(target Target, edge int, srgb bool) ProfileDefaults {
	d := ProfileDefaults{Pow2: true, MaxEdge: edge, Compress: true, SRGB: srgb, ASTCMode: "srgb", CompressFormat: "astc"}
	if !srgb {
		d.ASTCMode = "hdr"
	}
	if target == TargetPC {
		d.CompressFormat = "bc7"
	}
	return d
}

func dataMap(target Target, edge int, pcFormat, astcMode string) ProfileDefaults {
	d := ProfileDefaults{
		Pow2: true, MaxEdge: edge, Compress: true, SRGB: false,
		CompressFormat: "astc", ASTCMode: astcMode,
	}
	if target == TargetPC {
		d.CompressFormat = pcFormat
	}
	return d
}

func hdrMap(target Target, edge int, square bool) ProfileDefaults {
	d := ProfileDefaults{
		Pow2: true, Square: square, MaxEdge: edge, Compress: true, SRGB: false,
		CompressFormat: "astc", ASTCMode: "hdr",
	}
	if target == TargetPC {
		d.CompressFormat = "bc6h"
	}
	return d
}
