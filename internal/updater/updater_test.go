package updater

import "testing"

func TestCompareVersions(t *testing.T) {
	if compareVersions("v0.1.0", "v0.1.1") >= 0 {
		t.Fatal("expected older")
	}
	if compareVersions("v1.0.0", "v0.9.9") <= 0 {
		t.Fatal("expected newer")
	}
	if compareVersions("0.2.0", "v0.2.0") != 0 {
		t.Fatal("expected equal")
	}
}
