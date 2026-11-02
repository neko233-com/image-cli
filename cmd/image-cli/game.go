package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neko233-com/image-cli/internal/game"
	"github.com/neko233-com/image-cli/internal/img"
)

func runGame(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: image-cli game <prepare|batch-prepare|astc|bc|pack-orm|profiles>")
	}
	switch args[0] {
	case "prepare":
		return runGamePrepare(args[1:])
	case "batch-prepare":
		return runGameBatchPrepare(args[1:])
	case "astc":
		return runGameASTC(args[1:])
	case "bc":
		return runGameBC(args[1:])
	case "pack-orm":
		return runGamePackORM(args[1:])
	case "profiles":
		printGameProfiles()
		return nil
	default:
		return fmt.Errorf("unknown game subcommand: %s", args[0])
	}
}

func runGamePrepare(args []string) error {
	return runPrepareCommand("game prepare", args)
}

func runGameBatchPrepare(args []string) error {
	return runPrepareCommand("game batch-prepare", args)
}

func runGameASTC(args []string) error {
	fs := flag.NewFlagSet("game astc", flag.ExitOnError)
	out := fs.String("o", "", "输出 .astc")
	block := fs.String("block", "4x4", "块大小")
	quality := fs.String("quality", "medium", "质量")
	mode := fs.String("mode", "srgb", "srgb|hdr|norm|linear")
	mipmap := fs.Bool("mipmap", false, "astcenc 内置 mipmap")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli game astc [-o out.astc] [-mipmap] <input.png>")
	}
	in := fs.Arg(0)
	dest := *out
	if dest == "" {
		dest = strings.TrimSuffix(in, filepath.Ext(in)) + ".astc"
	}
	return game.EncodeASTC(in, dest, game.ASTCOptions{
		Block: *block, Quality: *quality, Mode: *mode, Mipmap: *mipmap,
	})
}

func runGameBC(args []string) error {
	fs := flag.NewFlagSet("game bc", flag.ExitOnError)
	out := fs.String("o", "", "输出 .dds")
	format := fs.String("format", "bc7", "bc7|bc5|bc6h|bc4|bc3")
	mipmap := fs.Bool("mipmap", false, "生成 mipmap")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli game bc [-format bc7] <input.png>")
	}
	in := fs.Arg(0)
	dest := *out
	if dest == "" {
		dest = strings.TrimSuffix(in, filepath.Ext(in)) + "_" + *format + ".dds"
	}
	return game.EncodeBC(in, dest, game.BCOptions{Format: *format, Mipmap: *mipmap})
}

func runGamePackORM(args []string) error {
	fs := flag.NewFlagSet("game pack-orm", flag.ExitOnError)
	ao := fs.String("ao", "", "AO 灰度图")
	rough := fs.String("roughness", "", "粗糙度")
	metal := fs.String("metallic", "", "金属度")
	out := fs.String("o", "", "输出 PNG")
	_ = fs.Parse(args)
	if *ao == "" || *rough == "" || *metal == "" {
		return fmt.Errorf("usage: image-cli game pack-orm -ao a.png -roughness r.png -metallic m.png [-o orm.png]")
	}
	aoIm, _, err := img.Load(*ao)
	if err != nil {
		return err
	}
	rIm, _, err := img.Load(*rough)
	if err != nil {
		return err
	}
	mIm, _, err := img.Load(*metal)
	if err != nil {
		return err
	}
	packed, err := game.PackORM(aoIm, rIm, mIm)
	if err != nil {
		return err
	}
	dest := *out
	if dest == "" {
		dest = "orm_packed.png"
	}
	return img.Save(dest, packed, img.EncodeOptions{Format: "png"})
}

func runMipmap(args []string) error {
	fs := flag.NewFlagSet("mipmap", flag.ExitOnError)
	outDir := fs.String("out-dir", ".", "输出目录")
	maxLevels := fs.Int("max-levels", 0, "最大级数")
	filter := fs.String("filter", "box", "box|lanczos")
	prefix := fs.String("prefix", "", "文件名前缀")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli mipmap [-out-dir dir] <input>")
	}
	im, _, err := img.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	levels, err := game.GenerateMipmaps(im, *maxLevels, *filter)
	if err != nil {
		return err
	}
	base := *prefix
	if base == "" {
		base = strings.TrimSuffix(filepath.Base(fs.Arg(0)), filepath.Ext(fs.Arg(0)))
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	for _, lv := range levels {
		name := fmt.Sprintf("%s_mip%d.png", base, lv.Level)
		path := filepath.Join(*outDir, name)
		if err := img.Save(path, lv.Image, img.EncodeOptions{Format: "png"}); err != nil {
			return err
		}
		fmt.Println(path)
	}
	return nil
}

func printGameProfiles() {
	type row struct {
		Profile        string `json:"profile"`
		MobileCompress string `json:"mobile_compress"`
		PCCompress     string `json:"pc_compress"`
		MipmapDefault  bool   `json:"mipmap_default"`
		Pow2Default    bool   `json:"pow2_default"`
	}
	var rows []row
	for _, p := range game.AllProfiles() {
		m := game.Defaults(p, game.TargetMobile)
		pc := game.Defaults(p, game.TargetPC)
		rows = append(rows, row{
			Profile:        string(p),
			MobileCompress: m.CompressFormat,
			PCCompress:     pc.CompressFormat,
			MipmapDefault:  m.Mipmap,
			Pow2Default:    m.Pow2,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(rows)
}

func boolPtr(v bool) *bool { return &v }
