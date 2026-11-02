package game

import (
	"path/filepath"
	"strings"
)

// DetectProfile 按常见引擎贴图命名识别场景（Unity / Unreal 后缀）。
// 未识别时 ok 为 false，由调用方决定回退 profile。
func DetectProfile(path string) (Profile, bool) {
	stem := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	stem = strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(stem)
	for _, rule := range detectSuffixes {
		if strings.HasSuffix(stem, rule.suffix) {
			return rule.profile, true
		}
	}
	return "", false
}

var detectSuffixes = []struct {
	suffix  string
	profile Profile
}{
	{"_occlusionroughnessmetallic", ProfileORM},
	{"_basecolor", ProfileAlbedo},
	{"_basecolour", ProfileAlbedo},
	{"_normalmap", ProfileNormal},
	{"_detailnormal", ProfileNormal},
	{"_ambientocclusion", ProfileAO},
	{"_displacement", ProfileHeight},
	{"_smoothness", ProfileRoughness},
	{"_metalness", ProfileMetallic},
	{"_roughness", ProfileRoughness},
	{"_metallic", ProfileMetallic},
	{"_emissive", ProfileEmissive},
	{"_emission", ProfileEmissive},
	{"_occlusion", ProfileAO},
	{"_lightmap", ProfileLightmap},
	{"_cubemap", ProfileCubemap},
	{"_specular", ProfileSpecular},
	{"_opacity", ProfileOpacity},
	{"_flowmap", ProfileFlow},
	{"_albedo", ProfileAlbedo},
	{"_diffuse", ProfileAlbedo},
	{"_height", ProfileHeight},
	{"_normal", ProfileNormal},
	{"_detail", ProfileDetail},
	{"_sprite", ProfileSprite},
	{"_atlas", ProfileAtlas},
	{"_particle", ProfileVFX},
	{"_noise", ProfileNoise},
	{"_mask", ProfileMask},
	{"_flow", ProfileFlow},
	{"_icon", ProfileIcon},
	{"_orm", ProfileORM},
	{"_rma", ProfileORM},
	{"_arm", ProfileORM},
	{"_vfx", ProfileVFX},
	{"_hdr", ProfileHDR},
	{"_nrm", ProfileNormal},
	{"_ao", ProfileAO},
	{"_fx", ProfileVFX},
	{"_nm", ProfileNormal},
	{"_n", ProfileNormal},
	{"_d", ProfileAlbedo},
	{"_e", ProfileEmissive},
	{"_r", ProfileRoughness},
	{"_m", ProfileMetallic},
	{"_h", ProfileHeight},
	{"_s", ProfileSpecular},
}
