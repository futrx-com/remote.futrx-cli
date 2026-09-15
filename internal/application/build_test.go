package application

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeVersion(t *testing.T) {
	for input, want := range map[string]string{"1.2.3": "1.2.3", " release 1 / beta ": "release-1---beta", "---": ""} {
		if got := safeVersion(input); got != want {
			t.Errorf("safeVersion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildDefaultOutputAndResult(t *testing.T) {
	parent := t.TempDir()
	dir, err := Scaffold(parent, "demo")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Build(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != filepath.Join(parent, "demo-0.1.0.zip") || result.ID != "demo" || result.Version != "0.1.0" || len(result.SHA256) != 64 || result.Files == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestBuildRejectsOutputInsideApplication(t *testing.T) {
	dir, err := Scaffold(t.TempDir(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Build(dir, filepath.Join(dir, "dist", "demo.zip"))
	if err == nil || !strings.Contains(err.Error(), "outside the application") {
		t.Fatalf("error = %v", err)
	}
}
