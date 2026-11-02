package version

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
)

const (
	Module     = "github.com/neko233-com/image-cli"
	License    = "MIT"
	DocsPath   = "blob/main/README.md"
	GameDocPath = "blob/main/docs/GAME.md"
)

// Commit 由 CI -ldflags 注入（短 SHA）。
var Commit = ""

type Info struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Module      string `json:"module"`
	Repository  string `json:"repository"`
	GitHub      string `json:"github"`
	Releases    string `json:"releases"`
	Issues      string `json:"issues"`
	RawInstall  string `json:"install_sh_raw"`
	License     string `json:"license"`
	Commit      string `json:"commit,omitempty"`
	GoVersion   string `json:"go_version"`
	Platform    string `json:"platform"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
}

func GitHubURL() string  { return "https://github.com/" + Repo }
func ReleasesURL() string { return GitHubURL() + "/releases" }
func IssuesURL() string  { return GitHubURL() + "/issues" }

func BuildInfo() Info {
	commit := Commit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	return Info{
		Name:       Binary,
		Version:    Version,
		Module:     Module,
		Repository: Repo,
		GitHub:     GitHubURL(),
		Releases:   ReleasesURL(),
		Issues:     IssuesURL(),
		RawInstall: "https://raw.githubusercontent.com/" + Repo + "/main/scripts/install.sh",
		License:    License,
		Commit:     commit,
		GoVersion:  runtime.Version(),
		Platform:   runtime.GOOS + "/" + runtime.GOARCH,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}
}

func Print(w io.Writer, asJSON bool) error {
	info := BuildInfo()
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}
	lines := []string{
		fmt.Sprintf("%s %s", info.Name, info.Version),
		fmt.Sprintf("platform:   %s", info.Platform),
		fmt.Sprintf("go:         %s", info.GoVersion),
		fmt.Sprintf("repository: %s", info.Repository),
		fmt.Sprintf("github:     %s", info.GitHub),
		fmt.Sprintf("releases:   %s", info.Releases),
		fmt.Sprintf("issues:     %s", info.Issues),
		fmt.Sprintf("install:    curl -fsSL %s | bash", info.RawInstall),
		fmt.Sprintf("license:    %s", info.License),
	}
	if info.Commit != "" {
		lines = append(lines, fmt.Sprintf("commit:     %s", info.Commit))
	}
	_, err := fmt.Fprintln(w, strings.Join(lines, "\n"))
	return err
}
