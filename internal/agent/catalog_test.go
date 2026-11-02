package agent

import "testing"

func TestPromptCoversProductionPaths(t *testing.T) {
	text := Prompt()
	for _, want := range []string{"不要编造", "色度键", "-profile auto", "atlas pack", "-jobs 0"} {
		if !contains(text, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	ids := map[string]bool{}
	for _, task := range Tasks() {
		if task.ID == "" || task.Command == "" || len(task.Argv) == 0 {
			t.Fatalf("incomplete task %+v", task)
		}
		if ids[task.ID] {
			t.Fatalf("duplicate %s", task.ID)
		}
		ids[task.ID] = true
	}
	for _, id := range []string{"game-ui", "game-textures", "game-atlas-unity", "film-key", "film-gif-key", "sticker"} {
		if !ids[id] {
			t.Fatalf("missing task %s", id)
		}
	}
}

func contains(s, part string) bool {
	return len(s) >= len(part) && (s == part || len(part) == 0 || indexOf(s, part) >= 0)
}

func indexOf(s, part string) int {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
