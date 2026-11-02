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
