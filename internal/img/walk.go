package img

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// WalkOptions 控制目录遍历行为。
type WalkOptions struct {
	Recursive bool
	// SkipDirs 跳过目录名（仅匹配路径段 basename）。
	SkipDirs []string
}

func (o WalkOptions) withDefaults() WalkOptions {
	out := o
	if len(out.SkipDirs) == 0 {
		out.SkipDirs = []string{".git", ".svn", ".hg", "node_modules", "__MACOSX", ".idea", ".vscode"}
	}
	return out
}

func (o WalkOptions) shouldSkipDir(name string) bool {
	name = strings.ToLower(name)
	for _, s := range o.SkipDirs {
		if strings.EqualFold(name, s) {
			return true
		}
	}
	return false
}

// EachRasterWithOptions 遍历文件或目录下的 raster 图。
func EachRasterWithOptions(root string, opts WalkOptions, fn func(path string) error) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	opts = opts.withDefaults()
	if !info.IsDir() {
		if isRaster(root) {
			return fn(root)
		}
		return fmt.Errorf("not a raster image: %s", root)
	}
	recursive := opts.Recursive
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root {
				if !recursive {
					return fs.SkipDir
				}
				if opts.shouldSkipDir(d.Name()) {
					return fs.SkipDir
				}
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

// ListRasters 先收集路径，供并行处理使用。
func ListRasters(root string, opts WalkOptions) ([]string, error) {
	var paths []string
	err := EachRasterWithOptions(root, opts, func(path string) error {
		paths = append(paths, path)
		return nil
	})
	return paths, err
}

// EachRaster 兼容旧 API；recursive 为 false 时仅处理 root 目录下文件（不进入子目录）。
func EachRaster(root string, recursive bool, fn func(path string) error) error {
	return EachRasterWithOptions(root, WalkOptions{Recursive: recursive}, fn)
}

// ResolveOutputPath 计算批量输出路径；mirror 为 true 时在 outRoot 下保留相对 root 的子目录结构。
func ResolveOutputPath(root, inputPath, outRoot string, mirror bool, newExt string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	inputPath, err = filepath.Abs(inputPath)
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	fileName := base
	if newExt != "" {
		ext := strings.TrimPrefix(newExt, ".")
		fileName = base + "." + ext
	}
	if outRoot == "" {
		return filepath.Join(filepath.Dir(inputPath), fileName), nil
	}
	outRoot, err = filepath.Abs(outRoot)
	if err != nil {
		return "", err
	}
	if !mirror {
		return filepath.Join(outRoot, fileName), nil
	}
	rel, err := filepath.Rel(root, filepath.Dir(inputPath))
	if err != nil {
		return "", err
	}
	if rel == "." {
		return filepath.Join(outRoot, fileName), nil
	}
	destDir := filepath.Join(outRoot, rel)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(destDir, fileName), nil
}
