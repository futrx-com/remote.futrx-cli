package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/futrx-com/remote.futrx-cli/internal/application"
)

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := Run(args, &stdout, &stderr, "test-version")
	return stdout.String(), stderr.String(), err
}

func TestCreateCommandOutputAndFlagPositions(t *testing.T) {
	for _, tt := range []struct {
		name string
		args func(string) []string
	}{
		{"canonical flags first", func(parent string) []string { return []string{"create", "app", "--dir", parent, "demo"} }},
		{"canonical flags last", func(parent string) []string { return []string{"create", "app", "demo", "--dir", parent} }},
		{"alias", func(parent string) []string { return []string{"app", "create", "demo", "--dir=" + parent} }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			stdout, stderr, err := runCLI(t, tt.args(parent)...)
			if err != nil {
				t.Fatalf("error = %v, stderr = %q", err, stderr)
			}
			dir := filepath.Join(parent, "demo")
			want := "Created " + dir + "\n\nNext:\n  cd " + dir + "\n  remote validate\n  remote build\n"
			if stdout != want || stderr != "" {
				t.Fatalf("stdout = %q, stderr = %q", stdout, stderr)
			}
		})
	}
}

func TestValidateCommandSuccessAndFailures(t *testing.T) {
	dir, err := application.Scaffold(t.TempDir(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := runCLI(t, "validate", dir)
	if err != nil || stdout != "Valid Remote application Demo (0.1.0)\n" || stderr != "" {
		t.Fatalf("stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	for _, args := range [][]string{{"validate", "one", "two"}, {"validate", filepath.Join(dir, "missing")}} {
		stdout, _, err = runCLI(t, args...)
		if err == nil || stdout != "" {
			t.Fatalf("Run(%v): stdout=%q err=%v", args, stdout, err)
		}
	}
}

func TestBuildCommandAndPackageAlias(t *testing.T) {
	for _, command := range []string{"build", "package"} {
		t.Run(command, func(t *testing.T) {
			parent := t.TempDir()
			dir, err := application.Scaffold(parent, "demo")
			if err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(parent, command+".zip")
			stdout, stderr, err := runCLI(t, command, dir, "--output="+output)
			if err != nil || stderr != "" {
				t.Fatalf("stdout=%q stderr=%q err=%v", stdout, stderr, err)
			}
			for _, fragment := range []string{"Built " + output, "application: demo@0.1.0", "sha256:", "files:"} {
				if !strings.Contains(stdout, fragment) {
					t.Errorf("stdout %q missing %q", stdout, fragment)
				}
			}
			if _, err := os.Stat(output); err != nil {
				t.Fatalf("artifact: %v", err)
			}
		})
	}
}

func TestCommandUsageErrors(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"create", "app"}, "usage: remote create app <name> [--dir PATH]"},
		{[]string{"create", "app", "one", "two"}, "usage: remote create app <name> [--dir PATH]"},
		{[]string{"validate", "one", "two"}, "usage: remote validate [PATH]"},
		{[]string{"build", "one", "two"}, "usage: remote build [PATH] [-o FILE]"},
	} {
		stdout, _, err := runCLI(t, tt.args...)
		if err == nil || err.Error() != tt.want {
			t.Fatalf("Run(%v): error = %v", tt.args, err)
		}
		if stdout != "" {
			t.Fatalf("Run(%v): stdout = %q", tt.args, stdout)
		}
	}
}

func TestFlagParsingErrorsUseStderr(t *testing.T) {
	stdout, stderr, err := runCLI(t, "build", "--unknown")
	if err == nil || stdout != "" || !strings.Contains(stderr, "flag provided but not defined") {
		t.Fatalf("stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
}

func TestNormalizeValueFlags(t *testing.T) {
	got := normalizeValueFlags([]string{"app", "--output", "one.zip", "--other", "--output=two.zip"}, map[string]bool{"--output": true})
	want := []string{"--output", "one.zip", "--output=two.zip", "app", "--other"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestShellPathQuotesWhitespace(t *testing.T) {
	got := shellPath(filepath.Join(t.TempDir(), "space here"))
	if !strings.HasPrefix(got, `"`) || !strings.HasSuffix(got, `"`) {
		t.Fatalf("shellPath = %q", got)
	}
}
