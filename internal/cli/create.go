package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/futrx-com/remote.futrx-cli/internal/application"
)

func runCreate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("create app", flag.ContinueOnError)
	fs.SetOutput(stderr)
	parent := fs.String("dir", ".", "parent directory")
	if err := fs.Parse(normalizeValueFlags(args, map[string]bool{"--dir": true})); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: remote create app <name> [--dir PATH]")
	}
	dir, err := application.Scaffold(*parent, fs.Arg(0))
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Created %s\n\nNext:\n  cd %s\n  remote validate\n  remote build\n", dir, shellPath(dir))
	return nil
}

func shellPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if strings.ContainsAny(path, " \t'\"") {
		return fmt.Sprintf("%q", path)
	}
	return path
}
