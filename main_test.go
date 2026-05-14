package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/ragcli"
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
		{"rag", "rank", "--file", "testdata/api_examples/pack_request.json"},
		{"rag", "pack", "--file", "testdata/api_examples/pack_request.json", "--token-budget", "1200"},
		{"rag", "hygiene", "--file", "testdata/api_examples/pack_request.json"},
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

func TestRAGCLIReadsCandidateSetFromStdin(t *testing.T) {
	data, err := os.ReadFile("testdata/api_examples/pack_request.json")
	if err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := ragcli.RunWithInput([]string{"route", "--file", "-"}, bytes.NewReader(data), &out, &errb); code != 0 {
		t.Fatalf("rag route stdin returned %d: %s", code, errb.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected rag route stdin output")
	}
}

func TestRAGCLIGoldenStdout(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"route", []string{"rag", "route", "--file", "testdata/api_examples/pack_request.json"}},
		{"rank", []string{"rag", "rank", "--file", "testdata/api_examples/pack_request.json"}},
		{"pack", []string{"rag", "pack", "--file", "testdata/api_examples/pack_request.json", "--token-budget", "1200"}},
		{"hygiene", []string{"rag", "hygiene", "--file", "testdata/api_examples/pack_request.json"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errb bytes.Buffer
			if code := run(tc.args, &out, &errb); code != 0 {
				t.Fatalf("run(%v) returned %d: %s", tc.args, code, errb.String())
			}
			want, err := os.ReadFile(filepath.Join("testdata", "rag_cli_golden", tc.name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			if out.String() != string(want) {
				t.Fatalf("golden mismatch for %s\n got: %s\nwant: %s", tc.name, out.String(), want)
			}
		})
	}
}
