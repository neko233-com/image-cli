package img

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var rasterExts = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".webp": {}, ".gif": {}, ".bmp": {}, ".tif": {}, ".tiff": {},
}

// EachRaster 遍历目录下 raster 图片（不递归时可传 recursive=false）。
func EachRaster(root string, recursive bool, fn func(path string) error) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if isRaster(root) {
			return fn(root)
		}
		return fmt.Errorf("not a raster image: %s", root)
	}
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && !recursive {
				return fs.SkipDir
			}
			return nil
		}
		if isRaster(path) {
			return fn(path)
		}
		return nil
	}
	return filepath.WalkDir(root, walkFn)
}

func isRaster(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := rasterExts[ext]
	return ok
}
