package game

import "testing"

func TestEvaluateBudgetMinigameFailLargeFile(t *testing.T) {
	in := BudgetInput{Width: 512, Height: 512, FileBytes: 900 << 10, NPOT: false}
	r := EvaluateBudget(in, []Platform{PlatformMinigame})
	if r.Overall != SeverityFail {
		t.Fatalf("expected fail for large minigame asset, got %s", r.Overall)
	}
}

func TestEvaluateBudgetUI2048OK(t *testing.T) {
	in := BudgetInput{Width: 1024, Height: 512, FileBytes: 200 << 10, NPOT: false}
	r := EvaluateBudget(in, []Platform{PlatformIOS, PlatformMinigame})
	if r.Overall == SeverityFail {
		t.Fatalf("unexpected fail: %+v", r.Platforms)
	}
}

func TestEvaluateBudgetNPOTWebGLWarn(t *testing.T) {
	in := BudgetInput{Width: 1000, Height: 500, FileBytes: 100 << 10, NPOT: true}
	r := EvaluateBudget(in, []Platform{PlatformWebGL})
	found := false
	for _, p := range r.Platforms {
		if p.Platform == PlatformWebGL && p.Status == SeverityWarn {
			found = true
		}
	}
	if !found {
		t.Fatal("expected NPOT warn on webgl")
	}
}
