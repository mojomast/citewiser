package main

import (
	"io"
	"os"

	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/ragcli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "rag" {
		return ragcli.Run(args[1:], stdout, stderr)
	}
	return citewise.Run(args, stdout, stderr)
}
