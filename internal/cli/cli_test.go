package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/futrx-com/remote.futrx-cli/internal/application"
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

func TestInformationalCommands(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "no arguments", args: nil, want: usage + "\n"},
		{name: "help command", args: []string{"help"}, want: usage + "\n"},
		{name: "long help flag", args: []string{"--help"}, want: usage + "\n"},
		{name: "short help flag", args: []string{"-h"}, want: usage + "\n"},
		{name: "version command", args: []string{"version"}, want: "test-version\n"},
		{name: "version flag", args: []string{"--version"}, want: "test-version\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if err := Run(tt.args, &stdout, &stderr, "test-version"); err != nil {
				t.Fatalf("Run(%v): %v", tt.args, err)
			}
			if stdout.String() != tt.want {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tt.want)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"unknown"}, &stdout, &stderr, "test")
	want := "unknown command \"unknown\"\n\n" + usage
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("stdout = %q, stderr = %q; want both empty", stdout.String(), stderr.String())
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
