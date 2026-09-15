package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/futrx-com/remote-cli/internal/application"
)

const usage = `Remote application developer CLI

Usage:
  remote create app <name> [--dir PATH]
  remote app create <name> [--dir PATH]
  remote validate [PATH]
  remote build [PATH] [-o FILE]
  remote version

Create scaffolds UI, backend, infrastructure, a skill, documentation, license,
and application.json. Build validates the source and writes a reproducible ZIP;
it also generates the infrastructure payload in-memory, replacing package.sh.`

func Run(args []string, stdout, stderr io.Writer, version string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(stdout, usage)
		return nil
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Fprintln(stdout, version)
		return nil
	}
	if len(args) >= 2 && args[0] == "create" && args[1] == "app" {
		return create(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "app" && args[1] == "create" {
		return create(args[2:], stdout, stderr)
	}
	switch args[0] {
	case "validate":
		return validate(args[1:], stdout, stderr)
	case "build", "package":
		return build(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func create(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("create app", flag.ContinueOnError)
	fs.SetOutput(stderr)
	parent := fs.String("dir", ".", "parent directory")
	if err := fs.Parse(flagsFirst(args, map[string]bool{"--dir": true})); err != nil {
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

func validate(args []string, stdout, stderr io.Writer) error {
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

func build(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("o", "", "output ZIP path")
	fs.StringVar(out, "output", "", "output ZIP path")
	if err := fs.Parse(flagsFirst(args, map[string]bool{"-o": true, "--output": true})); err != nil {
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

func shellPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if strings.ContainsAny(path, " \t'\"") {
		return fmt.Sprintf("%q", path)
	}
	return path
}

// The standard flag package stops at the first positional argument. CLI users
// reasonably write both `build -o x .` and `build . -o x`, so normalize known
// value flags before parsing.
func flagsFirst(args []string, valueFlags map[string]bool) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		if valueFlags[args[i]] && i+1 < len(args) {
			flags = append(flags, args[i], args[i+1])
			i++
			continue
		}
		matched := false
		for name := range valueFlags {
			if strings.HasPrefix(args[i], name+"=") {
				flags = append(flags, args[i])
				matched = true
				break
			}
		}
		if !matched {
			positional = append(positional, args[i])
		}
	}
	return append(flags, positional...)
}
