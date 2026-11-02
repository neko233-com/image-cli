package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/neko233-com/image-cli/internal/game"
	"github.com/neko233-com/image-cli/internal/img"
	"github.com/neko233-com/image-cli/internal/spec"
)

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	var walk walkFlags
	walk.bind(fs)
	platforms := fs.String("platform", "all", "平台列表逗号分隔或 all")
	profile := fs.String("profile", "", "场景或 auto")
	domain := fs.String("domain", "", "规范: game|design|film|vfx|auto|all；空则只做平台预算")
	strict := fs.Bool("strict", false, "警告也返回退出码 1（CI 用）")
	asJSON := fs.Bool("json", false, "JSON 输出")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: image-cli check [-jobs N] [-domain game|design|film|vfx] [-profile ui|auto] <image|dir>")
	}
	return runCheckRoot(fs.Arg(0), walk, *platforms, *profile, *domain, *strict, *asJSON)
}

func runCheckRoot(root string, walk walkFlags, platforms, profile, domain string, strict, asJSON bool) error {
	plist, err := resolvePlatforms(profile, platforms)
	if err != nil {
		return err
	}
	if _, err := spec.ParseDomain(emptyDomainOK(domain)); err != nil {
		return err
	}
	st, err := os.Stat(root)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return checkMany(root, walk, plist, profile, domain, strict, asJSON)
	}
	item, err := buildCheckItem(root, plist, profile, domain)
	if err != nil {
		return err
	}
	return emitItems([]checkItem{item}, asJSON, strict, false)
}

func emptyDomainOK(domain string) string {
	if domain == "" {
		return "game"
	}
	return domain
}

func resolvePlatforms(profile, platforms string) ([]game.Platform, error) {
	if profile != "" && profile != "auto" {
		p, err := game.ParseProfile(profile)
		if err != nil {
			return nil, err
		}
		if p == game.ProfileAuto {
			return game.AllPlatforms(), nil
		}
		return game.PlatformsForProfile(p), nil
	}
	return game.ParsePlatforms(platforms)
}

type checkItem struct {
	Path    string             `json:"path"`
	Overall game.Severity      `json:"overall"`
	Summary string             `json:"summary"`
	Budget  *game.BudgetReport `json:"budget,omitempty"`
	Spec    *spec.Report       `json:"spec,omitempty"`
}

func buildCheckItem(path string, plist []game.Platform, profile, domain string) (checkItem, error) {
	info, err := img.Inspect(path, false)
	if err != nil {
		return checkItem{}, err
	}
	prof := profileName(profile, path)
	domains, err := spec.DomainsFor(domain, path, prof)
	if err != nil {
		return checkItem{}, err
	}
	item := checkItem{Path: path, Overall: game.SeverityOK}
	if wantBudget(domain, domains) {
		report := game.EvaluateBudget(game.BudgetInput{
			Width: info.Width, Height: info.Height, FileBytes: info.SizeBytes, NPOT: info.NPOT,
			Usage: game.Profile(prof),
		}, plist)
		item.Budget = &report
		item.Overall = report.Overall
		item.Summary = report.Summary
	}
	if len(domains) > 0 {
		rep := spec.Evaluate(domains, spec.Input{
			Path: path, Format: info.Format, Width: info.Width, Height: info.Height,
			Bytes: info.SizeBytes, HasAlpha: info.HasAlpha, Bits: info.Bits, NPOT: info.NPOT, Profile: prof,
		})
		item.Spec = &rep
		if severityRank(game.Severity(rep.Overall)) > severityRank(item.Overall) {
			item.Overall = game.Severity(rep.Overall)
		}
		if item.Summary == "" {
			item.Summary = rep.Summary
		}
	}
	return item, nil
}

func wantBudget(domain string, domains []spec.Domain) bool {
	switch domain {
	case "", "game", "all":
		return true
	case "auto":
		return len(domains) == 1 && domains[0] == spec.DomainGame
	default:
		return false
	}
}

func profileName(flag, path string) string {
	if flag == "" {
		if p, ok := game.DetectProfile(path); ok {
			return string(p)
		}
		return ""
	}
	p, err := game.ParseProfile(flag)
	if err != nil {
		return ""
	}
	if p == game.ProfileAuto {
		if d, ok := game.DetectProfile(path); ok {
			return string(d)
		}
		return ""
	}
	return string(p)
}

func checkMany(root string, walk walkFlags, plist []game.Platform, profile, domain string, strict, asJSON bool) error {
	paths, err := walk.paths(root)
	if err != nil {
		return err
	}
	items := make([]checkItem, len(paths))
	err = walk.parallel(len(paths), func(i int) error {
		item, err := buildCheckItem(paths[i], plist, profile, domain)
		if err != nil {
			return fmt.Errorf("%s: %w", paths[i], err)
		}
		items[i] = item
		return nil
	})
	if err != nil && !walk.keepGoing {
		return err
	}
	emitErr := emitItems(items, asJSON, strict, true)
	if err != nil {
		return err
	}
	return emitErr
}

func emitItems(items []checkItem, asJSON, strict, multi bool) error {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		var payload any = items
		if !multi && len(items) == 1 && items[0].Spec == nil {
			payload = *items[0].Budget
		} else if !multi && len(items) == 1 {
			payload = items[0]
		}
		if err := enc.Encode(payload); err != nil {
			return err
		}
	} else {
		for _, item := range items {
			if item.Path == "" && item.Summary == "" {
				continue
			}
			if multi {
				fmt.Printf("== %s ==\n", item.Path)
			}
			if item.Budget != nil {
				printBudgetReport(*item.Budget)
			}
			if item.Spec != nil {
				printSpec(*item.Spec)
			}
			if multi {
				fmt.Println()
			}
		}
	}
	worst := game.SeverityOK
	for _, item := range items {
		if severityRank(item.Overall) > severityRank(worst) {
			worst = item.Overall
		}
	}
	if worst == game.SeverityFail || (strict && worst == game.SeverityWarn) {
		os.Exit(1)
	}
	return nil
}

func severityRank(s game.Severity) int {
	switch s {
	case game.SeverityFail:
		return 2
	case game.SeverityWarn:
		return 1
	default:
		return 0
	}
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

func printSpec(r spec.Report) {
	fmt.Printf("spec:       %s\n", r.Domain)
	fmt.Printf("overall:    %s — %s\n", r.Overall, r.Summary)
	for _, f := range r.Findings {
		fmt.Printf("  [%s] %s: %s\n", f.Severity, f.Code, f.Message)
	}
}
