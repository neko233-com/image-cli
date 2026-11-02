package spec

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Domain 是生产规范的使用场景。
type Domain string

const (
	DomainGame   Domain = "game"
	DomainDesign Domain = "design"
	DomainFilm   Domain = "film"
	DomainVFX    Domain = "vfx"
)

func AllDomains() []Domain {
	return []Domain{DomainGame, DomainDesign, DomainFilm, DomainVFX}
}

func ParseDomain(s string) (Domain, error) {
	d := Domain(strings.ToLower(strings.TrimSpace(s)))
	switch d {
	case DomainGame, DomainDesign, DomainFilm, DomainVFX, "auto", "all":
		return d, nil
	default:
		return "", fmt.Errorf("unknown domain %q (game|design|film|vfx|auto|all)", s)
	}
}

type Severity string

const (
	SeverityOK   Severity = "ok"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
)

type Finding struct {
	Domain   Domain   `json:"domain"`
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

type Report struct {
	Domain   Domain    `json:"domain"`
	Overall  Severity  `json:"overall"`
	Summary  string    `json:"summary"`
	Findings []Finding `json:"findings,omitempty"`
}

// Input 只依赖已解码的轻量元数据，不再次读盘。
type Input struct {
	Path     string
	Format   string
	Width    int
	Height   int
	Bytes    int64
	HasAlpha bool
	Bits     int
	NPOT     bool
	Profile  string
}

// DomainsFor 解析 -domain。auto 按文件名和 profile 选一个场景。
func DomainsFor(flag, path, profile string) ([]Domain, error) {
	flag = strings.ToLower(strings.TrimSpace(flag))
	if flag == "" {
		return nil, nil
	}
	if _, err := ParseDomain(flag); err != nil {
		return nil, err
	}
	if flag == "all" {
		return AllDomains(), nil
	}
	if flag == "auto" {
		return []Domain{detectDomain(path, profile)}, nil
	}
	return []Domain{Domain(flag)}, nil
}

func detectDomain(path, profile string) Domain {
	stem := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	for _, s := range []string{"_plate", "_comp", "_matte", "_exr", "_dpx", "_cinema"} {
		if strings.Contains(stem, s) {
			return DomainFilm
		}
	}
	switch strings.ToLower(profile) {
	case "vfx", "fx", "flow", "noise", "particle":
		return DomainVFX
	case "ui", "sprite", "atlas", "icon":
		return DomainDesign
	default:
		if profile == "" {
			return DomainGame
		}
		return DomainGame
	}
}

func Evaluate(domains []Domain, in Input) Report {
	var findings []Finding
	seen := map[string]struct{}{}
	for _, d := range domains {
		for _, f := range rules(d, in) {
			key := string(f.Domain) + ":" + f.Code
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			findings = append(findings, f)
		}
	}
	label := Domain("none")
	if len(domains) == 1 {
		label = domains[0]
	} else if len(domains) > 1 {
		label = "all"
	}
	rep := Report{Domain: label, Overall: SeverityOK, Findings: findings}
	fails, warns := 0, 0
	for _, f := range findings {
		switch f.Severity {
		case SeverityFail:
			fails++
			rep.Overall = SeverityFail
		case SeverityWarn:
			warns++
			if rep.Overall != SeverityFail {
				rep.Overall = SeverityWarn
			}
		}
	}
	switch {
	case fails > 0:
		rep.Summary = fmt.Sprintf("%d 项不合格，%d 项建议修改", fails, warns)
	case warns > 0:
		rep.Summary = fmt.Sprintf("%d 项建议修改", warns)
	default:
		rep.Summary = "规范检查通过"
	}
	return rep
}

func rules(d Domain, in Input) []Finding {
	edge := in.Width
	if in.Height > edge {
		edge = in.Height
	}
	var out []Finding
	add := func(sev Severity, code, msg string) {
		out = append(out, Finding{Domain: d, Code: code, Severity: sev, Message: msg})
	}
	if in.Width <= 0 || in.Height <= 0 {
		add(SeverityFail, "empty", "宽或高为 0")
		return out
	}
	odd := in.Width%2 != 0 || in.Height%2 != 0
	format := strings.ToLower(in.Format)
	switch d {
	case DomainGame:
		if edge > 16384 {
			add(SeverityFail, "edge", fmt.Sprintf("边长 %d 超过 GPU 常见上限 16384", edge))
		} else if edge > 8192 {
			add(SeverityWarn, "edge", fmt.Sprintf("边长 %d 超过游戏常见 8192", edge))
		}
		if odd {
			add(SeverityWarn, "even", "宽高不是偶数，部分压缩与 mip 会出问题")
		}
		if format == "jpeg" && isAlphaProfile(in.Profile) {
			add(SeverityWarn, "jpeg-alpha", "JPEG 不能保留透明，UI/精灵/特效应使用 PNG")
		}
		if in.Profile == "normal" && in.HasAlpha {
			add(SeverityWarn, "normal-alpha", "法线贴图通常不需要 alpha")
		}
		if in.NPOT && isDataProfile(in.Profile) {
			add(SeverityWarn, "npot", "3D 数据/颜色贴图建议 2 的幂，便于压缩与 mip")
		}
	case DomainDesign:
		switch format {
		case "bmp", "gif":
			add(SeverityWarn, "delivery-format", "设计交付建议 PNG、WebP 或 TIFF，而不是 "+format)
		}
		if edge > 16000 {
			add(SeverityFail, "edge", fmt.Sprintf("边长 %d 过大，设计稿建议不超过 16000", edge))
		} else if edge > 8000 {
			add(SeverityWarn, "edge", fmt.Sprintf("边长 %d 超过常见设计稿 8000", edge))
		}
		if in.Bytes > 40<<20 {
			add(SeverityWarn, "filesize", "文件超过 40MiB，交付和预览都会变慢")
		}
		if in.Width*8 < in.Height || in.Height*8 < in.Width {
			add(SeverityWarn, "aspect", "宽高比极端，确认是否为长图/切片")
		}
	case DomainFilm:
		if odd {
			add(SeverityFail, "even", "影视合成要求偶数宽高")
		}
		switch format {
		case "jpeg":
			add(SeverityWarn, "master-format", "JPEG 是 8-bit 有损格式，母版建议 TIFF 或 16-bit PNG")
		case "gif", "bmp":
			add(SeverityFail, "master-format", format+" 不适合影视母版")
		}
		if in.Bits > 0 && in.Bits < 16 && (format == "png" || format == "tiff") {
			add(SeverityWarn, "bit-depth", fmt.Sprintf("当前约 %d-bit，影视母版建议 16-bit", in.Bits))
		}
		if edge < 1280 {
			add(SeverityWarn, "below-hd", fmt.Sprintf("长边 %d 低于 HD 1280", edge))
		}
	case DomainVFX:
		switch format {
		case "jpeg", "gif":
			add(SeverityFail, "vfx-format", format+" 会损坏特效 alpha 或引入色度抽样")
		}
		if !in.HasAlpha {
			add(SeverityWarn, "vfx-alpha", "特效贴图通常需要 alpha")
		}
		if odd {
			add(SeverityWarn, "even", "宽高不是偶数，序列合成时可能错位")
		}
		if edge > 8192 {
			add(SeverityWarn, "edge", fmt.Sprintf("边长 %d 对实时特效偏大", edge))
		}
	}
	return out
}

func isAlphaProfile(p string) bool {
	switch p {
	case "ui", "sprite", "atlas", "icon", "vfx":
		return true
	default:
		return false
	}
}

func isDataProfile(p string) bool {
	switch p {
	case "albedo", "normal", "orm", "roughness", "metallic", "ao", "height", "emissive", "detail", "mask", "lightmap", "cubemap", "hdr":
		return true
	default:
		return false
	}
}
