package game

import "strings"

// Platform 发行目标，用于尺寸 / 体积预算检测。
type Platform string

const (
	PlatformWindows      Platform = "windows"
	PlatformMac          Platform = "mac"
	PlatformIOS          Platform = "ios"
	PlatformAndroid      Platform = "android"
	PlatformWebGL        Platform = "webgl"
	PlatformWebGLMobile  Platform = "webgl-mobile"
	PlatformMinigame     Platform = "minigame"
	PlatformSwitch       Platform = "switch"
)

func AllPlatforms() []Platform {
	return []Platform{
		PlatformWindows, PlatformMac, PlatformIOS, PlatformAndroid,
		PlatformWebGL, PlatformWebGLMobile, PlatformMinigame, PlatformSwitch,
	}
}

func ParsePlatforms(list string) ([]Platform, error) {
	list = strings.TrimSpace(strings.ToLower(list))
	if list == "" || list == "all" {
		return AllPlatforms(), nil
	}
	var out []Platform
	for _, part := range strings.Split(list, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		p := Platform(part)
		if !p.Valid() {
			return nil, fmtUnknownPlatform(part)
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return AllPlatforms(), nil
	}
	return out, nil
}

func (p Platform) Valid() bool {
	for _, v := range AllPlatforms() {
		if v == p {
			return true
		}
	}
	return false
}

func fmtUnknownPlatform(part string) error {
	return &platformError{part: part}
}

type platformError struct{ part string }

func (e *platformError) Error() string {
	return "unknown platform " + e.part + " (use windows|mac|ios|android|webgl|webgl-mobile|minigame|switch|all)"
}

// PlatformBudget 各平台保守预算（超出 recommended 告警，超出 max 判定不合理）。
type PlatformBudget struct {
	ID                   Platform
	Title                string
	MaxEdge              int
	RecommendedEdge      int
	MaxPixels            int64
	RecommendedPixels    int64
	MaxFileBytes         int64
	RecommendedFileBytes int64
	WarnNPOT             bool
	RequireMultipleOf    int
	Hint                 string
}

func BudgetFor(p Platform) PlatformBudget {
	switch p {
	case PlatformWindows:
		return PlatformBudget{
			ID: p, Title: "Windows PC",
			MaxEdge: 8192, RecommendedEdge: 4096,
			MaxPixels: 8192 * 8192, RecommendedPixels: 4096 * 4096,
			MaxFileBytes: 32 << 20, RecommendedFileBytes: 4 << 20,
			WarnNPOT: false, RequireMultipleOf: 1,
			Hint: "DXT/BC 贴图常用 4K；UI 建议 ≤2048",
		}
	case PlatformMac:
		return PlatformBudget{
			ID: p, Title: "macOS",
			MaxEdge: 8192, RecommendedEdge: 4096,
			MaxPixels: 8192 * 8192, RecommendedPixels: 4096 * 4096,
			MaxFileBytes: 32 << 20, RecommendedFileBytes: 4 << 20,
			WarnNPOT: false, RequireMultipleOf: 1,
			Hint: "Metal/BC 与 Windows 类似；Retina UI 注意 @2x 源图边长",
		}
	case PlatformIOS:
		return PlatformBudget{
			ID: p, Title: "iOS",
			MaxEdge: 4096, RecommendedEdge: 2048,
			MaxPixels: 4096 * 4096, RecommendedPixels: 2048 * 2048,
			MaxFileBytes: 8 << 20, RecommendedFileBytes: 1 << 20,
			WarnNPOT: true, RequireMultipleOf: 4,
			Hint: "ASTC 4x4；单张 UI 贴图建议 ≤2048，注意内存峰值",
		}
	case PlatformAndroid:
		return PlatformBudget{
			ID: p, Title: "Android",
			MaxEdge: 4096, RecommendedEdge: 2048,
			MaxPixels: 4096 * 4096, RecommendedPixels: 2048 * 2048,
			MaxFileBytes: 8 << 20, RecommendedFileBytes: 1 << 20,
			WarnNPOT: true, RequireMultipleOf: 4,
			Hint: "ASTC/ETC2；中低端机建议 albedo ≤2048",
		}
	case PlatformWebGL:
		return PlatformBudget{
			ID: p, Title: "WebGL (桌面浏览器)",
			MaxEdge: 4096, RecommendedEdge: 2048,
			MaxPixels: 4096 * 4096, RecommendedPixels: 2048 * 2048,
			MaxFileBytes: 4 << 20, RecommendedFileBytes: 512 << 10,
			WarnNPOT: true, RequireMultipleOf: 1,
			Hint: "下载体积敏感；图集建议 POT 且 ≤2048",
		}
	case PlatformWebGLMobile:
		return PlatformBudget{
			ID: p, Title: "WebGL (移动浏览器 / WebView)",
			MaxEdge: 2048, RecommendedEdge: 1024,
			MaxPixels: 2048 * 2048, RecommendedPixels: 1024 * 1024,
			MaxFileBytes: 2 << 20, RecommendedFileBytes: 256 << 10,
			WarnNPOT: true, RequireMultipleOf: 1,
			Hint: "移动 GPU 内存更小；优先 WebP/压缩纹理",
		}
	case PlatformMinigame:
		return PlatformBudget{
			ID: p, Title: "微信等小游戏 / 轻量宿主",
			MaxEdge: 2048, RecommendedEdge: 1024,
			MaxPixels: 2048 * 2048, RecommendedPixels: 1024 * 1024,
			MaxFileBytes: 512 << 10, RecommendedFileBytes: 128 << 10,
			WarnNPOT: true, RequireMultipleOf: 1,
			Hint: "远程单文件建议 ≤128KB；本地图集边长 ≤1024 更稳",
		}
	case PlatformSwitch:
		return PlatformBudget{
			ID: p, Title: "Nintendo Switch",
			MaxEdge: 4096, RecommendedEdge: 2048,
			MaxPixels: 4096 * 4096, RecommendedPixels: 2048 * 2048,
			MaxFileBytes: 16 << 20, RecommendedFileBytes: 2 << 20,
			WarnNPOT: false, RequireMultipleOf: 1,
			Hint: "BC 纹理；内存带宽有限，场景贴图建议 1K–2K",
		}
	default:
		return PlatformBudget{ID: p, Title: string(p), MaxEdge: 4096, RecommendedEdge: 2048,
			MaxPixels: 4096 * 4096, RecommendedPixels: 2048 * 2048}
	}
}

// PlatformsForProfile 按资源场景给出应重点关注的平台集合。
func PlatformsForProfile(profile Profile) []Platform {
	switch profile {
	case ProfileUI, ProfileIcon:
		return []Platform{PlatformWindows, PlatformMac, PlatformIOS, PlatformAndroid, PlatformWebGL, PlatformWebGLMobile, PlatformMinigame}
	case ProfileSprite, ProfileAtlas, ProfileVFX:
		return []Platform{PlatformIOS, PlatformAndroid, PlatformWebGL, PlatformWebGLMobile, PlatformMinigame, PlatformSwitch}
	case ProfileAlbedo, ProfileNormal, ProfileEmissive, ProfileORM, ProfileLightmap, ProfileCubemap:
		return AllPlatforms()
	default:
		return AllPlatforms()
	}
}
