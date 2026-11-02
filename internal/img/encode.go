package img

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
)

type EncodeOptions struct {
	Format   string
	Quality  int
	StripAlpha bool
}

// Save 按目标格式写入文件。
func Save(path string, src image.Image, opts EncodeOptions) error {
	format, err := NormalizeFormat(opts.Format)
	if err != nil {
		return err
	}
	im := src
	if opts.StripAlpha {
		im = flattenAlpha(src, color.White)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Encode(f, im, format, opts.Quality)
}

func Encode(w io.Writer, src image.Image, format string, quality int) error {
	format, err := NormalizeFormat(format)
	if err != nil {
		return err
	}
	if quality <= 0 {
		quality = 85
	}
	if quality > 100 {
		quality = 100
	}
	switch format {
	case "png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		return enc.Encode(w, src)
	case "jpeg":
		return jpeg.Encode(w, src, &jpeg.Options{Quality: quality})
	case "webp":
		level := nativewebp.DefaultCompression
		if quality >= 95 {
			level = nativewebp.BestCompression
		} else if quality <= 50 {
			level = nativewebp.BestSpeed
		}
		return nativewebp.Encode(w, src, &nativewebp.Options{CompressionLevel: level})
	case "gif":
		return gif.Encode(w, src, nil)
	case "bmp":
		return bmp.Encode(w, src)
	case "tiff":
		return tiff.Encode(w, src, &tiff.Options{Compression: tiff.Deflate})
	default:
		return fmt.Errorf("encode: %s", format)
	}
}

func flattenAlpha(src image.Image, bg color.Color) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, &image.Uniform{C: bg}, image.Point{}, draw.Src)
	draw.Draw(dst, b, src, b.Min, draw.Over)
	return dst
}
