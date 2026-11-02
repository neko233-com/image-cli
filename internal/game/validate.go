package game

import (
	"fmt"
	"sort"
)

type Severity string

const (
	SeverityOK   Severity = "ok"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
)

type BudgetInput struct {
	Width     int
	Height    int
	FileBytes int64
	NPOT      bool
	Usage     Profile // optional, empty = generic
}

type PlatformVerdict struct {
	Platform Platform       `json:"platform"`
	Title    string         `json:"title"`
	Status   Severity       `json:"status"`
	Messages []string       `json:"messages,omitempty"`
	Hint     string         `json:"hint,omitempty"`
	Budget   PlatformBudget `json:"-"`
}

type BudgetReport struct {
	Width      int               `json:"width"`
	Height     int               `json:"height"`
	FileBytes  int64             `json:"file_bytes"`
	Pixels     int64             `json:"pixels"`
	MaxEdge    int               `json:"max_edge"`
	Overall    Severity          `json:"overall"`
	Platforms  []PlatformVerdict `json:"platforms"`
	Summary    string            `json:"summary"`
}

func EvaluateBudget(in BudgetInput, platforms []Platform) BudgetReport {
	maxEdge := in.Width
	if in.Height > maxEdge {
		maxEdge = in.Height
	}
	pixels := int64(in.Width) * int64(in.Height)
	report := BudgetReport{
		Width: in.Width, Height: in.Height, FileBytes: in.FileBytes,
		Pixels: pixels, MaxEdge: maxEdge, Overall: SeverityOK,
	}
	for _, p := range platforms {
		b := BudgetFor(p)
		v := PlatformVerdict{Platform: p, Title: b.Title, Status: SeverityOK, Budget: b, Hint: b.Hint}
		edge := maxEdge
		if edge > b.MaxEdge {
			v.Status = SeverityFail
			v.Messages = append(v.Messages, fmt.Sprintf("边长 %d 超过上限 %d", edge, b.MaxEdge))
		} else if edge > b.RecommendedEdge {
			if v.Status != SeverityFail {
				v.Status = SeverityWarn
			}
			v.Messages = append(v.Messages, fmt.Sprintf("边长 %d 超过推荐 %d", edge, b.RecommendedEdge))
		}
		if pixels > b.MaxPixels {
			v.Status = SeverityFail
			v.Messages = append(v.Messages, fmt.Sprintf("像素数 %d 超过上限 %d", pixels, b.MaxPixels))
		} else if pixels > b.RecommendedPixels {
			if v.Status != SeverityFail {
				v.Status = SeverityWarn
			}
			v.Messages = append(v.Messages, fmt.Sprintf("像素数 %d 超过推荐 %d", pixels, b.RecommendedPixels))
		}
		if b.MaxFileBytes > 0 && in.FileBytes > b.MaxFileBytes {
			v.Status = SeverityFail
			v.Messages = append(v.Messages, fmt.Sprintf("文件 %s 超过上限 %s", formatBytes(in.FileBytes), formatBytes(b.MaxFileBytes)))
		} else if b.RecommendedFileBytes > 0 && in.FileBytes > b.RecommendedFileBytes {
			if v.Status != SeverityFail {
				v.Status = SeverityWarn
			}
			v.Messages = append(v.Messages, fmt.Sprintf("文件 %s 超过推荐 %s", formatBytes(in.FileBytes), formatBytes(b.RecommendedFileBytes)))
		}
		if b.WarnNPOT && in.NPOT {
			if v.Status == SeverityOK {
				v.Status = SeverityWarn
			}
			v.Messages = append(v.Messages, "非 2 的幂 (NPOT)，WebGL1/部分移动端 mip 与采样可能受限")
		}
		if b.RequireMultipleOf > 1 {
			if in.Width%b.RequireMultipleOf != 0 || in.Height%b.RequireMultipleOf != 0 {
				if v.Status != SeverityFail {
					v.Status = SeverityWarn
				}
				v.Messages = append(v.Messages, fmt.Sprintf("宽高应为 %d 的倍数（ASTC 块对齐）", b.RequireMultipleOf))
			}
		}
		if len(v.Messages) == 0 {
			v.Messages = append(v.Messages, "尺寸与体积在预算内")
		}
		report.Platforms = append(report.Platforms, v)
		if severityRank(v.Status) > severityRank(report.Overall) {
			report.Overall = v.Status
		}
	}
	sort.Slice(report.Platforms, func(i, j int) bool {
		return severityRank(report.Platforms[i].Status) > severityRank(report.Platforms[j].Status)
	})
	report.Summary = summarize(report)
	return report
}

func severityRank(s Severity) int {
	switch s {
	case SeverityFail:
		return 3
	case SeverityWarn:
		return 2
	default:
		return 1
	}
}

func summarize(r BudgetReport) string {
	fails, warns := 0, 0
	for _, p := range r.Platforms {
		switch p.Status {
		case SeverityFail:
			fails++
		case SeverityWarn:
			warns++
		}
	}
	if fails > 0 {
		return fmt.Sprintf("%d 个平台判定不合理，%d 个平台有警告", fails, warns)
	}
	if warns > 0 {
		return fmt.Sprintf("全部平台可用，但 %d 个平台建议优化", warns)
	}
	return "全部平台预算内"
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
