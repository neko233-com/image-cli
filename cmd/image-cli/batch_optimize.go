package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neko233-com/image-cli/internal/img"
)

func batchOptimizeRoot(root, format string, quality, maxWidth, maxHeight int, walk walkFlags) error {
	paths, err := walk.paths(root)
	if err != nil {
		return err
	}
	dests := make([]string, len(paths))
	err = walk.parallel(len(paths), func(i int) error {
		path := paths[i]
		im, srcFmt, err := img.Load(path)
		if err != nil {
			return err
		}
		if maxWidth > 0 || maxHeight > 0 {
			im = img.Resize(im, maxWidth, maxHeight, img.ResizeFit)
		}
		outFmt := format
		if outFmt == "" {
			outFmt = srcFmt
		}
		ext, err := img.NormalizeFormat(outFmt)
		if err != nil {
			return err
		}
		dest, err := img.ResolveOutputPath(root, path, walk.outDir, walk.mirrorTree, ext)
		if err != nil {
			return err
		}
		if walk.outDir == "" {
			dest = img.OutputPath(path, "", ext)
		} else if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := img.Save(dest, im, img.EncodeOptions{Format: outFmt, Quality: quality, StripAlpha: ext == "jpeg"}); err != nil {
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
