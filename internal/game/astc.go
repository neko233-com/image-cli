package game

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ASTCOptions 调用 ARM astcenc 进行块压缩。
type ASTCOptions struct {
	Block   string // 4x4 default
	Quality string // medium
	Mode    string // L sRGB, H HDR, norm normal map
	Mipmap  bool
}

func EncodeASTC(inputPNG, outputASTC string, opts ASTCOptions) error {
	tool, err := astcencPath()
	if err != nil {
		return err
	}
	block := opts.Block
	if block == "" {
		block = "4x4"
	}
	quality := opts.Quality
	if quality == "" {
		quality = "medium"
	}
	flag := "-cl"
	switch strings.ToLower(opts.Mode) {
	case "h", "hdr":
		flag = "-ch"
	case "norm", "normal", "n":
		flag = "-cn"
	case "linear", "l":
		flag = "-cl"
	case "srgb", "s", "":
		flag = "-cs"
	}
	args := []string{flag, inputPNG, outputASTC, block, "-quality", quality}
	if opts.Mipmap {
		args = append(args, "-mipmap")
	}
	if _, err := runCmd(filepath.Dir(inputPNG), tool, args...); err != nil {
		return err
	}
	if _, err := os.Stat(outputASTC); err != nil {
		return fmt.Errorf("astc output missing: %s", outputASTC)
	}
	return nil
}

func ValidateBlock(block string) error {
	switch block {
	case "4x4", "5x5", "6x6", "8x8", "10x10", "12x12":
		return nil
	default:
		return fmt.Errorf("invalid astc block %q", block)
	}
}
