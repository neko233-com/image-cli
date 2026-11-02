package img

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// Load 解码常见 raster 格式。
func Load(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".webp":
		im, err := webp.Decode(f)
		if err != nil {
			return nil, "", err
		}
		return im, "webp", nil
	case ".bmp", ".dib":
		im, err := bmp.Decode(f)
		if err != nil {
			return nil, "", err
		}
		return im, "bmp", nil
	case ".tif", ".tiff":
		im, err := tiff.Decode(f)
		if err != nil {
			return nil, "", err
		}
		return im, "tiff", nil
	default:
		im, format, err := image.Decode(f)
		if err != nil {
			return nil, "", err
		}
		return im, format, nil
	}
}

// NormalizeFormat 将用户输入格式名规范为小写扩展名（不含点）。
func NormalizeFormat(format string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	format = strings.TrimPrefix(format, ".")
	switch format {
	case "jpg", "jpeg":
		return "jpeg", nil
	case "png", "webp", "gif", "bmp", "tiff", "tif":
		if format == "tif" {
			return "tiff", nil
		}
		return format, nil
	default:
		return "", fmt.Errorf("unsupported format %q (use png, jpeg, webp, gif, bmp, tiff)", format)
	}
}
