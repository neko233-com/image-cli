package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neko233-com/image-cli/internal/img"
	"golang.org/x/image/font"
)

type keyOpts struct {
	color      string
	similarity float64
	smoothness float64
	despill    bool
	sample     string
	choke      int
	feather    float64
	bg         string
	out        string
}

func bindKey(fs *flag.FlagSet, k *keyOpts) {
	fs.StringVar(&k.color, "color", "green", "键色: green|screen|blue|magenta|#RRGGBB")
	fs.Float64Var(&k.similarity, "similarity", 0.28, "抠除强度，越大吃得越干净")
	fs.Float64Var(&k.smoothness, "smoothness", 0.15, "边缘过渡")
	fs.BoolVar(&k.despill, "despill", true, "去掉绿/蓝溢色")
	fs.StringVar(&k.sample, "sample", "", "从画面取样键色，格式 x,y")
	fs.IntVar(&k.choke, "choke", 0, "收缩抠像边缘像素")
	fs.Float64Var(&k.feather, "feather", 0, "羽化透明边缘")
	fs.StringVar(&k.bg, "bg", "", "叠到这张背景上")
	fs.StringVar(&k.out, "o", "", "输出路径")
}

func runKey(args []string) error {
	fs := flag.NewFlagSet("key", flag.ExitOnError)
	var k keyOpts
	bindKey(fs, &k)
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli key [-color green|screen|blue] [-bg plate.jpg] [-o out.png] <input>")
	}
	return applyKeyFile(fs.Arg(0), k)
}

func applyKeyFile(path string, k keyOpts) error {
	im, _, err := img.Load(path)
	if err != nil {
		return err
	}
	out, err := applyKey(im, path, k)
	if err != nil {
		return err
	}
	dest := k.out
	if dest == "" {
		dest = img.OutputPath(path, "", "png")
	}
	return img.Save(dest, out, img.EncodeOptions{Format: "png"})
}

func applyKey(im image.Image, path string, k keyOpts) (image.Image, error) {
	key, err := img.ParseKeyColor(k.color)
	if err != nil {
		return nil, err
	}
	if k.sample != "" {
		key, err = sampleColor(im, k.sample)
		if err != nil {
			return nil, err
		}
	}
	cut := img.ChromaKey(im, img.KeyOptions{
		Key: key, Similarity: k.similarity, Smoothness: k.smoothness, Despill: k.despill,
	})
	var result image.Image = cut
	if k.choke > 0 || k.feather > 0 {
		result = img.Matte(result, k.choke, 0, k.feather)
	}
	if k.bg != "" {
		bg, _, err := img.Load(k.bg)
		if err != nil {
			return nil, err
		}
		result = img.Composite(bg, result, 0, 0)
	}
	_ = path
	return result, nil
}

func sampleColor(im image.Image, spec string) (color.NRGBA, error) {
	xs, ys, ok := strings.Cut(spec, ",")
	if !ok {
		return color.NRGBA{}, fmt.Errorf("-sample 需要 x,y")
	}
	x, err1 := strconv.Atoi(strings.TrimSpace(xs))
	y, err2 := strconv.Atoi(strings.TrimSpace(ys))
	if err1 != nil || err2 != nil {
		return color.NRGBA{}, fmt.Errorf("-sample 需要 x,y")
	}
	b := im.Bounds()
	r, g, bl, a := im.At(b.Min.X+x, b.Min.Y+y).RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(a >> 8)}, nil
}

func runMatte(args []string) error {
	fs := flag.NewFlagSet("matte", flag.ExitOnError)
	choke := fs.Int("choke", 0, "收缩透明边缘")
	expand := fs.Int("expand", 0, "扩张透明边缘")
	feather := fs.Float64("feather", 0, "羽化")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli matte [-choke 1] [-feather 1] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Matte(im, *choke, *expand, *feather), img.EncodeOptions{Format: "png"})
}

func runComposite(args []string) error {
	fs := flag.NewFlagSet("composite", flag.ExitOnError)
	fg := fs.String("fg", "", "前景")
	x := fs.Int("x", 0, "前景 X")
	y := fs.Int("y", 0, "前景 Y")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 || *fg == "" {
		return fmt.Errorf("usage: image-cli composite -fg cut.png [-x 0 -y 0] <background>")
	}
	bg, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	front, _, err := img.Load(*fg)
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Composite(bg, front, *x, *y), img.EncodeOptions{Format: "png"})
}

func runLetterbox(args []string) error {
	fs := flag.NewFlagSet("letterbox", flag.ExitOnError)
	aspect := fs.Float64("aspect", 2.39, "目标宽高比")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli letterbox [-aspect 2.39] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Letterbox(im, *aspect, color.Black), img.EncodeOptions{Format: "png"})
}

func runVignette(args []string) error {
	fs := flag.NewFlagSet("vignette", flag.ExitOnError)
	strength := fs.Float64("strength", 0.6, "0-1")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli vignette [-strength 0.6] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.Vignette(im, *strength), img.EncodeOptions{Format: "png"})
}

func runSticker(args []string) error {
	fs := flag.NewFlagSet("sticker", flag.ExitOnError)
	var k keyOpts
	bindKey(fs, &k)
	k.color = "green"
	stroke := fs.Int("stroke", 8, "描边宽度，0 表示不要")
	strokeColor := fs.String("stroke-color", "white", "描边颜色")
	shadow := fs.Bool("shadow", false, "加投影")
	circle := fs.Bool("circle", false, "裁成圆形")
	trim := fs.Bool("trim", true, "裁掉透明边")
	text := fs.String("text", "", "贴纸上的字")
	fontPath := fs.String("font", "", "字体文件")
	size := fs.Float64("size", 42, "字号")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli sticker [-color green|none] [-stroke 8] [-text 哈哈] <input>")
	}
	if k.color == "none" {
		im, _, err := img.Load(fs.Arg(0))
		if err != nil {
			return err
		}
		out, err := finishSticker(im, *stroke, *strokeColor, *shadow, *circle, *trim, *text, *fontPath, *size)
		if err != nil {
			return err
		}
		return img.Save(img.OutputPath(fs.Arg(0), k.out, "png"), out, img.EncodeOptions{Format: "png"})
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	cut, err := applyKey(im, fs.Arg(0), k)
	if err != nil {
		return err
	}
	out, err := finishSticker(cut, *stroke, *strokeColor, *shadow, *circle, *trim, *text, *fontPath, *size)
	if err != nil {
		return err
	}
	return img.Save(img.OutputPath(fs.Arg(0), k.out, "png"), out, img.EncodeOptions{Format: "png"})
}

func finishSticker(im image.Image, stroke int, strokeColor string, shadow, circle, trim bool, text, fontPath string, size float64) (image.Image, error) {
	cur := im
	if trim {
		cur = img.TrimAlpha(cur)
	}
	if circle {
		cur = img.CircleMask(cur)
	}
	if shadow {
		cur = img.DropShadow(cur, 6, 8, 2, 0.45)
	}
	if stroke > 0 {
		col, err := img.ParseKeyColor(strokeColor)
		if err != nil {
			return nil, err
		}
		cur = img.Stroke(cur, stroke, col)
	}
	if strings.TrimSpace(text) != "" {
		face, err := openFace(fontPath, size)
		if err != nil {
			return nil, err
		}
		defer closeFace(face)
		b := cur.Bounds()
		cur = img.CenterText(cur, text, face, b.Dy()-12, color.White, color.Black, 2)
	}
	return cur, nil
}

func runMeme(args []string) error {
	fs := flag.NewFlagSet("meme", flag.ExitOnError)
	top := fs.String("top", "", "上方文字")
	bottom := fs.String("bottom", "", "下方文字")
	fontPath := fs.String("font", "", "字体文件")
	size := fs.Float64("size", 0, "字号，0 为宽度的 1/10")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli meme -top 上面 -bottom 下面 <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	px := *size
	if px <= 0 {
		px = float64(im.Bounds().Dx()) / 10
		if px < 16 {
			px = 16
		}
	}
	face, err := openFace(*fontPath, px)
	if err != nil {
		return err
	}
	defer closeFace(face)
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.MemeCaption(im, *top, *bottom, face), img.EncodeOptions{Format: "png"})
}

func runCaption(args []string) error {
	fs := flag.NewFlagSet("caption", flag.ExitOnError)
	text := fs.String("text", "", "文字，\\n 换行")
	x := fs.Int("x", 16, "左上角 X，实际按基线绘制")
	y := fs.Int("y", 48, "基线 Y")
	size := fs.Float64("size", 42, "字号")
	fontPath := fs.String("font", "", "字体文件")
	out := fs.String("o", "", "输出路径")
	_ = fs.Parse(args)
	if fs.NArg() != 1 || *text == "" {
		return fmt.Errorf("usage: image-cli caption -text 你好 [-font a.ttf] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	face, err := openFace(*fontPath, *size)
	if err != nil {
		return err
	}
	defer closeFace(face)
	body := strings.ReplaceAll(*text, `\n`, "\n")
	dest := img.OutputPath(fs.Arg(0), *out, "png")
	return img.Save(dest, img.DrawText(im, body, face, *x, *y, color.White, color.Black, 2), img.EncodeOptions{Format: "png"})
}

func openFace(path string, size float64) (font.Face, error) {
	found, err := img.FindFont(path)
	if err != nil {
		return nil, err
	}
	return img.OpenFace(found, size)
}

func closeFace(face font.Face) {
	if c, ok := face.(interface{ Close() error }); ok {
		_ = c.Close()
	}
}

func runBatchKey(args []string) error {
	fs := flag.NewFlagSet("batch key", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	var k keyOpts
	bindKey(fs, &k)
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch key [-color green] [-out-dir dir] <dir>")
	}
	return batchProcess(fs.Arg(0), walk, "png", func(path string) (image.Image, error) {
		im, _, err := img.Load(path)
		if err != nil {
			return nil, err
		}
		return applyKey(im, path, k)
	})
}

func runBatchSticker(args []string) error {
	fs := flag.NewFlagSet("batch sticker", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	var k keyOpts
	bindKey(fs, &k)
	stroke := fs.Int("stroke", 8, "描边宽度")
	strokeColor := fs.String("stroke-color", "white", "描边颜色")
	trim := fs.Bool("trim", true, "裁掉透明边")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli batch sticker [-color green] [-out-dir dir] <dir>")
	}
	return batchProcess(fs.Arg(0), walk, "png", func(path string) (image.Image, error) {
		im, _, err := img.Load(path)
		if err != nil {
			return nil, err
		}
		var cur image.Image = im
		if k.color != "none" {
			cur, err = applyKey(im, path, k)
			if err != nil {
				return nil, err
			}
		}
		return finishSticker(cur, *stroke, *strokeColor, false, false, *trim, "", "", 0)
	})
}

func batchProcess(root string, walk walkFlags, ext string, fn func(path string) (image.Image, error)) error {
	paths, err := walk.paths(root)
	if err != nil {
		return err
	}
	dests := make([]string, len(paths))
	err = walk.parallel(len(paths), func(i int) error {
		out, err := fn(paths[i])
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
		if err := img.Save(dest, out, img.EncodeOptions{Format: ext}); err != nil {
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
