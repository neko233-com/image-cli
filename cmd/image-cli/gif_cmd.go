package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neko233-com/image-cli/internal/img"
)

func runGIF(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: image-cli gif <info|extract|pack>")
	}
	switch args[0] {
	case "info":
		return runGIFInfo(args[1:])
	case "extract":
		return runGIFExtract(args[1:])
	case "pack":
		return runGIFPack(args[1:])
	default:
		return fmt.Errorf("unknown gif subcommand: %s", args[0])
	}
}

func runGIFInfo(args []string) error {
	fs := flag.NewFlagSet("gif info", flag.ExitOnError)
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli gif info <file.gif>")
	}
	info, err := img.InspectGIF(fs.Arg(0))
	if err != nil {
		return err
	}
	fmt.Printf("frames: %d\n", info.Frames)
	fmt.Printf("size:   %dx%d\n", info.Width, info.Height)
	fmt.Printf("loop:   %d\n", info.Loop)
	fmt.Printf("delays: %v\n", info.Delays)
	return nil
}

func runGIFExtract(args []string) error {
	fs := flag.NewFlagSet("gif extract", flag.ExitOnError)
	outDir := fs.String("out-dir", ".", "输出目录")
	prefix := fs.String("prefix", "frame", "文件名前缀")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli gif extract -out-dir dir <file.gif>")
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	paths, err := img.ExtractGIF(fs.Arg(0), *outDir, *prefix, func(path string, im image.Image) error {
		return img.Save(path, im, img.EncodeOptions{Format: "png"})
	})
	if err != nil {
		return err
	}
	for _, path := range paths {
		fmt.Println(path)
	}
	return nil
}

func runGIFPack(args []string) error {
	fs := flag.NewFlagSet("gif pack", flag.ExitOnError)
	out := fs.String("o", "out.gif", "输出 GIF")
	delay := fs.Int("delay", 10, "每帧延迟，单位 1/100 秒")
	recursive := fs.Bool("r", false, "目录输入时递归")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: image-cli gif pack -o out.gif [-delay 10] <dir|frames...>")
	}
	paths, err := collectFrames(fs.Args(), *recursive)
	if err != nil {
		return err
	}
	var frames []image.Image
	for _, path := range paths {
		im, _, err := img.Load(path)
		if err != nil {
			return err
		}
		frames = append(frames, im)
	}
	if err := img.SaveGIF(*out, frames, *delay); err != nil {
		return err
	}
	fmt.Println(*out)
	return nil
}

func collectFrames(args []string, recursive bool) ([]string, error) {
	if len(args) == 1 {
		info, err := os.Stat(args[0])
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			paths, err := img.ListRasters(args[0], img.WalkOptions{Recursive: recursive})
			if err != nil {
				return nil, err
			}
			sort.Strings(paths)
			return paths, nil
		}
	}
	var paths []string
	for _, arg := range args {
		ext := strings.ToLower(filepath.Ext(arg))
		if ext == "" {
			return nil, fmt.Errorf("not an image: %s", arg)
		}
		paths = append(paths, arg)
	}
	return paths, nil
}
