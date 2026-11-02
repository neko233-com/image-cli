package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neko233-com/image-cli/internal/atlas"
	"github.com/neko233-com/image-cli/internal/img"
)

func runAtlas(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: image-cli atlas <pack|unpack>")
	}
	switch args[0] {
	case "pack":
		return runAtlasPack(args[1:])
	case "unpack":
		return runAtlasUnpack(args[1:])
	default:
		return fmt.Errorf("unknown atlas subcommand: %s", args[0])
	}
}

func runAtlasPack(args []string) error {
	fs := flag.NewFlagSet("atlas pack", flag.ExitOnError)
	format := fs.String("format", "cocos", "unity|ue|cocos")
	outDir := fs.String("out-dir", ".", "输出目录")
	name := fs.String("name", "atlas", "图集文件名（不含扩展名）")
	padding := fs.Int("padding", 2, "精灵间距")
	pot := fs.Bool("pot", true, "画布扩成 2 的幂")
	maxSize := fs.Int("max-size", 4096, "边长上限")
	recursive := fs.Bool("r", false, "递归子目录")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli atlas pack -format unity|ue|cocos -out-dir dir <sprites-dir>")
	}
	kind, err := normalizeAtlasFormat(*format)
	if err != nil {
		return err
	}
	root := fs.Arg(0)
	paths, err := img.ListRasters(root, img.WalkOptions{Recursive: *recursive})
	if err != nil {
		return err
	}
	sort.Strings(paths)
	var sprites []atlas.Sprite
	seen := map[string]int{}
	for _, path := range paths {
		im, _, err := img.Load(path)
		if err != nil {
			return err
		}
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		seen[base]++
		if seen[base] > 1 {
			base = fmt.Sprintf("%s_%d", base, seen[base])
		}
		sprites = append(sprites, atlas.Sprite{Name: base, Image: im})
	}
	sheet, err := atlas.Pack(sprites, atlas.PackOptions{Padding: *padding, MaxSize: *maxSize, POT: *pot})
	if err != nil {
		return err
	}
	sheet.Name = *name
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	pngName := *name + ".png"
	pngPath := filepath.Join(*outDir, pngName)
	if err := img.Save(pngPath, sheet.Image, img.EncodeOptions{Format: "png"}); err != nil {
		return err
	}
	var descPath string
	switch kind {
	case "unity":
		text, err := atlas.WriteUnityMeta(pngName, sheet)
		if err != nil {
			return err
		}
		descPath = pngPath + ".meta"
		if err := os.WriteFile(descPath, []byte(text), 0o644); err != nil {
			return err
		}
	case "ue":
		raw, err := atlas.WriteUE(pngName, sheet)
		if err != nil {
			return err
		}
		descPath = filepath.Join(*outDir, *name+".json")
		if err := os.WriteFile(descPath, raw, 0o644); err != nil {
			return err
		}
	case "cocos":
		text, err := atlas.WriteCocosPlist(pngName, sheet)
		if err != nil {
			return err
		}
		descPath = filepath.Join(*outDir, *name+".plist")
		if err := os.WriteFile(descPath, []byte(text), 0o644); err != nil {
			return err
		}
	}
	fmt.Println(pngPath)
	fmt.Println(descPath)
	return nil
}

func runAtlasUnpack(args []string) error {
	fs := flag.NewFlagSet("atlas unpack", flag.ExitOnError)
	format := fs.String("format", "auto", "auto|unity|ue|cocos")
	outDir := fs.String("out-dir", ".", "输出目录")
	_ = fs.Parse(args)
	if fs.NArg() < 1 || fs.NArg() > 2 {
		return fmt.Errorf("usage: image-cli atlas unpack [-format auto] -out-dir dir <desc> [sheet.png]")
	}
	descPath := fs.Arg(0)
	pngPath := ""
	if fs.NArg() == 2 {
		pngPath = fs.Arg(1)
		if looksLikeSheet(descPath) && !looksLikeSheet(pngPath) {
			descPath, pngPath = pngPath, descPath
		}
	}
	if pngPath == "" {
		pngPath = siblingSheet(descPath)
	}
	sheet, formatName, err := img.Load(pngPath)
	if err != nil {
		return err
	}
	_ = formatName
	raw, err := os.ReadFile(descPath)
	if err != nil {
		return err
	}
	kind := *format
	if kind == "auto" || kind == "" {
		kind = detectAtlasFormat(descPath, raw)
	}
	kind, err = normalizeAtlasFormat(kind)
	if err != nil {
		return err
	}
	var sprites []atlas.Sprite
	switch kind {
	case "unity":
		sprites, err = atlas.ParseUnityMeta(string(raw), sheet.Bounds().Dy())
	case "ue":
		_, sprites, err = atlas.ParseUE(raw)
	case "cocos":
		_, sprites, err = atlas.ParseCocosPlist(raw)
	}
	if err != nil {
		return err
	}
	sort.Slice(sprites, func(i, j int) bool { return sprites[i].Name < sprites[j].Name })
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	for _, sp := range sprites {
		im, err := atlas.Crop(sheet, sp)
		if err != nil {
			return err
		}
		name := sp.Name
		if name == "" {
			name = "sprite"
		}
		dest := filepath.Join(*outDir, name+".png")
		if err := img.Save(dest, im, img.EncodeOptions{Format: "png"}); err != nil {
			return err
		}
		fmt.Println(dest)
	}
	return nil
}

func normalizeAtlasFormat(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "unity", "unity-meta":
		return "unity", nil
	case "ue", "unreal", "ue-json":
		return "ue", nil
	case "cocos", "cocos-plist", "plist":
		return "cocos", nil
	default:
		return "", fmt.Errorf("unknown atlas format %q (unity|ue|cocos)", s)
	}
}

func detectAtlasFormat(path string, raw []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".meta":
		return "unity"
	case ".plist":
		return "cocos"
	case ".json":
		if strings.Contains(string(raw), "ue-atlas") || strings.Contains(string(raw), `"sprites"`) {
			return "ue"
		}
	}
	text := string(raw)
	if strings.Contains(text, "spriteSheet:") {
		return "unity"
	}
	if strings.Contains(text, "<plist") {
		return "cocos"
	}
	return "ue"
}

func looksLikeSheet(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".webp", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func siblingSheet(desc string) string {
	lower := strings.ToLower(desc)
	if strings.HasSuffix(lower, ".png.meta") {
		return strings.TrimSuffix(desc, ".meta")
	}
	base := strings.TrimSuffix(desc, filepath.Ext(desc))
	for _, ext := range []string{".png", ".webp"} {
		candidate := base + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return base + ".png"
}
