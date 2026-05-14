package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/rag"
	"github.com/mojomast/citewiseussy/pkg/router"
)

func TestRunStdioRouterGolden(t *testing.T) {
	input := `{"operation":"router","request":{"query":"latest refund policy","metadata":{}}}`
	var out bytes.Buffer
	if err := runStdio(strings.NewReader(input), &out, rag.NewPipeline()); err != nil {
		t.Fatal(err)
	}
	var resp struct {
		OK       bool                  `json:"ok"`
		Response router.Recommendation `json:"response"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK || resp.Response.QueryType != packer.QueryTemporal {
		t.Fatalf("stdio router response = %s", out.String())
	}
}

func TestRunStdioPackGolden(t *testing.T) {
	req := stdioRequest{Operation: "pack"}
	data, err := json.Marshal(serveRequest("q-stdio"))
	if err != nil {
		t.Fatal(err)
	}
	req.Request = data
	input, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runStdio(bytes.NewReader(input), &out, rag.NewPipeline()); err != nil {
		t.Fatal(err)
	}
	var resp struct {
		OK       bool         `json:"ok"`
		Response packResponse `json:"response"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK || resp.Response.Plan.QueryID != "q-stdio" || len(resp.Response.Plan.Slots) == 0 {
		t.Fatalf("stdio pack response = %s", out.String())
	}
}
