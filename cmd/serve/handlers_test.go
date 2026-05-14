package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/citewise"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/rag"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
	"github.com/mojomast/citewiseussy/pkg/router"
)

func TestHealthAndRouterHandlers(t *testing.T) {
	h := newServer(rag.NewPipeline())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/health code = %d", rec.Code)
	}

	body := routeRequest{Query: "latest refund policy", Metadata: router.GraphMetadata{}}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, jsonReq(http.MethodPost, "/router", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("/router code = %d body=%s", rec.Code, rec.Body.String())
	}
	var got router.Recommendation
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.QueryType != packer.QueryTemporal {
		t.Fatalf("query type = %s, want Temporal", got.QueryType)
	}
}

func TestBadJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	newServer(rag.NewPipeline()).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/rank", bytes.NewBufferString("{")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad JSON code = %d, want 400", rec.Code)
	}
}

func TestRankAccessSuppression(t *testing.T) {
	req := serveRequest("q1")
	req.Nodes[0].Sensitivity = ragnode.SensitivityRestricted
	req.Access = access.Context{Clearance: access.ClearancePublic}
	rec := httptest.NewRecorder()
	newServer(rag.NewPipeline()).ServeHTTP(rec, jsonReq(http.MethodPost, "/rank", req))
	if rec.Code != http.StatusOK {
		t.Fatalf("/rank code = %d body=%s", rec.Code, rec.Body.String())
	}
	var got rankResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Ranked.Suppressed) != 1 || got.Ranked.Suppressed[0].SuppressionReason != access.ReasonAccessControl {
		t.Fatalf("suppressed = %#v", got.Ranked.Suppressed)
	}
}

func TestPackRedAgenticReturnsOK(t *testing.T) {
	req := serveRequest("q2")
	req.Query = "Can the agent approve a refund?"
	req.Nodes = append(req.Nodes, ragnode.RAGNode{Item: citewise.Item{ID: "perm", Title: "Permission", Trust: 1}, Text: "permission", ChunkType: ragnode.ChunkPermissionRecord, Sensitivity: ragnode.SensitivityInternal, ApprovedBy: []string{"ops"}})
	req.Candidates = append(req.Candidates, ragnode.Candidate{NodeID: "perm", QueryRelevance: 1})
	req.Edges = append(req.Edges, ragnode.Edge{SourceID: "perm", TargetID: "n1", Type: ragnode.EdgeDecisionBasis, Confidence: 1})
	req.Access = access.Context{Clearance: access.ClearanceInternal, TrustedApprovers: []string{"ops"}}
	rec := httptest.NewRecorder()
	newServer(rag.NewPipeline()).ServeHTTP(rec, jsonReq(http.MethodPost, "/pack", req))
	if rec.Code != http.StatusOK {
		t.Fatalf("/pack code = %d body=%s", rec.Code, rec.Body.String())
	}
	var got packResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Plan.HygieneSignal != packer.HygieneRed {
		t.Fatalf("hygiene signal = %s, want red", got.Plan.HygieneSignal)
	}
}

func TestHygieneHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	newServer(rag.NewPipeline()).ServeHTTP(rec, jsonReq(http.MethodPost, "/hygiene", serveRequest("q3")))
	if rec.Code != http.StatusOK {
		t.Fatalf("/hygiene code = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIExampleFixtures(t *testing.T) {
	h := newServer(rag.NewPipeline())
	for _, tc := range []struct {
		path     string
		endpoint string
	}{
		{"router_request.json", "/router"},
		{"pack_request.json", "/pack"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "api_examples", tc.path))
			if err != nil {
				t.Fatal(err)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, tc.endpoint, bytes.NewReader(data)))
			if rec.Code != http.StatusOK {
				t.Fatalf("%s code = %d body=%s", tc.endpoint, rec.Code, rec.Body.String())
			}
		})
	}
}

func jsonReq(method, path string, value any) *http.Request {
	data, _ := json.Marshal(value)
	return httptest.NewRequest(method, path, bytes.NewReader(data))
}

func serveRequest(queryID string) candidateRequest {
	return candidateRequest{
		QueryID: queryID,
		Query:   "refund policy",
		Access:  access.Context{Clearance: access.ClearanceInternal},
		Nodes: []ragnode.RAGNode{
			{Item: citewise.Item{ID: "n1", Title: "Refund policy", Notes: "Policy text", Trust: 1, GoalFit: 1}, Text: "Refund policy text", ChunkType: ragnode.ChunkDocument, Sensitivity: ragnode.SensitivityInternal, Version: "v1", CommunityID: "refunds"},
		},
		Candidates: []ragnode.Candidate{{NodeID: "n1", QueryRelevance: 1}},
	}
}
