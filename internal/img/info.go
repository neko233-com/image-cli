package img

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

type FileInfo struct {
	Path       string         `json:"path"`
	Format     string         `json:"format"`
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	Aspect     string         `json:"aspect"`
	SizeBytes  int64          `json:"size_bytes"`
	ColorModel string         `json:"color_model"`
	HasAlpha   bool           `json:"has_alpha"`
	Bits       int            `json:"bits"`
	NPOT       bool           `json:"npot"`
	PowerOfTwo bool           `json:"power_of_two"`
	EXIF       map[string]any `json:"exif,omitempty"`
}

func Inspect(path string, withEXIF bool) (FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	im, format, err := Load(path)
	if err != nil {
		return FileInfo{}, err
	}
	b := im.Bounds()
	w, h := b.Dx(), b.Dy()
	info := FileInfo{
		Path:       filepath.Clean(path),
		Format:     format,
		Width:      w,
		Height:     h,
		Aspect:     fmt.Sprintf("%d:%d", w, h),
		SizeBytes:  stat.Size(),
		ColorModel: colorModelName(im),
		HasAlpha:   imageHasAlpha(im),
		Bits:       channelBits(im),
		NPOT:       !isPowerOfTwo(w) || !isPowerOfTwo(h),
		PowerOfTwo: isPowerOfTwo(w) && isPowerOfTwo(h),
	}
	if withEXIF {
		info.EXIF = readEXIF(path)
	}
	return info, nil
}

func PrintInfo(info FileInfo, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}
	fmt.Printf("path:        %s\n", info.Path)
	fmt.Printf("format:      %s\n", info.Format)
	fmt.Printf("size:        %dx%d (%s)\n", info.Width, info.Height, info.Aspect)
	fmt.Printf("file_bytes:  %d\n", info.SizeBytes)
	fmt.Printf("color_model: %s\n", info.ColorModel)
	fmt.Printf("has_alpha:   %v\n", info.HasAlpha)
	fmt.Printf("bits:        %d\n", info.Bits)
	fmt.Printf("power_of_2:  %v\n", info.PowerOfTwo)
	fmt.Printf("npot:        %v\n", info.NPOT)
	if len(info.EXIF) > 0 {
		fmt.Println("exif:")
		for k, v := range info.EXIF {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	return nil
}

func readEXIF(path string) map[string]any {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	x, err := exif.Decode(f)
	if err != nil {
		return nil
	}
	out := map[string]any{}
	for _, tag := range []exif.FieldName{
		exif.Make, exif.Model, exif.DateTime, exif.Orientation,
		exif.FNumber, exif.ExposureTime, exif.ISOSpeedRatings, exif.FocalLength,
	} {
		if val, err := x.Get(tag); err == nil {
			out[string(tag)] = val.String()
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func colorModelName(im image.Image) string {
	switch im.ColorModel() {
	case color.RGBAModel, color.NRGBAModel, color.RGBA64Model, color.NRGBA64Model:
		return "rgba"
	case color.YCbCrModel:
		return "ycbcr"
	case color.GrayModel, color.Gray16Model:
		return "gray"
	default:
		name := fmt.Sprintf("%T", im.ColorModel())
		if i := strings.LastIndex(name, "."); i >= 0 {
			name = name[i+1:]
		}
		return strings.TrimSuffix(name, "Model")
	}
}

func channelBits(im image.Image) int {
	switch im.ColorModel() {
	case color.Gray16Model, color.RGBA64Model, color.NRGBA64Model:
		return 16
	default:
		return 8
	}
}

func imageHasAlpha(im image.Image) bool {
	b := im.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := im.At(x, y).RGBA()
			if a>>8 < 255 {
				return true
			}
		}
	}
	return false
}

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}
