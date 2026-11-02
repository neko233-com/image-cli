package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/neko233-com/image-cli/internal/img"
)

// walkFlags 目录批量处理通用参数。
type walkFlags struct {
	recursive  bool
	noRecurse  bool
	outDir     string
	mirrorTree bool
	jobs       int
	keepGoing  bool
}

func (w *walkFlags) bind(fs *flag.FlagSet) {
	fs.BoolVar(&w.recursive, "r", false, "递归子目录（对目录输入默认开启，可用 -no-recurse 关闭）")
	fs.BoolVar(&w.noRecurse, "no-recurse", false, "不递归，仅处理目录顶层文件")
	fs.StringVar(&w.outDir, "out-dir", "", "批量输出根目录")
	fs.BoolVar(&w.mirrorTree, "mirror-tree", true, "out-dir 下保留源目录相对结构")
	fs.IntVar(&w.jobs, "jobs", 0, "并行数，0=自动取逻辑 CPU 数")
	fs.BoolVar(&w.keepGoing, "continue", false, "单张失败后继续处理其余文件")
}

func (w walkFlags) walkOpts(root string) img.WalkOptions {
	rec := w.recursive
	if info, err := os.Stat(root); err == nil && info.IsDir() {
		if w.noRecurse {
			rec = false
		} else if !w.recursive {
			rec = true
		}
	}
	if w.recursive {
		rec = true
	}
	return img.WalkOptions{Recursive: rec}
}

func (w walkFlags) paths(root string) ([]string, error) {
	paths, err := img.ListRasters(root, w.walkOpts(root))
	if err != nil {
		return nil, err
	}
	if len(paths) > 1 {
		fmt.Fprintf(os.Stderr, "batch: files=%d workers=%s\n", len(paths), img.DescribeWorkers(w.jobs))
	}
	return paths, nil
}

func (w walkFlags) parallel(n int, fn func(i int) error) error {
	return img.Parallel(n, img.ParallelOptions{Jobs: w.jobs, Continue: w.keepGoing}, fn)
}
