package main

import (
	"bytes"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/citewise"
)

func TestMainDelegatesHelp(t *testing.T) {
	var out, errb bytes.Buffer
	if code := citewise.Run([]string{"--help"}, &out, &errb); code != 0 {
		t.Fatalf("help returned %d: %s", code, errb.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected help output")
	}
}
