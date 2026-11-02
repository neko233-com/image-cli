package game

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func resolveTool(envKey string, names ...string) (string, error) {
	if p := strings.TrimSpace(os.Getenv(envKey)); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("%s=%q not found: %w", envKey, p, err)
		}
		return p, nil
	}
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("tool not found in PATH (set %s): tried %s", envKey, strings.Join(names, ", "))
}

func astcencPath() (string, error) {
	names := []string{"astcenc", "astcenc-avx2", "astcenc-sse4.1", "astcenc-neon"}
	if runtime.GOOS == "windows" {
		names = append([]string{"astcenc.exe", "astcenc-avx2.exe"}, names...)
	}
	return resolveTool("IMAGE_CLI_ASTCENC", names...)
}

func texconvPath() (string, error) {
	names := []string{"texconv", "texconv.exe"}
	return resolveTool("IMAGE_CLI_TEXCONV", names...)
}

func runCmd(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %s: %w\n%s", filepath.Base(name), strings.Join(args, " "), err, out)
	}
	return string(out), nil
}
