package game

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BCOptions 使用 DirectXTex texconv 生成 DDS（PC 常用）。
type BCOptions struct {
	Format string // bc7|bc5|bc6h|bc3
	Mipmap bool
}

func EncodeBC(inputPNG, outputDDS string, opts BCOptions) error {
	tool, err := texconvPath()
	if err != nil {
		return err
	}
	format := strings.ToLower(opts.Format)
	var dxgi string
	switch format {
	case "bc7", "bc7_unorm":
		dxgi = "BC7_UNORM"
	case "bc5", "bc5_unorm":
		dxgi = "BC5_UNORM"
	case "bc6h", "bc6h_ufloat":
		dxgi = "BC6H_UF16"
	case "bc3", "dxt5":
		dxgi = "BC3_UNORM"
	default:
		return fmt.Errorf("unsupported bc format %q", opts.Format)
	}
	outDir := filepath.Dir(outputDDS)
	base := strings.TrimSuffix(filepath.Base(outputDDS), filepath.Ext(outputDDS))
	args := []string{"-y", "-f", dxgi, "-o", outDir}
	if opts.Mipmap {
		args = append(args, "-m", "0")
	} else {
		args = append(args, "-m", "1")
	}
	args = append(args, inputPNG)
	if _, err := runCmd(outDir, tool, args...); err != nil {
		return err
	}
	generated := filepath.Join(outDir, base+".dds")
	// texconv 命名跟随输入文件名
	if _, err := os.Stat(generated); err != nil {
		alt := filepath.Join(outDir, strings.TrimSuffix(filepath.Base(inputPNG), filepath.Ext(inputPNG))+".dds")
		if _, err2 := os.Stat(alt); err2 != nil {
			return fmt.Errorf("texconv dds not found (expected %s or %s)", generated, alt)
		}
		generated = alt
	}
	if generated != outputDDS {
		if err := os.Rename(generated, outputDDS); err != nil {
			return err
		}
	}
	return nil
}
