package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestReleaseDependencyGuardrails(t *testing.T) {
	normal := goListDeps(t, "./...")
	forbidden := []string{
		"github.com/redis/go-redis",
		"github.com/neo4j/neo4j-go-driver",
		"github.com/blevesearch/bleve",
		"gonum.org/v1/gonum",
		"github.com/apache/arrow-go",
		"github.com/sashabaranov/go-openai",
		"github.com/anthropics/anthropic-sdk-go",
		"github.com/parquet-go/parquet-go",
	}
	for _, dep := range forbidden {
		if containsDep(normal, dep) {
			t.Fatalf("non-tagged dependency list unexpectedly contains %s", dep)
		}
	}

	parquet := goListDeps(t, "-tags", "graphrag_parquet", "./pkg/integrations/graphrag")
	if !containsDep(parquet, "github.com/parquet-go/parquet-go") {
		t.Fatal("parquet dependency missing when graphrag_parquet tag is enabled")
	}
}

func goListDeps(t *testing.T, args ...string) []string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"list", "-deps"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	return lines
}

func containsDep(deps []string, prefix string) bool {
	for _, dep := range deps {
		if dep == prefix || strings.HasPrefix(dep, prefix+"/") {
			return true
		}
	}
	return false
}
