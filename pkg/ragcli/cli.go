// Package ragcli exposes additive CitewiseRAG commands for the root CLI.
package ragcli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/rag"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

// Run executes `citewise rag` subcommands.
func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithInput(args, os.Stdin, stdout, stderr)
}

// RunWithInput executes `citewise rag` subcommands with an explicit stdin.
func RunWithInput(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printHelp(stdout)
		return 0
	}
	cmd := args[0]
	fs := flag.NewFlagSet("rag "+cmd, flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "", "RAG candidate JSON file")
	clearance := fs.String("clearance", string(access.ClearanceInternal), "caller clearance")
	tokenBudget := fs.Int("token-budget", 0, "context token budget")
	queryType := fs.String("query-type", "", "optional query type override")
	allowDegraded := fs.Bool("allow-degraded", true, "return red/yellow plans instead of failing")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(stderr, "citewise rag: --file is required")
		return 2
	}
	set, err := loadCandidateSet(*file, stdin)
	if err != nil {
		fmt.Fprintln(stderr, "citewise rag:", err)
		return 1
	}
	ctx := access.Context{Clearance: access.Clearance(*clearance)}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	switch cmd {
	case "route":
		analysis, err := rag.Analyze(set)
		if err != nil {
			fmt.Fprintln(stderr, "citewise rag:", err)
			return 1
		}
		_ = enc.Encode(rag.DefaultRouter().Route(analysis.Query, rag.Metadata(analysis)))
	case "rank":
		analysis, err := rag.Analyze(set)
		if err != nil {
			fmt.Fprintln(stderr, "citewise rag:", err)
			return 1
		}
		ranked, err := rag.DefaultRanker().Rank(ctx, analysis, *tokenBudget)
		if err != nil {
			fmt.Fprintln(stderr, "citewise rag:", err)
			return 1
		}
		_ = enc.Encode(ranked)
	case "pack":
		resp, err := rag.NewPipeline().Run(rag.Request{CandidateSet: set, Access: ctx, QueryType: packer.QueryType(*queryType), TokenBudget: *tokenBudget, AllowDegradedPlan: *allowDegraded})
		if err != nil && !*allowDegraded {
			fmt.Fprintln(stderr, "citewise rag:", err)
			return 1
		}
		_ = enc.Encode(resp.Plan)
	case "hygiene":
		analysis, err := rag.Analyze(set)
		if err != nil {
			fmt.Fprintln(stderr, "citewise rag:", err)
			return 1
		}
		_ = enc.Encode(rag.DefaultHygieneAnalyzer().Analyze(analysis, *allowDegraded))
	default:
		fmt.Fprintf(stderr, "citewise rag: unknown command %q\n\n", cmd)
		printHelp(stderr)
		return 2
	}
	return 0
}

func loadCandidateSet(path string, stdin io.Reader) (ragnode.CandidateSet, error) {
	if path == "-" {
		return ragnode.ParseCandidateSet(stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return ragnode.CandidateSet{}, err
	}
	defer f.Close()
	return ragnode.ParseCandidateSet(f)
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `CitewiseRAG commands

Usage:
  citewise rag <command> --file candidates.json [options]

Commands:
  route    route a RAG candidate set query and metadata to query type/mode hints
  rank     rank RAG candidates after access gating
  pack     assemble a RAG candidate set into a context plan
  hygiene  inspect graph hygiene and corrective signals

Options:
  --file PATH          RAG candidate JSON file, or - for stdin
  --clearance LEVEL    public, internal, confidential, or restricted (default internal)
  --token-budget N     context token budget for pack
  --query-type TYPE    optional pack query type override
  --allow-degraded     emit red/yellow plans instead of returning an error (default true)
`)
}
