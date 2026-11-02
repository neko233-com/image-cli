package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/neko233-com/image-cli/internal/game"
	"github.com/neko233-com/image-cli/internal/img"
)

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	platforms := fs.String("platform", "all", "平台列表逗号分隔或 all")
	profile := fs.String("profile", "", "按场景筛选平台: ui|albedo|...")
	strict := fs.Bool("strict", false, "警告也返回退出码 1（CI 用）")
	asJSON := fs.Bool("json", false, "JSON 输出")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli check [-platform all] [-profile ui] <image>")
	}

	info, err := img.Inspect(fs.Arg(0), false)
	if err != nil {
		return err
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
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		printBudgetReport(report)
	}
	if report.Overall == game.SeverityFail || (*strict && report.Overall == game.SeverityWarn) {
		os.Exit(1)
	}
	return nil
}

func printBudgetReport(r game.BudgetReport) {
	fmt.Printf("image:      %dx%d (%d px)\n", r.Width, r.Height, r.Pixels)
	fmt.Printf("file:       %d bytes\n", r.FileBytes)
	fmt.Printf("overall:    %s — %s\n", r.Overall, r.Summary)
	fmt.Println("platforms:")
	for _, p := range r.Platforms {
		fmt.Printf("  [%s] %s — %s\n", p.Status, p.Title, p.Platform)
		for _, m := range p.Messages {
			fmt.Printf("      - %s\n", m)
		}
		if p.Hint != "" && p.Status != game.SeverityOK {
			fmt.Printf("      hint: %s\n", p.Hint)
		}
	}
}
