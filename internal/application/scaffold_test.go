package application

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlug(t *testing.T) {
	for input, want := range map[string]string{" Demo App ": "demo-app", "a---b": "a-b", "Hello_world!": "hello-world", "---": ""} {
		if got := slug(input); got != want {
			t.Errorf("slug(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestScaffoldCreatesExpectedFilesAndModes(t *testing.T) {
	dir, err := Scaffold(t.TempDir(), "demo app")
	if err != nil {
		t.Fatal(err)
	}
	for name := range scaffoldFiles("demo-app", "Demo app") {
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		want := os.FileMode(0644)
		if name == "infra/install.sh" {
			want = 0755
		}
		if info.Mode().Perm() != want {
			t.Errorf("%s mode = %o, want %o", name, info.Mode().Perm(), want)
		}
	}
}

func TestScaffoldRejectsInvalidOrExistingDestination(t *testing.T) {
	parent := t.TempDir()
	if _, err := Scaffold(parent, "---"); err == nil {
		t.Fatal("invalid name accepted")
	}
	if _, err := Scaffold(parent, "demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := Scaffold(parent, "demo"); err == nil {
		t.Fatal("existing destination accepted")
	}
}
