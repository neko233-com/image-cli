package img

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// DrawText 在图上写字。strokeRadius > 0 时先画描边，表情包常用白字黑边。
func DrawText(src image.Image, text string, face font.Face, x, y int, fill, stroke color.Color, strokeRadius int) *image.NRGBA {
	dst := cloneNRGBA(src)
	if strings.TrimSpace(text) == "" || face == nil {
		return dst
	}
	lines := strings.Split(text, "\n")
	metrics := face.Metrics()
	lineH := metrics.Height.Ceil()
	if lineH < 1 {
		lineH = face.Metrics().Ascent.Ceil() + 2
	}
	if strokeRadius > 0 && stroke != nil {
		for _, line := range lines {
			for oy := -strokeRadius; oy <= strokeRadius; oy++ {
				for ox := -strokeRadius; ox <= strokeRadius; ox++ {
					if ox == 0 && oy == 0 {
						continue
					}
					drawLine(dst, face, line, x+ox, y+oy, stroke)
				}
			}
			y += lineH
		}
		y -= lineH * len(lines)
	}
	for _, line := range lines {
		drawLine(dst, face, line, x, y, fill)
		y += lineH
	}
	return dst
}

func drawLine(dst *image.NRGBA, face font.Face, text string, x, y int, col color.Color) {
	if col == nil {
		return
	}
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

// TextWidth 返回一行文字的像素宽度。
func TextWidth(face font.Face, text string) int {
	return font.MeasureString(face, text).Ceil()
}

// OpenFace 从字体文件创建字面。支持 ttf/otf，也支持 ttc 的第一个字体。
func OpenFace(path string, size float64) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		size = 48
	}
	var parsed *opentype.Font
	parsed, err = opentype.Parse(data)
	if err != nil {
		col, err2 := opentype.ParseCollection(data)
		if err2 != nil {
			return nil, err
		}
		parsed, err = col.Font(0)
		if err != nil {
			return nil, err
		}
	}
	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: size, DPI: 72, Hinting: font.HintingFull,
	})
}

// FindFont 按 -font、IMAGE_CLI_FONT 和系统常见中文字体查找。
func FindFont(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("font not found: %s", explicit)
		}
		return explicit, nil
	}
	if env := os.Getenv("IMAGE_CLI_FONT"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env, nil
		}
	}
	for _, path := range systemFonts() {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no font found; pass -font or set IMAGE_CLI_FONT")
}

func systemFonts() []string {
	switch runtime.GOOS {
	case "windows":
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = `C:\Windows`
		}
		dir := filepath.Join(windir, "Fonts")
		return []string{
			filepath.Join(dir, "msyh.ttc"),
			filepath.Join(dir, "msyhbd.ttc"),
			filepath.Join(dir, "simhei.ttf"),
			filepath.Join(dir, "simsun.ttc"),
			filepath.Join(dir, "arial.ttf"),
		}
	case "darwin":
		return []string{
			"/System/Library/Fonts/PingFang.ttc",
			"/System/Library/Fonts/STHeiti Light.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
			"/System/Library/Fonts/Supplemental/Arial.ttf",
		}
	default:
		return []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/freefont/FreeSans.ttf",
		}
	}
}

// CenterText 把文字画在水平居中、给定基线的位置。
func CenterText(src image.Image, text string, face font.Face, baseline int, fill, stroke color.Color, strokeRadius int) *image.NRGBA {
	dst := cloneNRGBA(src)
	lines := strings.Split(text, "\n")
	lineH := face.Metrics().Height.Ceil()
	if lineH < 1 {
		lineH = 16
	}
	y := baseline
	for _, line := range lines {
		w := TextWidth(face, line)
		x := (dst.Bounds().Dx() - w) / 2
		if strokeRadius > 0 && stroke != nil {
			for oy := -strokeRadius; oy <= strokeRadius; oy++ {
				for ox := -strokeRadius; ox <= strokeRadius; ox++ {
					if ox == 0 && oy == 0 {
						continue
					}
					drawLine(dst, face, line, x+ox, y+oy, stroke)
				}
			}
		}
		drawLine(dst, face, line, x, y, fill)
		y += lineH
	}
	return dst
}

// MemeCaption 在画面上方和下方写经典表情包字幕。
func MemeCaption(src image.Image, top, bottom string, face font.Face) *image.NRGBA {
	dst := cloneNRGBA(src)
	fill := color.White
	stroke := color.Black
	ascent := face.Metrics().Ascent.Ceil()
	if strings.TrimSpace(top) != "" {
		dst = CenterText(dst, top, face, ascent+8, fill, stroke, 2)
	}
	if strings.TrimSpace(bottom) != "" {
		lineH := face.Metrics().Height.Ceil()
		lines := strings.Split(bottom, "\n")
		baseline := dst.Bounds().Dy() - 12 - lineH*(len(lines)-1)
		dst = CenterText(dst, bottom, face, baseline, fill, stroke, 2)
	}
	return dst
}
