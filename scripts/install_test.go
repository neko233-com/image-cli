package scripts

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallScriptValidatesChecksum(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("install.sh tests run on Unix shell")
	}
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprintf("corrupt=%v", corrupt), func(t *testing.T) {
			root := t.TempDir()
			artifacts := filepath.Join(root, "artifacts")
			installed := filepath.Join(root, "installed")
			bin := filepath.Join(root, "bin")
			for _, d := range []string{artifacts, installed, bin} {
				if err := os.MkdirAll(d, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			osName := "linux"
			archName := "amd64"
			switch runtime.GOOS {
			case "darwin":
				osName = "darwin"
			case "linux":
				osName = "linux"
			default:
				t.Skip("install.sh integration on " + runtime.GOOS)
			}
			switch runtime.GOARCH {
			case "amd64":
				archName = "amd64"
			case "arm64":
				archName = "arm64"
			default:
				t.Skip("install.sh integration on " + runtime.GOARCH)
			}
			name := "image-cli-" + osName + "-" + archName
			content := []byte{0x7f, 'E', 'L', 'F'}
			if runtime.GOOS == "darwin" {
				content = []byte{0xcf, 0xfa, 0xed, 0xfe} // Mach-O magic stub
			}
			sum := sha256.Sum256(content)
			manifest := fmt.Sprintf("%x  dist/%s\n", sum, name)
			if corrupt {
				content = []byte("bad")
			}
			if err := os.WriteFile(filepath.Join(artifacts, name), content, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(artifacts, "SHA256SUMS"), []byte(manifest), 0o600); err != nil {
				t.Fatal(err)
			}
			curl := `#!/bin/bash
set -eu
url=""; target=""; metric=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -o) target="$2"; shift 2;;
    -w) metric="$2"; shift 2;;
    --range|--max-time|-fsSL|-L|-f|-s|-S) shift;;
    https:*) url="$1"; shift;;
    *) shift;;
  esac
done
if [[ -n "$metric" ]]; then
  echo "200 0.001"
  exit 0
fi
if [[ "$target" == "/dev/null" || -z "$target" ]]; then
  exit 0
fi
cp "$FIXTURE_ARTIFACTS/${url##*/}" "$target"
`
			if err := os.WriteFile(filepath.Join(bin, "curl"), []byte(curl), 0o755); err != nil {
				t.Fatal(err)
			}
			script, err := filepath.Abs("install.sh")
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", script, "0.3.1")
			cmd.Dir = filepath.Dir(script)
			cmd.Env = append(os.Environ(),
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"FIXTURE_ARTIFACTS="+artifacts,
				"IMAGE_CLI_INSTALL_DIR="+installed,
				"IMAGE_CLI_DIRECT_ONLY=1",
			)
			out, err := cmd.CombinedOutput()
			if corrupt {
				if err == nil || !strings.Contains(string(out), "checksum mismatch") {
					t.Fatalf("expected checksum failure: err=%v out=%s", err, out)
				}
				return
			}
			if err != nil {
				t.Fatalf("install failed: %v\n%s", err, out)
			}
			binName := "image-cli"
			if runtime.GOOS == "windows" {
				binName += ".exe"
			}
			if _, err := os.Stat(filepath.Join(installed, binName)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
