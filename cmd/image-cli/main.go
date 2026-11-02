package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/neko233-com/image-cli/internal/game"
	"github.com/neko233-com/image-cli/internal/img"
	"github.com/neko233-com/image-cli/internal/updater"
	"github.com/neko233-com/image-cli/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "help", "-h", "--help":
		usage()
	case "version", "-v", "--version":
		runVersion(os.Args[2:])
	case "update", "upgrade":
		runUpdate(os.Args[2:])
	default:
		maybeLazyUpdate(cmd)
		if err := dispatch(cmd, os.Args[2:]); err != nil {
			fail("%v", err)
		}
	}
}

func maybeLazyUpdate(cmd string) {
	if os.Getenv("IMAGE_CLI_NO_UPDATE_CHECK") == "1" {
		return
	}
	switch cmd {
	case "update", "upgrade", "version", "help":
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		msg, err := updater.NotifyIfNewer(ctx, updater.NotifyOptions{
			Repo:           version.Repo,
			CurrentVersion: version.Version,
		})
		if err != nil || msg == "" {
			return
		}
		fmt.Fprintln(os.Stderr, msg)
	}()
}

func dispatch(cmd string, args []string) error {
	switch cmd {
	case "info":
		return runInfo(args)
	case "convert":
		return runConvert(args)
	case "optimize":
		return runOptimize(args)
	case "resize":
		return runResize(args)
	case "crop":
		return runCrop(args)
	case "rotate":
		return runRotate(args)
	case "flip":
		return runFlip(args)
	case "pad":
		return runPad(args)
	case "rgb":
		return runRGB(args)
	case "palette":
		return runPalette(args)
	case "gray", "grayscale":
		return runGray(args)
	case "blur":
		return runBlur(args)
	case "sharpen":
		return runSharpen(args)
	case "adjust":
		return runAdjust(args)
	case "pow2":
		return runPow2(args)
	case "split":
		return runSplit(args)
	case "watermark":
		return runWatermark(args)
	case "batch-convert":
		return runBatchConvert(args)
	case "check":
		return runCheck(args)
	case "game":
		return runGame(args)
	case "mipmap":
		return runMipmap(args)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func runUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	_ = fs.Parse(args)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	result, err := updater.Run(ctx, updater.Options{
		Repo:           version.Repo,
		CurrentVersion: version.Version,
		BinaryName:     version.Binary,
	})
	if err != nil {
		fail("update: %v", err)
	}
	if !result.Updated {
		fmt.Printf("image-cli %s 已是最新 (%s)\n", result.CurrentVersion, result.LatestVersion)
		fmt.Printf("route: %s\n", result.Route)
		return
	}
	fmt.Printf("image-cli 已更新: %s -> %s\n", result.CurrentVersion, result.LatestVersion)
	fmt.Printf("route: %s\n", result.Route)
}

func runInfo(args []string) error {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	exif := fs.Bool("exif", false, "读取 JPEG/TIFF EXIF")
	asJSON := fs.Bool("json", false, "JSON 输出")
	budget := fs.Bool("budget", false, "多平台尺寸/体积预算检测")
	platforms := fs.String("platform", "all", "budget 检测的平台列表")
	profile := fs.String("profile", "", "budget 按场景筛选平台")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli info [-exif] [-json] [-budget] <file>")
	}
	info, err := img.Inspect(fs.Arg(0), *exif)
	if err != nil {
		return err
	}
	if !*budget {
		return img.PrintInfo(info, *asJSON)
	}
	var plist []game.Platform
	if *profile != "" {
		p, err := game.ParseProfile(*profile)
		if err != nil {
			return err
		}
		plist = game.PlatformsForProfile(p)
	} else {
		plist, err = game.ParsePlatforms(*platforms)
		if err != nil {
			return err
		}
	}
	report := game.EvaluateBudget(game.BudgetInput{
		Width: info.Width, Height: info.Height, FileBytes: info.SizeBytes, NPOT: info.NPOT,
	}, plist)
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			img.FileInfo
			Budget game.BudgetReport `json:"budget"`
		}{info, report})
	}
	if err := img.PrintInfo(info, false); err != nil {
		return err
	}
	fmt.Println()
	printBudgetReport(report)
	return nil
}

func runConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	format := fs.String("format", "png", "目标格式: png/jpeg/webp/gif/bmp/tiff")
	out := fs.String("o", "", "输出路径")
	quality := fs.Int("quality", 90, "JPEG/WebP 质量 1-100")
	stripAlpha := fs.Bool("strip-alpha", false, "JPEG 等无 alpha 格式时铺白底")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli convert -format png [-o out] <input>")
	}
	in := fs.Arg(0)
	im, _, err := img.Load(in)
	if err != nil {
		return err
	}
	dest := *out
	if dest == "" {
		ext, _ := img.NormalizeFormat(*format)
		dest = img.OutputPath(in, "", ext)
	}
	return img.Save(dest, im, img.EncodeOptions{Format: *format, Quality: *quality, StripAlpha: *stripAlpha})
}

func runOptimize(args []string) error {
	fs := flag.NewFlagSet("optimize", flag.ExitOnError)
	format := fs.String("format", "", "输出格式（默认同输入或 png）")
	out := fs.String("o", "", "输出路径")
	quality := fs.Int("quality", 85, "有损质量")
	maxWidth := fs.Int("max-width", 0, "超过则等比缩小")
	maxHeight := fs.Int("max-height", 0, "超过则等比缩小")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli optimize [-format webp] [-max-width N] <input>")
	}
	in := fs.Arg(0)
	im, srcFmt, err := img.Load(in)
	if err != nil {
		return err
	}
	if *maxWidth > 0 || *maxHeight > 0 {
		im = img.Resize(im, *maxWidth, *maxHeight, img.ResizeFit)
	}
	outFmt := *format
	if outFmt == "" {
		outFmt = srcFmt
		if outFmt == "jpeg" {
			outFmt = "jpeg"
		}
	}
	dest := *out
	if dest == "" {
		ext, _ := img.NormalizeFormat(outFmt)
		dest = img.OutputPath(in, "", ext)
	}
	return img.Save(dest, im, img.EncodeOptions{Format: outFmt, Quality: *quality, StripAlpha: outFmt == "jpeg"})
}

func runResize(args []string) error {
	fs := flag.NewFlagSet("resize", flag.ExitOnError)
	w := fs.Int("w", 0, "目标宽度")
	h := fs.Int("h", 0, "目标高度")
	mode := fs.String("mode", "fit", "fit|fill|exact|width|height")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli resize -w 512 [-h 512] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	outIm := img.Resize(im, *w, *h, img.ResizeMode(*mode))
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, outIm, img.EncodeOptions{Format: "png"})
}

func runCrop(args []string) error {
	fs := flag.NewFlagSet("crop", flag.ExitOnError)
	x := fs.Int("x", 0, "左上角 X")
	y := fs.Int("y", 0, "左上角 Y")
	w := fs.Int("w", 0, "宽度")
	h := fs.Int("h", 0, "高度")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 || *w <= 0 || *h <= 0 {
		return fmt.Errorf("usage: image-cli crop -x 0 -y 0 -w 100 -h 100 <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	outIm := img.Crop(im, *x, *y, *w, *h)
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, outIm, img.EncodeOptions{Format: "png"})
}

func runRotate(args []string) error {
	fs := flag.NewFlagSet("rotate", flag.ExitOnError)
	angle := fs.Float64("angle", 90, "旋转角度（度）")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli rotate [-angle 90] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Rotate(im, *angle), img.EncodeOptions{Format: "png"})
}

func runFlip(args []string) error {
	fs := flag.NewFlagSet("flip", flag.ExitOnError)
	horizontal := fs.Bool("h", false, "水平翻转")
	vertical := fs.Bool("v", false, "垂直翻转")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli flip [-h] [-v] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	if !*horizontal && !*vertical {
		*horizontal = true
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Flip(im, *horizontal, *vertical), img.EncodeOptions{Format: "png"})
}

func runPad(args []string) error {
	fs := flag.NewFlagSet("pad", flag.ExitOnError)
	top := fs.Int("top", 0, "上")
	right := fs.Int("right", 0, "右")
	bottom := fs.Int("bottom", 0, "下")
	left := fs.Int("left", 0, "左")
	colorStr := fs.String("color", "transparent", "填充色 #RRGGBB 或 transparent")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli pad -top 10 -right 10 ... <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	bg, err := parseColor(*colorStr)
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Pad(im, *top, *right, *bottom, *left, bg), img.EncodeOptions{Format: "png"})
}

func runRGB(args []string) error {
	fs := flag.NewFlagSet("rgb", flag.ExitOnError)
	outDir := fs.String("out-dir", ".", "输出目录")
	prefix := fs.String("prefix", "", "文件名前缀")
	raw := fs.Bool("raw", false, "导出合并 .rgb 原始字节")
	separate := fs.Bool("separate", true, "导出 R/G/B(/A) 灰度 PNG")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli rgb [-out-dir dir] [-raw] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	if *prefix == "" {
		*prefix = strings.TrimSuffix(filepath.Base(fs.Arg(0)), filepath.Ext(fs.Arg(0)))
	}
	paths, err := img.ExtractRGB(im, img.RGBExtractOptions{
		OutDir: *outDir, Prefix: *prefix, RawRGB: *raw, Separate: *separate,
	})
	if err != nil {
		return err
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}

func runPalette(args []string) error {
	fs := flag.NewFlagSet("palette", flag.ExitOnError)
	n := fs.Int("n", 16, "颜色数量上限")
	asJSON := fs.Bool("json", false, "JSON 输出")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli palette [-n 16] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	entries, err := img.ExtractPalette(im, *n)
	if err != nil {
		return err
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}
	for _, e := range entries {
		fmt.Printf("%s\t%d\n", e.Hex, e.Count)
	}
	return nil
}

func runGray(args []string) error {
	fs := flag.NewFlagSet("gray", flag.ExitOnError)
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli gray <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Grayscale(im), img.EncodeOptions{Format: "png"})
}

func runBlur(args []string) error {
	fs := flag.NewFlagSet("blur", flag.ExitOnError)
	sigma := fs.Float64("sigma", 2, "高斯 sigma")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli blur [-sigma 2] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Blur(im, *sigma), img.EncodeOptions{Format: "png"})
}

func runSharpen(args []string) error {
	fs := flag.NewFlagSet("sharpen", flag.ExitOnError)
	sigma := fs.Float64("sigma", 1, "锐化强度")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli sharpen <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Sharpen(im, *sigma), img.EncodeOptions{Format: "png"})
}

func runAdjust(args []string) error {
	fs := flag.NewFlagSet("adjust", flag.ExitOnError)
	brightness := fs.Float64("brightness", 0, "亮度 ±%")
	contrast := fs.Float64("contrast", 0, "对比度 ±%")
	saturation := fs.Float64("saturation", 0, "饱和度 ±%")
	gamma := fs.Float64("gamma", 0, "伽马 (>0 启用)")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli adjust [-brightness 10] [-contrast 5] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	if *brightness != 0 {
		im = img.AdjustBrightness(im, *brightness)
	}
	if *contrast != 0 {
		im = img.AdjustContrast(im, *contrast)
	}
	if *saturation != 0 {
		im = img.AdjustSaturation(im, *saturation)
	}
	if *gamma > 0 {
		im = img.AdjustGamma(im, *gamma)
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, im, img.EncodeOptions{Format: "png"})
}

func runPow2(args []string) error {
	fs := flag.NewFlagSet("pow2", flag.ExitOnError)
	square := fs.Bool("square", false, "扩成正方形 POT")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli pow2 [-square] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.PowerOfTwoCanvas(im, *square), img.EncodeOptions{Format: "png"})
}

func runSplit(args []string) error {
	fs := flag.NewFlagSet("split", flag.ExitOnError)
	cols := fs.Int("cols", 1, "列数")
	rows := fs.Int("rows", 1, "行数")
	outDir := fs.String("out-dir", ".", "输出目录")
	prefix := fs.String("prefix", "tile", "文件名前缀")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli split -cols 4 -rows 4 <spritesheet>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	tiles := img.SplitGrid(im, *cols, *rows)
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	for i, tile := range tiles {
		name := fmt.Sprintf("%s_%02d.png", *prefix, i)
		path := filepath.Join(*outDir, name)
		if err := img.Save(path, tile, img.EncodeOptions{Format: "png"}); err != nil {
			return err
		}
		fmt.Println(path)
	}
	return nil
}

func runWatermark(args []string) error {
	fs := flag.NewFlagSet("watermark", flag.ExitOnError)
	mark := fs.String("mark", "", "水印图路径")
	x := fs.Int("x", 0, "X")
	y := fs.Int("y", 0, "Y")
	opacity := fs.Float64("opacity", 0.5, "不透明度 0-1")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 || *mark == "" {
		return fmt.Errorf("usage: image-cli watermark -mark logo.png <base>")
	}
	base, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	wm, _, err := img.Load(*mark)
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Watermark(base, wm, *x, *y, *opacity), img.EncodeOptions{Format: "png"})
}

func runBatchConvert(args []string) error {
	fs := flag.NewFlagSet("batch-convert", flag.ExitOnError)
	format := fs.String("format", "png", "目标格式")
	outDir := fs.String("out-dir", "", "输出目录（默认同目录）")
	recursive := fs.Bool("r", false, "递归子目录")
	quality := fs.Int("quality", 90, "JPEG/WebP 质量")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch-convert -format png [-out-dir dir] <dir|file>")
	}
	root := fs.Arg(0)
	return img.EachRaster(root, *recursive, func(path string) error {
		im, _, err := img.Load(path)
		if err != nil {
			return err
		}
		ext, _ := img.NormalizeFormat(*format)
		dir := filepath.Dir(path)
		if *outDir != "" {
			dir = *outDir
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		dest := filepath.Join(dir, base+"."+ext)
		if err := img.Save(dest, im, img.EncodeOptions{Format: *format, Quality: *quality, StripAlpha: ext == "jpeg"}); err != nil {
			return err
		}
		fmt.Println(dest)
		return nil
	})
}

func parseColor(s string) (color.Color, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "transparent" {
		return color.Transparent, nil
	}
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return nil, fmt.Errorf("invalid color %q", s)
	}
	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, fmt.Errorf("invalid color #%s", s)
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "image-cli: "+format+"\n", args...)
	os.Exit(2)
}

func usage() {
	fmt.Fprintf(os.Stderr, `image-cli %s — 跨平台图片处理 CLI（游戏资源 / 后期常用）

安装:
  curl -fsSL https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.sh | bash
  irm https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex

用法:
  image-cli <command> [flags] [args]

命令:
  info            图片元信息（-budget 多平台预算检测）
  check           专用于平台尺寸/体积是否合理（CI 友好）
  convert         格式转换 png/jpeg/webp/gif/bmp/tiff
  batch-convert   批量转换目录
  optimize        游戏向压缩（可选缩放到 max-width/height）
  resize crop rotate flip pad
  rgb             提取 R/G/B/A 通道或 raw RGB
  palette         主色 / 调色板统计
  gray blur sharpen adjust
  pow2            扩边到 2 的幂（可选 square）
  split           精灵表切格
  watermark       叠加水印
  mipmap          显式生成 mipmap PNG 链（默认 UI 流程不启用）
  game prepare    游戏资源一站式预处理（profile / ASTC / BC）
  game batch-prepare  批量预处理
  game astc|bc|pack-orm  纹理压缩与 ORM 打包
  game profiles   打印各场景默认策略 JSON
  update          检查并更新到最新 Release
  version         显示版本、GitHub 仓库与安装地址（-json）

环境变量:
  IMAGE_CLI_NO_UPDATE_CHECK=1     关闭懒更新提示
  IMAGE_CLI_UPDATE_MIRRORS        GitHub 加速镜像列表（逗号分隔）
  IMAGE_CLI_DIRECT_ONLY=1         仅直连 GitHub

`, version.String())
	fmt.Fprintf(os.Stderr, "平台: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
