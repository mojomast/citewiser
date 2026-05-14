package main

import (
	"io"
	"os"

	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragcli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "rag" {
		return ragcli.RunWithInput(args[1:], os.Stdin, stdout, stderr)
	}
	return citewise.Run(args, stdout, stderr)
}
