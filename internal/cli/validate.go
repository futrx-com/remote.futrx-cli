package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/futrx-com/remote.futrx-cli/internal/application"
)

func runValidate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return errors.New("usage: remote validate [PATH]")
	}
	dir := "."
	if fs.NArg() == 1 {
		dir = fs.Arg(0)
	}
	manifest, err := application.Validate(dir)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Valid Remote application %s (%s)\n", manifest.Name, manifest.Version)
	return nil
}
