package cli

import (
	"fmt"
	"io"
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
		return runCreate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "app" && args[1] == "create" {
		return runCreate(args[2:], stdout, stderr)
	}
	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "build", "package":
		return runBuild(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}
