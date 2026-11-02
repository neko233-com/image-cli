package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neko233-com/image-cli/internal/game"
	"github.com/neko233-com/image-cli/internal/img"
)

type prepareFlags struct {
	profile    string
	target     string
	fallback   string
	mipmap     bool
	noMipmap   bool
	maxMip     int
	mipFilter  string
	pow2       bool
	noPow2     bool
	square     bool
	noSquare   bool
	maxEdge    int
	compress   bool
	noCompress bool
	format     string
	block      string
	quality    string
	astcMode   string
	manifest   bool
	premul     bool
	noPremul   bool
	walk       walkFlags
}

func (p *prepareFlags) bind(fs *flag.FlagSet) {
	fs.StringVar(&p.profile, "profile", "ui", "场景，或 auto（按文件名识别）")
	fs.StringVar(&p.target, "platform", "mobile", "mobile|pc")
	fs.StringVar(&p.fallback, "fallback", "ui", "profile=auto 且无法识别时使用的场景")
	fs.BoolVar(&p.mipmap, "mipmap", false, "生成 mipmap")
	fs.BoolVar(&p.noMipmap, "no-mipmap", false, "强制关闭 mipmap")
	fs.IntVar(&p.maxMip, "max-mip-levels", 0, "限制 mip 级数，0=直到 1x1")
	fs.StringVar(&p.mipFilter, "mip-filter", "box", "mip 缩放: box|lanczos")
	fs.BoolVar(&p.pow2, "pow2", false, "强制扩边到 2 的幂")
	fs.BoolVar(&p.noPow2, "no-pow2", false, "强制不扩边")
	fs.BoolVar(&p.square, "square", false, "强制正方形 POT")
	fs.BoolVar(&p.noSquare, "no-square", false, "强制不要求正方形")
	fs.IntVar(&p.maxEdge, "max-edge", 0, "最长边上限，0=用 profile 默认")
	fs.BoolVar(&p.compress, "compress", false, "强制 ASTC/BC")
	fs.BoolVar(&p.noCompress, "no-compress", false, "仅 PNG 预处理")
	fs.StringVar(&p.format, "format", "auto", "auto|astc|bc7|bc5|bc6h|bc4|bc3|none")
	fs.StringVar(&p.block, "block", "4x4", "ASTC 块大小")
	fs.StringVar(&p.quality, "quality", "medium", "astcenc 质量")
	fs.StringVar(&p.astcMode, "astc-mode", "", "ASTC: srgb|hdr|norm|linear")
	fs.BoolVar(&p.manifest, "manifest", true, "写入 .image-cli.json")
	fs.BoolVar(&p.premul, "premultiply", false, "强制预乘 alpha")
	fs.BoolVar(&p.noPremul, "no-premultiply", false, "强制不预乘")
	p.walk.bind(fs)
}

func (p prepareFlags) build(profile game.Profile) (game.PrepareOptions, error) {
	tg, err := game.ParseTarget(p.target)
	if err != nil {
		return game.PrepareOptions{}, err
	}
	def := game.Defaults(profile, tg)
	useCompress := def.Compress
	if p.noCompress {
		useCompress = false
	}
	if p.compress {
		useCompress = true
	}
	opts := game.PrepareOptions{
		Profile:       profile,
		Target:        tg,
		MaxMipLevels:  p.maxMip,
		MipFilter:     p.mipFilter,
		MaxEdge:       p.maxEdge,
		Format:        p.format,
		Block:         p.block,
		Quality:       p.quality,
		ASTCMode:      p.astcMode,
		WriteManifest: p.manifest,
		SkipCompress:  !useCompress,
		Compress:      boolPtr(useCompress),
	}
	switch {
	case p.noMipmap:
		opts.Mipmap = boolPtr(false)
	case p.mipmap:
		opts.Mipmap = boolPtr(true)
	}
	switch {
	case p.noPow2:
		opts.Pow2 = boolPtr(false)
	case p.pow2:
		opts.Pow2 = boolPtr(true)
	}
	switch {
	case p.noSquare:
		opts.Square = boolPtr(false)
	case p.square:
		opts.Square = boolPtr(true)
	}
	switch {
	case p.noPremul:
		opts.Premultiply = boolPtr(false)
	case p.premul:
		opts.Premultiply = boolPtr(true)
	}
	return opts, nil
}

func runPrepareCommand(name string, args []string) error {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	var pf prepareFlags
	pf.bind(fs)
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli %s [-profile auto|ui|vfx|albedo|...] [-platform mobile|pc] <file|dir>", name)
	}
	root := fs.Arg(0)
	mode, err := game.ParseProfile(pf.profile)
	if err != nil {
		return err
	}
	fallback := game.ProfileUI
	if mode == game.ProfileAuto {
		fallback, err = game.ParseProfile(pf.fallback)
		if err != nil {
			return err
		}
		if fallback == game.ProfileAuto {
			return fmt.Errorf("-fallback 不能是 auto")
		}
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		opts, err := pf.optionsFor(root, mode, fallback)
		if err != nil {
			return err
		}
		if pf.walk.outDir != "" {
			opts.OutDir = pf.walk.outDir
		}
		return writePrepare(root, opts)
	}

	paths, err := pf.walk.paths(root)
	if err != nil {
		return err
	}
	manifests := make([]game.Manifest, len(paths))
	err = pf.walk.parallel(len(paths), func(i int) error {
		path := paths[i]
		opts, err := pf.optionsFor(path, mode, fallback)
		if err != nil {
			return err
		}
		opts.OutDir = filepath.Dir(path)
		if pf.walk.outDir != "" {
			dest, err := img.ResolveOutputPath(root, path, pf.walk.outDir, pf.walk.mirrorTree, "")
			if err != nil {
				return err
			}
			opts.OutDir = filepath.Dir(dest)
		}
		res, err := game.Prepare(path, opts)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		manifests[i] = res.Manifest
		return nil
	})
	if err != nil && !pf.walk.keepGoing {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(manifests); encErr != nil {
		return encErr
	}
	return err
}

func (p prepareFlags) optionsFor(path string, mode, fallback game.Profile) (game.PrepareOptions, error) {
	profile := mode
	if mode == game.ProfileAuto {
		if detected, ok := game.DetectProfile(path); ok {
			profile = detected
		} else {
			profile = fallback
		}
	}
	return p.build(profile)
}

func writePrepare(path string, opts game.PrepareOptions) error {
	res, err := game.Prepare(path, opts)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res.Manifest)
}
