package main

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var cliBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "image-cli-e2e-*")
	if err != nil {
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	cliBin = filepath.Join(dir, "image-cli")
	if runtime.GOOS == "windows" {
		cliBin += ".exe"
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		os.Exit(1)
	}
	cmd := exec.Command("go", "build", "-o", cliBin, "./cmd/image-cli")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Stderr.Write(out)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func runCLI(t *testing.T, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(cliBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v\n%s", cliBin, args, out)
	}
	return out
}

func TestCLIVersionJSON(t *testing.T) {
	out := runCLI(t, "version", "-json")
	if !strings.Contains(string(out), "neko233-com/image-cli") {
		t.Fatalf("%s", out)
	}
	var v struct {
		Repository string `json:"repository"`
	}
	if err := json.Unmarshal(out, &v); err != nil || v.Repository == "" {
		t.Fatal(err)
	}
}

func TestCLICheckOnSmallPNG(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "tiny.png")
	writeMinimalPNG(t, png)
	out := runCLI(t, "check", "-profile", "ui", png)
	if !strings.Contains(string(out), "overall:") {
		t.Fatalf("%s", out)
	}
}

func TestCLICheckDomainFilmParallel(t *testing.T) {
	dir := t.TempDir()
	writeMinimalPNG(t, filepath.Join(dir, "a.png"))
	writeMinimalPNG(t, filepath.Join(dir, "b.png"))
	out := runCLI(t, "check", "-domain", "film", "-jobs", "2", dir)
	if !strings.Contains(string(out), "workers=2") {
		t.Fatalf("%s", out)
	}
	if !strings.Contains(string(out), "低于 HD") {
		t.Fatalf("%s", out)
	}
}

func TestCLICheckRecursiveDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMinimalPNG(t, filepath.Join(dir, "root.png"))
	writeMinimalPNG(t, filepath.Join(sub, "child.png"))
	out := runCLI(t, "check", "-profile", "ui", dir)
	if strings.Count(string(out), "overall:") != 2 {
		t.Fatalf("expected 2 reports, got:\n%s", out)
	}
}

func TestCLIBatchConvertMirrorTree(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "icons")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMinimalPNG(t, filepath.Join(sub, "a.png"))
	outDir := filepath.Join(dir, "out")
	runCLI(t, "batch", "convert", "-format", "png", "-out-dir", outDir, dir)
	want := filepath.Join(outDir, "icons", "a.png")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("missing %s", want)
	}
}

func TestCLIPrepareAutoProfiles(t *testing.T) {
	dir := t.TempDir()
	writeMinimalPNG(t, filepath.Join(dir, "hero_albedo.png"))
	writeMinimalPNG(t, filepath.Join(dir, "hero_n.png"))
	writeMinimalPNG(t, filepath.Join(dir, "fx_fire_vfx.png"))
	outDir := filepath.Join(dir, "prepared")
	out := runCLI(t, "game", "prepare", "-profile", "auto", "-fallback", "albedo", "-no-compress", "-out-dir", outDir, dir)
	for _, name := range []string{"hero_albedo", "hero_n", "fx_fire_vfx"} {
		if !strings.Contains(string(out), name) {
			t.Fatalf("missing %s in %s", name, out)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "hero_n_prepared.png")); err != nil {
		t.Fatal(err)
	}
}

func TestCLIGamePrepareUI(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "ui.png")
	writeMinimalPNG(t, png)
	out := runCLI(t, "game", "prepare", "-profile", "ui", "-no-compress", png)
	if !strings.Contains(string(out), "prepared_png") {
		t.Fatalf("%s", out)
	}
}

func writeMinimalPNG(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
}
