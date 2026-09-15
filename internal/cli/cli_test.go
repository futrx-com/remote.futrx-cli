package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/futrx-com/remote-cli/internal/application"
)

func TestCreateSyntaxes(t *testing.T) {
	for _, args := range [][]string{{"create", "app", "first", "--dir", t.TempDir()}, {"app", "create", "second", "--dir", t.TempDir()}} {
		var out, stderr bytes.Buffer
		if err := Run(args, &out, &stderr, "test"); err != nil {
			t.Fatalf("Run(%v): %v (%s)", args, err, stderr.String())
		}
		if !strings.Contains(out.String(), "Created") {
			t.Fatalf("output = %q", out.String())
		}
	}
}

func TestBuildDefaultsToCurrentDirectory(t *testing.T) {
	parent := t.TempDir()
	dir, err := application.Scaffold(parent, "sample")
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if err := Run([]string{"build", ".", "-o", filepath.Join(parent, "sample.zip")}, &out, &stderr, "test"); err != nil {
		t.Fatalf("build: %v (%s)", err, stderr.String())
	}
}
