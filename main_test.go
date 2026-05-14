package main

import (
	"bytes"
	"testing"
)

func TestMainDelegatesHelp(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--help"}, &out, &errb); code != 0 {
		t.Fatalf("help returned %d: %s", code, errb.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected help output")
	}
}

func TestMainRAGCommands(t *testing.T) {
	for _, args := range [][]string{
		{"rag", "route", "--file", "testdata/api_examples/pack_request.json"},
		{"rag", "pack", "--file", "testdata/api_examples/pack_request.json", "--token-budget", "1200"},
	} {
		var out, errb bytes.Buffer
		if code := run(args, &out, &errb); code != 0 {
			t.Fatalf("run(%v) returned %d: %s", args, code, errb.String())
		}
		if out.Len() == 0 {
			t.Fatalf("run(%v) produced no output", args)
		}
	}
}

func TestMainStillDelegatesLegacyCommands(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"roles", "--file", "testdata/citewise_backlog.json"}, &out, &errb); code != 0 {
		t.Fatalf("legacy roles returned %d: %s", code, errb.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected legacy roles output")
	}
}
