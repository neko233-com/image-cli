package version

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildInfoURLs(t *testing.T) {
	info := BuildInfo()
	if !strings.HasPrefix(info.GitHub, "https://github.com/") {
		t.Fatal(info.GitHub)
	}
	if info.Releases != GitHubURL()+"/releases" {
		t.Fatal(info.Releases)
	}
	if info.Repository != Repo {
		t.Fatal(info.Repository)
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Print(&buf, true); err != nil {
		t.Fatal(err)
	}
	var decoded Info
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != Binary {
		t.Fatal(decoded.Name)
	}
}
