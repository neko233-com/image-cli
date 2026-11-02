package game

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/neko233-com/image-cli/internal/img"
)

type PrepareOptions struct {
	Profile     Profile
	Target      Target
	Mipmap      *bool
	MaxMipLevels int
	MipFilter   string
	Pow2        *bool
	Square      *bool
	MaxEdge     int
	Compress    *bool
	Format      string // auto|astc|bc7|bc5|bc6h|none
	Block       string
	Quality     string
	ASTCMode    string
	OutDir      string
	WriteManifest bool
	SkipCompress  bool // 仅输出 prepared PNG / mip PNG
	Premultiply   *bool
}

type PrepareResult struct {
	Manifest Manifest
}

func Prepare(input string, opts PrepareOptions) (PrepareResult, error) {
	if opts.Profile == "" {
		opts.Profile = ProfileUI
	}
	if opts.Target == "" {
		opts.Target = TargetMobile
	}
	def := Defaults(opts.Profile, opts.Target)

	mipmap := def.Mipmap
	if opts.Mipmap != nil {
		mipmap = *opts.Mipmap
	}
	pow2 := def.Pow2
	if opts.Pow2 != nil {
		pow2 = *opts.Pow2
	}
	square := def.Square
	if opts.Square != nil {
		square = *opts.Square
	}
	maxEdge := def.MaxEdge
	if opts.MaxEdge > 0 {
		maxEdge = opts.MaxEdge
	}
	compress := def.Compress
	if opts.Compress != nil {
		compress = *opts.Compress
	}
	if opts.SkipCompress {
		compress = false
	}
	compressFmt := def.CompressFormat
	if opts.Format != "" && opts.Format != "auto" {
		compressFmt = strings.ToLower(opts.Format)
	}
	if compress && (compressFmt == "" || compressFmt == "none") {
		if opts.Target == TargetPC {
			compressFmt = "bc7"
		} else {
			compressFmt = "astc"
		}
	}
	premul := def.Premultiply
	if opts.Premultiply != nil {
		premul = *opts.Premultiply
	}
	astcMode := def.ASTCMode
	if opts.ASTCMode != "" {
		astcMode = opts.ASTCMode
	}

	im, srcFmt, err := img.Load(input)
	if err != nil {
		return PrepareResult{}, err
	}

	var warnings []string
	b := im.Bounds()
	var fileBytes int64
	if st, err := os.Stat(input); err == nil {
		fileBytes = st.Size()
	}
	npot := !isPowerOfTwo(b.Dx()) || !isPowerOfTwo(b.Dy())
	budget := EvaluateBudget(BudgetInput{
		Width: b.Dx(), Height: b.Dy(), FileBytes: fileBytes, NPOT: npot, Usage: opts.Profile,
	}, PlatformsForProfile(opts.Profile))
	for _, pv := range budget.Platforms {
		if pv.Status == SeverityOK {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("budget/%s (%s): %s", pv.Platform, pv.Status, pv.Messages[0]))
	}
	if budget.Overall == SeverityFail {
		warnings = append(warnings, "budget: "+budget.Summary)
	}

	b = im.Bounds()
	if maxEdge > 0 && (b.Dx() > maxEdge || b.Dy() > maxEdge) {
		im = img.Resize(im, maxEdge, maxEdge, img.ResizeFit)
		warnings = append(warnings, fmt.Sprintf("scaled to max edge %d", maxEdge))
	}
	if pow2 {
		im = img.PowerOfTwoCanvas(im, square)
	}
	if premul {
		im = premultiplyAlpha(im)
	}

	outDir := opts.OutDir
	if outDir == "" {
		outDir = filepath.Dir(input)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return PrepareResult{}, err
	}
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	prepared := filepath.Join(outDir, base+"_prepared.png")
	if err := img.Save(prepared, im, img.EncodeOptions{Format: "png"}); err != nil {
		return PrepareResult{}, err
	}

	result := PrepareResult{Manifest: Manifest{
		Input:    filepath.Clean(input),
		Profile:  opts.Profile,
		Target:   opts.Target,
		Mipmap:   mipmap,
		Warnings: warnings,
		Meta: map[string]string{
			"source_format": srcFmt,
			"budget_overall": string(budget.Overall),
			"budget_summary": budget.Summary,
		},
	}}
	result.Manifest.Outputs = append(result.Manifest.Outputs, ManifestOutput{
		Kind: "prepared_png", Path: prepared, Width: im.Bounds().Dx(), Height: im.Bounds().Dy(),
	})

	if mipmap {
		filter := opts.MipFilter
		if filter == "" {
			filter = "box"
		}
		levels, err := GenerateMipmaps(im, opts.MaxMipLevels, filter)
		if err != nil {
			return PrepareResult{}, err
		}
		result.Manifest.MipLevels = len(levels)
		for _, lv := range levels {
			if lv.Level == 0 && compress {
				continue
			}
			name := fmt.Sprintf("%s_mip%d.png", base, lv.Level)
			path := filepath.Join(outDir, name)
			if err := img.Save(path, lv.Image, img.EncodeOptions{Format: "png"}); err != nil {
				return PrepareResult{}, err
			}
			result.Manifest.Outputs = append(result.Manifest.Outputs, ManifestOutput{
				Kind: "mip_png", Path: path, Level: lv.Level, Width: lv.Width, Height: lv.Height,
			})
		}
	}

	if !compress || compressFmt == "none" {
		return finishManifest(outDir, base, opts, result)
	}

	compressInput := prepared
	if mipmap {
		// astcenc / texconv 自行生成 mip 链
		compressInput = prepared
	}

	switch compressFmt {
	case "astc":
		if err := ValidateBlock(opts.Block); err != nil && opts.Block != "" {
			return PrepareResult{}, err
		}
		outASTC := filepath.Join(outDir, base+".astc")
		if err := EncodeASTC(compressInput, outASTC, ASTCOptions{
			Block: opts.Block, Quality: opts.Quality, Mode: astcMode, Mipmap: mipmap,
		}); err != nil {
			return PrepareResult{}, fmt.Errorf("astc: %w (install astcenc or set IMAGE_CLI_ASTCENC)", err)
		}
		result.Manifest.Outputs = append(result.Manifest.Outputs, ManifestOutput{Kind: "astc", Path: outASTC})
	case "bc7", "bc5", "bc6h", "bc4", "bc3":
		outDDS := filepath.Join(outDir, base+"_"+compressFmt+".dds")
		if err := EncodeBC(compressInput, outDDS, BCOptions{Format: compressFmt, Mipmap: mipmap}); err != nil {
			return PrepareResult{}, fmt.Errorf("bc: %w (install texconv or set IMAGE_CLI_TEXCONV)", err)
		}
		result.Manifest.Outputs = append(result.Manifest.Outputs, ManifestOutput{Kind: "dds", Path: outDDS})
	default:
		return PrepareResult{}, fmt.Errorf("unsupported compress format %q", compressFmt)
	}

	return finishManifest(outDir, base, opts, result)
}

func finishManifest(outDir, base string, opts PrepareOptions, result PrepareResult) (PrepareResult, error) {
	if opts.WriteManifest {
		path := filepath.Join(outDir, base+".image-cli.json")
		if err := WriteManifest(path, result.Manifest); err != nil {
			return PrepareResult{}, err
		}
		result.Manifest.Outputs = append(result.Manifest.Outputs, ManifestOutput{Kind: "manifest", Path: path})
	}
	return result, nil
}

func premultiplyAlpha(src image.Image) image.Image {
	b := src.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.At(x, y)
			r, g, bl, a := c.RGBA()
			if a == 0xffff {
				out.Set(x, y, c)
				continue
			}
			af := float64(a) / 65535
			out.Set(x, y, color.NRGBA{
				R: uint8(float64(r>>8) * af),
				G: uint8(float64(g>>8) * af),
				B: uint8(float64(bl>>8) * af),
				A: uint8(a >> 8),
			})
		}
	}
	return out
}

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}
