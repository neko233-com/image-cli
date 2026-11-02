package main

import (
	"flag"
	"fmt"
)

func runBatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: image-cli batch <check|convert|optimize|prepare|key|sticker> [flags] <dir|file>")
	}
	switch args[0] {
	case "check":
		return runBatchCheck(args[1:])
	case "convert":
		return runBatchConvertDir(args[1:])
	case "optimize":
		return runBatchOptimize(args[1:])
	case "prepare":
		return runGameBatchPrepare(args[1:])
	case "key":
		return runBatchKey(args[1:])
	case "sticker":
		return runBatchSticker(args[1:])
	default:
		return fmt.Errorf("unknown batch subcommand: %s", args[0])
	}
}

func runBatchCheck(args []string) error {
	fs := flag.NewFlagSet("batch check", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	platforms := fs.String("platform", "all", "平台列表")
	profile := fs.String("profile", "", "按场景筛选平台")
	domain := fs.String("domain", "", "规范: game|design|film|vfx|auto|all")
	strict := fs.Bool("strict", false, "警告也失败")
	asJSON := fs.Bool("json", false, "JSON 数组输出")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch check [-jobs N] [-domain film] <dir|file>")
	}
	return runCheckRoot(fs.Arg(0), walk, *platforms, *profile, *domain, *strict, *asJSON)
}

func runBatchOptimize(args []string) error {
	fs := flag.NewFlagSet("batch optimize", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	format := fs.String("format", "", "输出格式（默认同输入）")
	quality := fs.Int("quality", 85, "质量")
	maxWidth := fs.Int("max-width", 0, "超过则等比缩小")
	maxHeight := fs.Int("max-height", 0, "超过则等比缩小")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch optimize [-r] [-out-dir dir] <dir|file>")
	}
	return batchOptimizeRoot(fs.Arg(0), *format, *quality, *maxWidth, *maxHeight, walk)
}
