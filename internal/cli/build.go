package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/futrx-com/remote.futrx-cli/internal/application"
)

func runBuild(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("o", "", "output ZIP path")
	fs.StringVar(out, "output", "", "output ZIP path")
	if err := fs.Parse(normalizeValueFlags(args, map[string]bool{"-o": true, "--output": true})); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return errors.New("usage: remote build [PATH] [-o FILE]")
	}
	dir := "."
	if fs.NArg() == 1 {
		dir = fs.Arg(0)
	}
	result, err := application.Build(dir, *out)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Built %s\n  application: %s@%s\n  sha256: %s\n  files: %d\n", result.Path, result.ID, result.Version, result.SHA256, result.Files)
	return nil
}
