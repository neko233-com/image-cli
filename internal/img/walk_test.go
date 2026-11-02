package img

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEachRasterSkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "a.png"), 4, 4)
	gitDir := filepath.Join(dir, ".git", "objects")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writePNG(t, filepath.Join(gitDir, "hidden.png"), 4, 4)
	count := 0
	err := EachRasterWithOptions(dir, WalkOptions{Recursive: true}, func(string) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 file, got %d", count)
	}
}

func TestResolveOutputPathMirror(t *testing.T) {
	root := filepath.Join("Assets", "UI")
	input := filepath.Join("Assets", "UI", "icons", "a.png")
	out, err := ResolveOutputPath(root, input, "out", true, "webp")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(filepath.Dir(out)) != "icons" {
		t.Fatalf("expected mirrored icons dir, got %s", out)
	}
}
