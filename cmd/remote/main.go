package main

import (
	"fmt"
	"os"

	"github.com/futrx-com/remote-cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout, os.Stderr, version); err != nil {
		fmt.Fprintln(os.Stderr, "remote:", err)
		os.Exit(1)
	}
}
