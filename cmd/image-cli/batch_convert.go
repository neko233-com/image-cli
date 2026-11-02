package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neko233-com/image-cli/internal/img"
)

func runBatchConvertDir(args []string) error {
	fs := flag.NewFlagSet("batch convert", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	format := fs.String("format", "png", "目标格式")
	quality := fs.Int("quality", 90, "JPEG/WebP 质量")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch convert -format png [-r] [-out-dir dir] <dir|file>")
	}
	return batchConvertRoot(fs.Arg(0), *format, *quality, walk)
}

func runBatchConvert(args []string) error {
	fs := flag.NewFlagSet("batch-convert", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	format := fs.String("format", "png", "目标格式")
	quality := fs.Int("quality", 90, "JPEG/WebP 质量")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch-convert -format png [-out-dir dir] <dir|file>")
	}
	return batchConvertRoot(fs.Arg(0), *format, *quality, walk)
}

func batchConvertRoot(root, format string, quality int, walk walkFlags) error {
	ext, err := img.NormalizeFormat(format)
	if err != nil {
		return err
	}
	paths, err := walk.paths(root)
	if err != nil {
		return err
	}
	dests := make([]string, len(paths))
	err = walk.parallel(len(paths), func(i int) error {
		im, _, err := img.Load(paths[i])
		if err != nil {
			return err
		}
		dest, err := img.ResolveOutputPath(root, paths[i], walk.outDir, walk.mirrorTree, ext)
		if err != nil {
			return err
		}
		if walk.outDir == "" {
			dest = img.OutputPath(paths[i], "", ext)
		} else if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := img.Save(dest, im, img.EncodeOptions{Format: format, Quality: quality, StripAlpha: ext == "jpeg"}); err != nil {
			return err
		}
		dests[i] = dest
		return nil
	})
	for _, dest := range dests {
		if dest != "" {
			fmt.Println(dest)
		}
	}
	return err
}
