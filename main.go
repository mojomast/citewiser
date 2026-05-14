package main

import (
	"os"

	"github.com/mojomast/citewiseussy/pkg/citewise"
)

func main() {
	os.Exit(citewise.Run(os.Args[1:], os.Stdout, os.Stderr))
}
