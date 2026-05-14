CitewiseRAG Technical Specification
1. Executive Summary

The problem CitewiseRAG solves is the assembly problem, not the retrieval problem: before an AI agent answers or acts, it needs the current authoritative record, controlling policy, user permissions, document section, metric source, prior decision, and a source trail that lets a human reviewer reconstruct why the agent acted. Vector search can find semantically related chunks, but it does not assemble a role-aware, access-controlled, provenance-tagged context package. CitewiseRAG is therefore specified as a deterministic Go knowledge layer that sits after upstream retrieval and prepares the context an agent needs before execution. This framing and the required package list come from the provided build brief.

CitewiseRAG extends the existing github.com/mojomast/citewiseussy project instead of replacing it. The current code already has the core primitive this system needs: item/edge modeling, role classification, centrality, duplicate detection, hygiene reporting, prerequisite-aware readiness, and explainable scoring. The existing engine defines roles such as foundation, overview, bridge, counterpoint, stale-hype, duplicate, and curiosity-leaf, and it already scores items using goal fit, centrality, readiness, freshness, redundancy, and energy/time fit. The existing item and edge structs already provide a compact graph model with item metadata, trust, topics, normalized edge types, and confidence.

The new system keeps the old CLI behavior intact while adding a GraphRAG-oriented library layer: candidate intake, access gating, role classification, deterministic ranking, hygiene/corrective signals, context packing, provenance trails, HTTP/stdio integration, and memory write-back. It deliberately does not own embeddings, vector databases, document chunking, LLM extraction, LLM reranking, or graph community detection. Those belong upstream. CitewiseRAG owns the final, auditable act of assembling the right context in the right order.

2. Architecture Overview

Microsoft GraphRAG is the closest upstream analogue: it extracts entities, relationships, and claims from text units; builds a hierarchical community graph; generates community summaries; and supports global, local, DRIFT, and basic search modes. CitewiseRAG does not duplicate that indexing work. It consumes its outputs, especially community summaries, text units, entities, relationships, and claims, then uses Citewise-style roles to assemble a context plan.

State-of-the-Art Technique Mapping
Technique	What it is	Citewise mapping	Decision	Package/component touched
Microsoft GraphRAG	Entity/relationship/claim extraction, hierarchical community summaries, global/local/DRIFT/basic search.	community-summary nodes become overview; local entity neighborhoods feed foundation, bridge, and counterpoint; DRIFT follow-ups map to prerequisite/bridge expansion.	Adopt as upstream index format; do not reimplement indexing.	pkg/integrations/graphrag, pkg/router, pkg/packer.
HippoRAG 2	Knowledge graph plus Personalized PageRank for multi-hop retrieval and memory-like association. HippoRAG 2 builds on PPR and deeper passage integration.	PPR becomes query-personalized graph importance; existing Citewise weighted centrality remains global graph importance.	Adapt. PPR augments, not replaces, Citewise centrality.	pkg/ranker/ppr.go.
LightRAG	Dual-level retrieval using low-level entity neighborhoods and high-level knowledge discovery/community retrieval; combines graph structures with vector representations.	Low-level local retrieval maps to foundation/procedure; high-level retrieval maps to overview; cross-community nodes map to bridge.	Adopt as upstream candidate source.	pkg/integrations/lightrag, pkg/router.
RAPTOR	Recursively embeds, clusters, and summarizes chunks into a tree for multi-level retrieval.	Summary-tree nodes become ChunkType=community-summary; lower chunks become chunk/section; tree parent summaries become overview or foundation depending on authority.	Adapt as accepted graph node type, not generated internally.	pkg/ragnode, pkg/packer.
DRIFT search	Combines global community context with local follow-up exploration and iterative query refinement.	prerequisite, bridge, and community-member-of edges trigger follow-up retrieval requests when required slots are empty.	Adapt as corrective signal, not internal LLM loop.	pkg/hygiene, pkg/router, pkg/packer.
Contextual Retrieval	Prepends chunk-specific situational context before embedding/BM25 indexing; Anthropic reports large retrieval-failure reductions when combined with BM25 and reranking.	The prepended text is stored in RAGNode.ContextPrefix; CitewiseRAG does not generate it.	Adopt as pre-ingestion metadata.	pkg/ragnode.
Hybrid retrieval with RRF	Fuses ranked result lists from BM25, dense, and graph methods using reciprocal rank; Azure documents the 1/(rank+k) formula and notes k=60 as a common effective value.	Fused score is injected as Candidate.QueryRelevance.	Adopt upstream only.	pkg/ragnode.Candidate, pkg/integrations/hybrid.
Cross-encoder / ColBERT-style reranking	Neural reranking or late-interaction scoring for fine-grained query-passage relevance; ColBERT introduced efficient late interaction, and ColBERTv2 improves quality/space via compression.	Reranker score becomes part of QueryRelevance, but the critic still enforces roles, ACL, diversity, and provenance.	Run before CitewiseRAG, after coarse ACL.	Upstream handoff schema, pkg/ranker.
Lost in the Middle mitigation	LLMs often use evidence better at the beginning or end of context than in the middle.	foundation, permission-record, and counterpoint slots are placed at context edges; supporting details go in the middle.	Adopt.	pkg/packer/ordering.go.
Token budget optimization	Context windows are finite; long low-density nodes displace better evidence.	Replaces EnergyTimeFit with TokenBudgetFit.	Adopt.	pkg/ranker/token_budget.go, pkg/packer.
Self-RAG / CRAG	Self-RAG adaptively retrieves and critiques retrieved passages; CRAG evaluates retrieval quality and triggers corrective actions.	HygieneSignal tells the caller whether to proceed, re-retrieve, or route to DRIFT-style expansion.	Adapt deterministically; no internal LLM reflection.	pkg/hygiene, pkg/packer.
Agent memory write-back	Persist successful context assemblies so agents do not rebuild the same context repeatedly.	MemoryWriteBack stores ContextPlan and evidence paths.	Adopt as interface plus file-backed implementation.	pkg/memory.
Access-control-aware retrieval	Permissions are enforced before scoring and packing.	Sensitivity, ApprovedBy, and caller clearance are hard gates.	Adopt as non-negotiable invariant.	pkg/access, pkg/packer, pkg/memory.
3. Package Specifications
3.1 Module Structure
citewiseussy/
  go.mod
  main.go                                  # Existing CLI entrypoint; unchanged.
  cmd/
    citewise/
      main.go                              # Optional relocation of existing CLI entrypoint.
    serve/
      main.go                              # Optional HTTP JSON API server.
      handlers.go                          # /rank, /pack, /hygiene, /router, /health handlers.
      schema.go                            # HTTP request/response DTOs.
  pkg/
    citewise/
      cli.go                               # Existing commands: roles, score, queue, explain, hygiene, export.
      engine.go                            # Existing role classifier, centrality, score, hygiene.
      types.go                             # Existing Item, Edge, Goal, Backlog parsing.
      engine_test.go                       # Existing tests.
    ragnode/
      node.go                              # RAGNode, ChunkType, Sensitivity, source location.
      edge.go                              # Extended edge ontology and normalization.
      candidate.go                         # Upstream candidate and score handoff types.
      analysis.go                          # RAGAnalysis graph container.
      convert.go                           # Conversion to/from citewise.Item and citewise.Edge.
    access/
      access.go                            # Clearance hierarchy and AccessController interface.
      approved.go                          # ApprovedBy trust rules.
    ranker/
      ranker.go                            # Ranker interfaces and score structs.
      scorer.go                            # Deterministic scoring formula.
      authority.go                         # AuthorityScore calculation.
      ppr.go                               # Personalized PageRank implementation.
      token_budget.go                      # TokenBudgetFit and density calculation.
      diversity.go                         # Source/community diversity scoring.
      explain.go                           # Human-readable inclusion/exclusion rationales.
    packer/
      querytype.go                         # QueryType enum.
      plan.go                              # ContextPlan, ContextSlot, SuppressedEntry.
      slots.go                             # Slot policies per QueryType.
      ordering.go                          # Lost-in-the-Middle positioning.
      packer.go                            # Pack() implementation.
    hygiene/
      hygiene.go                           # Wrapper around existing HygieneReport.
      suggestions.go                       # SuggestMissingEdges.
      signal.go                            # HygieneScore and CorrectiveSignal.
    router/
      router.go                            # QueryRouter interface.
      heuristics.go                        # Deterministic decision tree.
      metadata.go                          # GraphMetadata.
    memory/
      interface.go                         # MemoryWriteBack interface.
      file_store.go                        # JSON file-backed default memory store.
      redact.go                            # Re-gating prior plans on load.
    provenance/
      trail.go                             # SourceTrail and SourceRef builders.
      redact.go                            # Unauthorized provenance redaction.
    integrations/
      graphrag/
        mapper.go                          # Microsoft GraphRAG table-to-node mapping.
        parquet.go                         # Optional parquet reader, build tag: graphrag_parquet.
      lightrag/
        mapper.go                          # LightRAG local/global candidate mapping.
      hybrid/
        schema.go                          # BM25/dense/graph/RRF handoff schema.
    rag/
      pipeline.go                          # Orchestrates access -> classify -> rank -> pack.
      interfaces.go                        # Top-level mockable interfaces.
  testdata/
    graphrag_minimal/
    lightrag_minimal/
    hybrid_rrf/

The existing pkg/citewise package remains the compatibility anchor. Its CLI already exposes roles, score, queue, explain, hygiene, and export, and those commands must continue to work unchanged.

3.2 pkg/ragnode

pkg/ragnode extends the existing citewise.Item model into a production GraphRAG node. It must embed, not fork, the old type.

package ragnode

import (
	"time"

	"github.com/mojomast/citewiseussy/pkg/citewise"
)

type ChunkType string

const (
	ChunkDocument         ChunkType = "document"
	ChunkSection          ChunkType = "section"
	ChunkChunk            ChunkType = "chunk"
	ChunkEntity           ChunkType = "entity"
	ChunkClaim            ChunkType = "claim"
	ChunkCommunitySummary ChunkType = "community-summary"
	ChunkProcedure        ChunkType = "procedure"
	ChunkDefinition       ChunkType = "definition"
	ChunkDecision         ChunkType = "decision"
	ChunkPermissionRecord ChunkType = "permission-record"
)

type Sensitivity string

const (
	SensitivityPublic       Sensitivity = "public"
	SensitivityInternal     Sensitivity = "internal"
	SensitivityConfidential Sensitivity = "confidential"
	SensitivityRestricted   Sensitivity = "restricted"
)

type SourceHop struct {
	NodeID     string  `json:"node_id"`
	EdgeType   string  `json:"edge_type"`
	Confidence float64 `json:"confidence"`
}

type SourceRef struct {
	NodeID       string    `json:"node_id"`
	Origin       string    `json:"origin"`        // e.g. "graphrag", "lightrag", "notion", "gdrive", "postgres"
	Source       string    `json:"source"`        // inherited from citewise.Item.Source
	URL          string    `json:"url,omitempty"`  // inherited from citewise.Item.URL
	Version      string    `json:"version"`
	ObservedAt  time.Time `json:"observed_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Locator     Locator   `json:"locator"`
	CommunityID string    `json:"community_id,omitempty"`
}

type Locator struct {
	DocumentID  string `json:"document_id,omitempty"`
	SectionID   string `json:"section_id,omitempty"`
	SectionPath string `json:"section_path,omitempty"`
	PageStart   int    `json:"page_start,omitempty"`
	PageEnd     int    `json:"page_end,omitempty"`
	TableID     string `json:"table_id,omitempty"`
	RowStart    int    `json:"row_start,omitempty"`
	RowEnd      int    `json:"row_end,omitempty"`
	CharStart   int    `json:"char_start,omitempty"`
	CharEnd     int    `json:"char_end,omitempty"`
}

type RAGNode struct {
	citewise.Item

	Text           string      `json:"text"`
	ChunkType      ChunkType   `json:"chunk_type"`
	TokenCount     int         `json:"token_count"`
	Version        string      `json:"version"`
	Sensitivity    Sensitivity `json:"sensitivity"`
	EmbeddingModel string      `json:"embedding_model,omitempty"`
	ApprovedBy     []string    `json:"approved_by,omitempty"`
	SupersededBy    string      `json:"superseded_by,omitempty"`
	ContextPrefix   string      `json:"context_prefix,omitempty"`
	CommunityID     string      `json:"community_id,omitempty"`
	SourceTrail     []SourceHop `json:"source_trail,omitempty"`

	Origin       string            `json:"origin,omitempty"`
	ObservedAt   time.Time         `json:"observed_at,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at,omitempty"`
	Locator      Locator           `json:"locator,omitempty"`
	SemanticType string            `json:"semantic_type,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type Edge struct {
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence,omitempty"`
	Note       string  `json:"note,omitempty"`

	Origin     string    `json:"origin,omitempty"`
	Version    string    `json:"version,omitempty"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
}

type MethodScore struct {
	Method string  `json:"method"`          // bm25, dense, graph, lightrag-local, graphrag-global, cross-encoder
	Rank   int     `json:"rank"`            // 1-based rank from that method
	Score  float64 `json:"score"`           // raw upstream score
	Weight float64 `json:"weight,omitempty"`// optional RRF/source weight
}

type Candidate struct {
	NodeID          string        `json:"node_id"`
	QueryRelevance float64       `json:"query_relevance"` // normalized 0..1; usually RRF/reranker fused upstream
	MethodScores   []MethodScore `json:"method_scores,omitempty"`
	RerankerScore  float64       `json:"reranker_score,omitempty"`
	Rank           int           `json:"rank,omitempty"`
	RetrievalMode  string        `json:"retrieval_mode,omitempty"`
	SourceTrail    []SourceHop   `json:"source_trail,omitempty"`
}

type CandidateSet struct {
	QueryID    string      `json:"query_id"`
	Query     string      `json:"query"`
	Nodes     []RAGNode   `json:"nodes"`
	Edges     []Edge      `json:"edges"`
	Candidates []Candidate `json:"candidates"`
	GeneratedAt time.Time `json:"generated_at"`
}

type RAGAnalysis struct {
	QueryID     string
	Query      string
	Nodes      map[string]RAGNode
	Edges      []Edge
	EdgesOut   map[string][]Edge
	EdgesIn    map[string][]Edge
	Candidates map[string]Candidate

	Roles      map[string]string
	Centrality map[string]float64
	PPR        map[string]float64
	Duplicates [][]string

	BaseAnalysis citewise.Analysis
	Now          time.Time
}
Required behavior

RAGNode.ToItem() must produce a valid citewise.Item so existing classification and hygiene logic can be reused. Edge.ToCitewiseEdge() must normalize edge types using the extended ontology and preserve Confidence. If TokenCount == 0, the ranker must estimate tokens as ceil(len(Text)/4) and mark the estimate in the rationale.

3.3 pkg/access

Access control is a hard filter before scoring. It is not a ranking feature.

package access

import "github.com/mojomast/citewiseussy/pkg/ragnode"

type Clearance string

const (
	ClearancePublic       Clearance = "public"
	ClearanceInternal     Clearance = "internal"
	ClearanceConfidential Clearance = "confidential"
	ClearanceRestricted   Clearance = "restricted"
)

type Context struct {
	CallerID         string             `json:"caller_id"`
	Groups           []string           `json:"groups,omitempty"`
	Clearance        Clearance          `json:"clearance"`
	TrustedApprovers []string          `json:"trusted_approvers,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"` // allowed | access-control
	Detail  string `json:"detail,omitempty"`
}

type Controller interface {
	CanSeeNode(ctx Context, node ragnode.RAGNode) Decision
	CanUseEdge(ctx Context, edge ragnode.Edge) Decision
	RedactNode(ctx Context, node ragnode.RAGNode) ragnode.RAGNode
}

Rules:

Convert Sensitivity and Clearance to ordinal values: public=0, internal=1, confidential=2, restricted=3.
If node.Sensitivity > ctx.Clearance, suppress with reason access-control.
If node.Sensitivity >= confidential and node.ApprovedBy is non-empty, at least one ApprovedBy principal must be in ctx.TrustedApprovers; otherwise suppress with reason access-control.
For ChunkType=permission-record, decision, and procedure in Agentic mode, ApprovedBy must be non-empty unless ctx.Attributes["allow_unapproved_agentic_nodes"] == "true".
Unauthorized suppressed entries may expose only node_id, reason, and detail; never Text, Title, URL, or SourceTrail.
3.4 pkg/ranker

pkg/ranker implements the retrieval critic. It receives a candidate set, applies access gating, calculates role-aware scores, and returns ranked candidates plus suppression reasons.

package ranker

import (
	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

type Score struct {
	NodeID             string   `json:"node_id"`
	Total             float64  `json:"total"`
	QueryRelevance    float64  `json:"query_relevance"`
	GraphImportance   float64  `json:"graph_importance"`
	AuthorityScore    float64  `json:"authority_score"`
	Readiness          float64  `json:"readiness"`
	Freshness          float64  `json:"freshness"`
	TokenBudgetFit     float64  `json:"token_budget_fit"`
	BridgeBonus        float64  `json:"bridge_bonus"`
	CounterpointBonus  float64  `json:"counterpoint_bonus"`
	DiversityBonus     float64  `json:"diversity_bonus"`
	RedundancyPenalty  float64  `json:"redundancy_penalty"`
	StalenessPenalty   float64  `json:"staleness_penalty"`
	LowTrustPenalty    float64  `json:"low_trust_penalty"`
	Role               string   `json:"role"`
	AccessAllowed      bool     `json:"access_allowed"`
	SuppressionReason  string   `json:"suppression_reason,omitempty"`
	Rationale          []string `json:"rationale"`
}

type RankedNode struct {
	Node      ragnode.RAGNode `json:"node"`
	Candidate ragnode.Candidate `json:"candidate"`
	Score     Score         `json:"score"`
}

type RankedSet struct {
	QueryID    string       `json:"query_id"`
	Query     string       `json:"query"`
	Ranked    []RankedNode `json:"ranked"`
	Suppressed []Score      `json:"suppressed"`
}

type Ranker interface {
	Rank(ctx access.Context, analysis ragnode.RAGAnalysis, tokenBudget int) (RankedSet, error)
}

type AuthorityScorer interface {
	Authority(node ragnode.RAGNode, analysis ragnode.RAGAnalysis) float64
}

type GraphScorer interface {
	GlobalCentrality(analysis ragnode.RAGAnalysis) map[string]float64
	PersonalizedPageRank(analysis ragnode.RAGAnalysis, seeds map[string]float64) map[string]float64
}

type DiversityScorer interface {
	Diversity(node ragnode.RAGNode, selected []ragnode.RAGNode) float64
}

type TokenBudgetScorer interface {
	TokenBudgetFit(node ragnode.RAGNode, tokenBudget int) float64
}
AuthorityScore
AuthorityScore =
  clamp01(
    0.35 * Trust
  + 0.20 * normalizedIncomingCites
  + 0.20 * approvedByScore
  + 0.15 * chunkTypeAuthority
  + 0.10 * versionCurrentness
  )

chunkTypeAuthority defaults:

ChunkType	Authority prior
permission-record	1.00
decision	0.95
procedure	0.90
definition	0.85
document	0.80
section	0.75
community-summary	0.70
claim	0.65
chunk	0.60
entity	0.55
TokenBudgetFit
tokenDensity = AuthorityScore / max(1, log2(TokenCount + 2))
lengthFit =
  1.00 if TokenCount <= 0.08 * tokenBudget
  0.80 if TokenCount <= 0.15 * tokenBudget
  0.55 if TokenCount <= 0.25 * tokenBudget
  0.25 otherwise

TokenBudgetFit = clamp01(0.60 * lengthFit + 0.40 * normalizedTokenDensity)

This replaces the old EnergyTimeFit in RAG mode. The old score remains unchanged for the existing CLI.

PPR

HippoRAG 2 makes PPR relevant for multi-hop retrieval, but CitewiseRAG must not let PPR replace all global authority. The ranker therefore computes:

GraphImportance = 0.55 * WeightedCentrality + 0.45 * PersonalizedPageRank

WeightedCentrality comes from the existing Citewise edge-weight model. PPR is seeded by upstream QueryRelevance, with restart probability alpha=0.15, max iterations 100, tolerance 1e-8, and edge transition weights from the ontology table. If the graph has fewer than two usable edges, PPR returns zeroes and the ranker uses only weighted centrality.

3.5 pkg/packer

pkg/packer converts ranked nodes into an ordered ContextPlan.

package packer

import "github.com/mojomast/citewiseussy/pkg/ragnode"

type QueryType string

const (
	QueryFactual     QueryType = "Factual"
	QueryComparative QueryType = "Comparative"
	QueryProcedural QueryType = "Procedural"
	QueryExploratory QueryType = "Exploratory"
	QueryTemporal    QueryType = "Temporal"
	QueryAdversarial QueryType = "Adversarial"
	QueryAgentic     QueryType = "Agentic"
)

type SlotType string

const (
	SlotOverview     SlotType = "overview"
	SlotFoundation   SlotType = "foundation"
	SlotBridge       SlotType = "bridge"
	SlotCounterpoint SlotType = "counterpoint"
	SlotProcedure    SlotType = "procedure"
	SlotPermission   SlotType = "permission"
	SlotDecision     SlotType = "decision"
	SlotSupport      SlotType = "support"
)

type SlotPosition string

const (
	PositionFront  SlotPosition = "front"
	PositionMiddle SlotPosition = "middle"
	PositionBack   SlotPosition = "back"
)

type HygieneSignal string

const (
	HygieneGreen  HygieneSignal = "green"
	HygieneYellow HygieneSignal = "yellow"
	HygieneRed    HygieneSignal = "red"
)

type ContextSlot struct {
	Index       int                 `json:"index"`
	SlotType    SlotType            `json:"slot_type"`
	NodeID      string              `json:"node_id"`
	Role        string              `json:"role"`
	Title       string              `json:"title"`
	Text        string              `json:"text"`
	TokenCount  int                 `json:"token_count"`
	Score       float64             `json:"score"`
	Position    SlotPosition        `json:"position"`
	Source      ragnode.SourceRef   `json:"source"`
	SourceTrail []ragnode.SourceHop `json:"source_trail"`
	MustCite    bool                `json:"must_cite"`
	Rationale   string              `json:"rationale"`
}

type SuppressedEntry struct {
	NodeID string  `json:"node_id"`
	Reason string  `json:"reason"` // duplicate | stale | low-trust | budget | access-control
	Detail string  `json:"detail,omitempty"`
	Score  float64 `json:"score,omitempty"`
}

type ContextPlan struct {
	QueryID         string              `json:"query_id"`
	QueryType       QueryType           `json:"query_type"`
	Slots           []ContextSlot       `json:"slots"`
	Suppressed      []SuppressedEntry   `json:"suppressed"`
	EvidencePath    []string            `json:"evidence_path"`
	SourceTrail     []ragnode.SourceHop `json:"source_trail"`
	CritiqueSummary string              `json:"critique_summary"`
	TokensUsed      int                 `json:"tokens_used"`
	HygieneSignal   HygieneSignal       `json:"hygiene_signal"`
	WriteBackPayload any                `json:"write_back_payload,omitempty"`
}

type Packer interface {
	Pack(analysis ragnode.RAGAnalysis, queryType QueryType, tokenBudget int, callerClearance string) ContextPlan
}
Required packing behavior
Never include nodes suppressed by access control.
Never include more than one node from a duplicate cluster unless the second node has a different ChunkType and the query type is Comparative or Adversarial.
If a node has SupersededBy, suppress it as stale unless the query type is Temporal and the query explicitly asks for historical state.
If a required slot cannot be filled, set HygieneSignal=red.
If all required slots are filled but at least one required slot has score < 0.55, set HygieneSignal=yellow.
Otherwise set HygieneSignal=green.
Agentic mode must fail closed: no permission slot means HygieneSignal=red.
3.6 pkg/hygiene

The existing HygieneReport already detects duplicates, orphans, stale items, and missing prerequisite bridges. The new package wraps and extends it.

package hygiene

import (
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

type EdgeSuggestion struct {
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type HygieneReport struct {
	DuplicateClusters [][]string        `json:"duplicate_clusters"`
	OrphanNodes       []string          `json:"orphan_nodes"`
	StaleNodes        []string          `json:"stale_nodes"`
	MissingBridges    []string          `json:"missing_bridges"`
	MissingEdges      []EdgeSuggestion  `json:"missing_edges"`
	Score             float64           `json:"score"`
	Signal            packer.HygieneSignal `json:"signal"`
}

type Analyzer interface {
	SuggestMissingEdges(analysis ragnode.RAGAnalysis) []EdgeSuggestion
	HygieneScore(analysis ragnode.RAGAnalysis) float64
	CorrectiveSignal(analysis ragnode.RAGAnalysis, threshold float64) packer.HygieneSignal
}
SuggestMissingEdges

Infer possible missing edges using deterministic heuristics:

topicOverlap >= 0.60 and same CommunityID -> same-question, confidence 0.70
titleSimilarity >= 0.88 -> duplicate, confidence 0.90
same CommunityID and one node is community-summary -> community-member-of, confidence 0.80
one title contains "policy", other contains "exception" -> exception-to, confidence 0.55
one node SupersededBy == other.ID -> supersedes, confidence 1.00
HygieneScore
HygieneScore =
  clamp01(
    1.00
  - 0.18 * orphanRatio
  - 0.18 * duplicateRatio
  - 0.20 * staleRatio
  - 0.20 * missingBridgeRatio
  - 0.14 * lowTrustRatio
  - 0.10 * unapprovedAgenticRatio
  )
CorrectiveSignal
if HygieneScore < threshold: red
else if HygieneScore < 0.70: yellow
else: green

Default threshold: 0.55.

A red corrective signal means do not pack final context unless the caller sets AllowDegradedPlan=true; instead return NeedsMoreRetrieval with suggested retrieval targets.

3.7 pkg/router

The router is deterministic. It does not call an LLM.

package router

import "github.com/mojomast/citewiseussy/pkg/packer"

type RetrievalMode string

const (
	ModeGlobalGraph       RetrievalMode = "GlobalGraph"
	ModeLocalNeighborhood RetrievalMode = "LocalNeighborhood"
	ModeDRIFTChain        RetrievalMode = "DRIFTChain"
	ModeHybridBM25Dense   RetrievalMode = "HybridBM25Dense"
	ModeBasicVector       RetrievalMode = "BasicVector"
)

type GraphMetadata struct {
	EntityIDs          []string          `json:"entity_ids"`
	Topics            []string          `json:"topics"`
	CommunityIDs      []string          `json:"community_ids"`
	RoleCounts        map[string]int    `json:"role_counts"`
	ChunkTypeCounts   map[string]int    `json:"chunk_type_counts"`
	HasPermissionNode bool              `json:"has_permission_node"`
	HasDecisionBasis  bool              `json:"has_decision_basis"`
	MaxTopicSpan       int              `json:"max_topic_span"`
}

type Recommendation struct {
	QueryType         packer.QueryType `json:"query_type"`
	RetrievalMode    RetrievalMode    `json:"retrieval_mode"`
	ContextBudgetHint int             `json:"context_budget_hint"`
	Reasons          []string         `json:"reasons"`
	CounterpointRequired bool          `json:"counterpoint_required"`
}

type QueryRouter interface {
	Route(query string, metadata GraphMetadata) Recommendation
}
Complete decision tree

Apply rules in order:

If metadata has permission-record or decision-basis nodes and the query contains action verbs (create, issue, approve, delete, send, change, execute, deploy, grant, revoke, purchase, refund, file) → QueryType=Agentic, RetrievalMode=DRIFTChain, budget 6000.
If query contains explicit temporal cues (current, latest, today, as of, changed, new, old, superseded, version, year-like token) → Temporal, HybridBM25Dense, budget 4500.
If query contains procedural cues (how do I, steps, runbook, procedure, workflow, SOP, troubleshoot) → Procedural, LocalNeighborhood, budget 5000.
If query contains comparison cues (compare, versus, vs, tradeoff, difference, which is better) → Comparative, DRIFTChain, budget 5500.
If query contains adversarial/safety cues (should I, is it safe, risk, can I, allowed, not, never, forbidden, exception, counterexample) → Adversarial, DRIFTChain, budget 6000, counterpoint required.
If query references one exact entity in metadata.EntityIDs → Factual, LocalNeighborhood, budget 3500.
If inferred topic span > 3 or query starts with broad sensemaking cues (what are the themes, summarize, overview, map, explain the landscape) → Exploratory, GlobalGraph, budget 6500.
If query length <= 8 tokens and no entity match → Factual, HybridBM25Dense, budget 3000.
Fallback → Factual, HybridBM25Dense, budget 4000.
3.8 pkg/memory
package memory

import "github.com/mojomast/citewiseussy/pkg/packer"

type MemoryWriteBack interface {
	StoreContextPlan(queryID string, plan packer.ContextPlan) error
	LoadPriorPlan(queryID string) (packer.ContextPlan, bool, error)
	SimilarPriorPlans(topics []string, limit int) ([]packer.ContextPlan, error)
}

Default implementation:

type FileStore struct {
	Path string
}

type StoredPlan struct {
	QueryID    string             `json:"query_id"`
	Topics     []string           `json:"topics"`
	Plan       packer.ContextPlan `json:"plan"`
	CreatedAt  string             `json:"created_at"`
	PlanHash   string             `json:"plan_hash"`
}

Rules:

Store one JSON object per line in citewiserag_memory.jsonl.
PlanHash = sha256(QueryType + EvidencePath + node versions).
On LoadPriorPlan, re-run access gating against the current caller before returning any slot.
SimilarPriorPlans uses deterministic topic Jaccard similarity, not embeddings.
File writes must be atomic: write temp file or append with O_APPEND, then Sync.
3.9 cmd/serve

cmd/serve is optional and must compile separately from the library. It uses only net/http, encoding/json, context, and log/slog.

Endpoints:

Endpoint	Method	Purpose
/health	GET	Return service status and graph hygiene signal.
/router	POST	Return QueryType, retrieval mode, and budget hint.
/rank	POST	Rank candidates after access gating.
/pack	POST	Produce a ContextPlan.
/hygiene	POST	Return hygiene score, signal, and missing-edge suggestions.
/rank request
{
  "query_id": "q-123",
  "query": "Can the agent issue a refund for account A?",
  "token_budget": 6000,
  "access": {
    "caller_id": "user-17",
    "groups": ["support-tier-2"],
    "clearance": "confidential",
    "trusted_approvers": ["legal", "finance-ops"]
  },
  "nodes": [],
  "edges": [],
  "candidates": [
    {
      "node_id": "perm-refund-policy-v3",
      "query_relevance": 0.94,
      "method_scores": [
        {"method": "bm25", "rank": 1, "score": 17.2},
        {"method": "dense", "rank": 4, "score": 0.82},
        {"method": "graph", "rank": 2, "score": 0.91}
      ],
      "reranker_score": 0.97
    }
  ]
}
/pack response
{
  "query_id": "q-123",
  "query_type": "Agentic",
  "tokens_used": 3120,
  "hygiene_signal": "green",
  "evidence_path": [
    "perm-refund-policy-v3",
    "decision-refund-exception-2026-04",
    "procedure-refund-runbook-v2"
  ],
  "slots": [
    {
      "index": 0,
      "slot_type": "permission",
      "node_id": "perm-refund-policy-v3",
      "role": "foundation",
      "title": "Refund Permission Policy v3",
      "token_count": 820,
      "score": 91.4,
      "position": "front",
      "must_cite": true,
      "source": {
        "node_id": "perm-refund-policy-v3",
        "origin": "graphrag",
        "source": "Finance Policy KB",
        "version": "v3",
        "observed_at": "2026-05-14T12:00:00Z",
        "updated_at": "2026-04-30T09:00:00Z",
        "locator": {"document_id": "finance-refunds", "section_path": "Refunds > Permissions"}
      },
      "source_trail": [
        {"node_id": "perm-refund-policy-v3", "edge_type": "retrieved", "confidence": 0.94}
      ],
      "rationale": "Required permission-record for Agentic mode; approved by finance-ops; current version."
    }
  ],
  "suppressed": [
    {
      "node_id": "refund-policy-v1",
      "reason": "stale",
      "detail": "superseded by perm-refund-policy-v3",
      "score": 44.1
    }
  ],
  "critique_summary": "Agentic plan includes permission, controlling decision, and procedure. One stale policy was suppressed. No restricted nodes were exposed.",
  "source_trail": [
    {"node_id": "perm-refund-policy-v3", "edge_type": "retrieved", "confidence": 0.94},
    {"node_id": "decision-refund-exception-2026-04", "edge_type": "decision-basis", "confidence": 0.88}
  ],
  "write_back_payload": {
    "plan_hash": "sha256:...",
    "topics": ["refunds", "permissions", "customer-support"]
  }
}
4. Scoring Formula and Justification

The scoring formula is deterministic and inspectable. The cited systems justify the factors, but not the exact constants; the constants below are the initial production policy and must be calibrated with golden-set evaluation before release. GraphRAG motivates graph structure, community summaries, and local/global modes; HippoRAG 2 motivates PPR; Anthropic’s Contextual Retrieval and reranking findings motivate giving query relevance the largest single weight; RRF motivates accepting fused upstream relevance; and Lost-in-the-Middle motivates token budget and context-position handling.

4.1 Default Formula
Positive =
  0.28 * QueryRelevance
+ 0.18 * AuthorityScore
+ 0.14 * GraphImportance
+ 0.10 * Readiness
+ 0.10 * Freshness
+ 0.10 * TokenBudgetFit
+ 0.05 * DiversityBonus
+ 0.03 * BridgeBonus
+ 0.02 * CounterpointBonus

Negative =
  0.14 * RedundancyPenalty
+ 0.10 * StalenessPenalty
+ 0.08 * LowTrustPenalty

Total = round1(100 * clamp01(Positive - Negative))
4.2 Weight Rationale
Factor	Weight	Reason
QueryRelevance	0.28	Upstream RRF/cross-encoder relevance is the best direct query signal. Anthropic’s retrieval writeup emphasizes BM25+dense fusion and reranking as major recall improvements.
AuthorityScore	0.18	Assembly should prefer canonical, approved, high-trust sources over semantically similar commentary. This is the central assembly distinction.
GraphImportance	0.14	GraphRAG and HippoRAG-style systems show value from graph structure and multi-hop association, but graph centrality must not swamp query relevance.
Readiness	0.10	Existing Citewise prerequisite readiness prevents advanced nodes from appearing without required background.
Freshness	0.10	Current policy, version, and timestamp are mandatory for production agents.
TokenBudgetFit	0.10	Long, low-density nodes must not displace short authoritative nodes; this also reduces lost-in-the-middle risk.
DiversityBonus	0.05	Avoids over-packing one source/community and missing counter-evidence.
BridgeBonus	0.03	Important for multi-hop/cross-community queries, but slot policy also handles bridges.
CounterpointBonus	0.02	Ensures counterpoints surface when present, but adversarial/comparative packing enforces stronger requirements.
RedundancyPenalty	0.14	Duplicate context wastes tokens and hurts assembly.
StalenessPenalty	0.10	Superseded or old trend content can be actively harmful.
LowTrustPenalty	0.08	Low-trust content may still be useful as a lead but should not dominate.
4.3 Query-Type Modifiers

After default scoring, apply query-type multipliers:

QueryType	Modifier
Factual	foundation +8, definition +5, community-summary -3
Comparative	bridge +8, counterpoint +8, duplicate cluster diversity required
Procedural	procedure +12, permission-record +6, overview -5
Exploratory	overview +12, bridge +8, curiosity-leaf +3
Temporal	Freshness weight +0.08, SupersededBy penalty hard unless historical query
Adversarial	counterpoint +15, evidence-against +10, low-trust penalty doubled
Agentic	permission-record +20, decision +15, procedure +12; fail closed if permission missing
5. QueryType Routing and Slot-Filling
5.1 Slot-Filling Table

Notation: R = required, O = optional, F = forbidden unless explicitly requested.

QueryType	overview	foundation	bridge	counterpoint	procedure	permission	decision
Factual	0–1 O	1–3 R	0–1 O	0–1 O	F	0–1 O if policy-gated	0–1 O
Comparative	1 O	2–4 R	1–2 R	1 R	F	0–1 O	0–1 O
Procedural	0–1 O	1–2 R	0–1 O	0–1 O	1–3 R	0–1 R if action	0–1 O
Exploratory	1–3 R	1–2 O	1–3 R	0–1 O	F	F	0–1 O
Temporal	0–1 O	1–3 R current	0–1 O	0–1 O stale/changed	F	0–1 O	1 O if decision history
Adversarial	0–1 O	1–3 R	0–1 O	1–2 R	0–1 O	0–1 R if action	0–1 O
Agentic	0–1 O	1–2 R	0–1 O	1 O, R if risk	1–3 R	1–2 R	1–2 R
5.2 Agentic Mode

Agentic mode is used when an AI agent will act, not merely answer. It prioritizes:

permission-record
decision with decision-basis edges
procedure
foundation
counterpoint or evidence-against for safety/risk
source trail completeness

Hard failures:

Failure	Signal
No permitted permission-record	red
Permission record present but unapproved	red
Procedure required but absent	red
Source trail missing for any required slot	red
Required node suppressed by access control	red
Only stale/superseded policy available	red
6. Edge Ontology

All edge types are lowercase kebab-case after normalization. Unknown edge types are accepted but weighted as related-to with warning rationale.

Edge type	Weight	Direction	Aliases	Meaning
prerequisite	1.50	prerequisite → dependent	prereq, requires	Source must be understood or satisfied before target.
supersedes	1.40	current → old	updates, replaces, supersedes-version	Source replaces target; target is stale unless historical query.
decision-basis	1.35	decision → evidence/policy	basis, decided-from	Decision relies on target evidence or policy.
approved-by	1.30	node → approver/authority	approved, reviewed-by	Node has human/system approval provenance.
overview	1.25	overview → covered node	review, surveys, primer-for	Source summarizes or introduces target.
evidence-for	1.20	evidence → claim/decision	supports, proves	Source supports target.
evidence-against	1.20	evidence → claim/decision	refutes, opposes	Source contradicts or weakens target.
implements	1.20	implementation → requirement	implements-policy, satisfies	Source implements target requirement.
exception-to	1.15	exception → rule/policy	exempts, waives	Source is an exception to target.
cites	1.10	citing → cited	mentions, mention	Source cites, mentions, or depends on target as reference.
applies-to	1.10	policy/procedure → entity/context	governs, covers	Source applies to target entity/process/user segment.
version-of	1.00	version → canonical family	revision-of, variant-of	Source is a version of target family.
contradicts	1.00	counterpoint → target	counterpoint, responds-to	Source disputes target.
community-member-of	0.90	member → community-summary	in-community, cluster-member	Source belongs to a community/cluster.
same-question	0.80	peer ↔ peer	same-topic, similar	Nodes answer the same or very similar question.
owned-by	0.70	node → owner	owner, maintained-by	Node is owned by a team/person/system.
related-to	0.60	source → target	relates, associated-with	Fallback relation when semantics are weak.
duplicate	0.20	duplicate → duplicate	duplicates, duplicate-of, dupe	Nodes are duplicate or near-duplicate evidence.

Normalization must preserve the original edge type in Edge.Note or Edge.Attributes["raw_type"] when available.

7. Integration Contracts
7.1 Microsoft GraphRAG Python Indexer Output

Microsoft GraphRAG writes default pipeline outputs as parquet tables and documents schemas for communities, community_reports, covariates, documents, entities, relationships, and text_units.

Mapping:

GraphRAG table	CitewiseRAG mapping
documents	RAGNode{ChunkType: document, Text: text, Title: title}
text_units	RAGNode{ChunkType: chunk, Text: text, TokenCount: n_tokens}
entities	RAGNode{ChunkType: entity, Title: title, Text: description}
relationships	Edge{SourceID: source, TargetID: target, Type: related-to, Confidence: normalized(weight)}
community_reports	RAGNode{ChunkType: community-summary, Role: overview, Text: full_content or summary, CommunityID: community}
communities	community-member-of edges from entities/text units to community-summary node
covariates	RAGNode{ChunkType: claim} plus evidence-for or evidence-against depending on claim status

Importer behavior:

Accept JSON exports by default.
Accept parquet only under build tag graphrag_parquet.
Store GraphRAG human_readable_id in Attributes["human_readable_id"].
Store GraphRAG period as ObservedAt when present.
Store community_reports.rank as an authority prior, not as QueryRelevance.
7.2 LightRAG

LightRAG retrieval provides local and global/dual-level graph-aware results. Its paper describes graph structures in indexing/retrieval and a dual-level retrieval system for low-level and high-level knowledge discovery.

Mapping:

LightRAG result	CitewiseRAG mapping
Local entity result	ChunkType=entity or chunk; role likely foundation or bridge
Global result	ChunkType=community-summary; role overview
Relationship/path result	Edge with best matching ontology type or related-to
LightRAG relevance score	Candidate.QueryRelevance after 0..1 normalization
Local/global mode	Candidate.RetrievalMode = "lightrag-local" or "lightrag-global"
7.3 Hybrid BM25 + Dense + Graph with RRF

Upstream must provide per-method score details and a final normalized fused score. RRF should be calculated upstream using:

RRFScore(d) = Σ method_weight_m * 1 / (rank_m(d) + 60)
QueryRelevance = minMaxNormalize(RRFScore)

The k=60 constant follows common documented RRF practice.

Inbound schema:

{
  "query_id": "q-456",
  "query": "What policy controls vendor access to customer data?",
  "candidates": [
    {
      "node_id": "policy-vendor-access-v5",
      "query_relevance": 0.98,
      "method_scores": [
        {"method": "bm25", "rank": 1, "score": 23.4, "weight": 1.0},
        {"method": "dense", "rank": 6, "score": 0.78, "weight": 1.0},
        {"method": "graph", "rank": 2, "score": 0.91, "weight": 1.2}
      ]
    }
  ]
}
7.4 Cross-Encoder Reranker

Decision: run before CitewiseRAG, after coarse access filtering and after initial hybrid retrieval.

Justification:

Cross-encoders and ColBERT-style rerankers are relevance models, not assembly models.
CitewiseRAG needs the relevance output as an input feature.
Running reranking after CitewiseRAG could undo slot diversity, counterpoint inclusion, or lost-in-the-middle ordering.
Upstream must not send unauthorized text to an external reranker. Coarse ACL must happen before reranking; CitewiseRAG re-applies the hard access gate.

Handoff:

{
  "node_id": "security-policy-2026",
  "query_relevance": 0.93,
  "reranker_score": 0.97,
  "method_scores": [
    {"method": "rrf", "rank": 1, "score": 0.046},
    {"method": "cross-encoder", "rank": 1, "score": 0.97}
  ]
}

Recommended upstream normalization:

QueryRelevance = 0.60 * normalizedCrossEncoderScore + 0.40 * normalizedRRFScore
7.5 AI Agent / LLM Answer Generator

Downstream receives only the ContextPlan:

ContextPlan.Slots ordered by position
ContextPlan.SourceTrail
ContextPlan.CritiqueSummary
ContextPlan.Suppressed redacted summary
ContextPlan.HygieneSignal

The agent must preserve these minimum provenance fields in its answer or action log:

Field	Required
query_id	yes
node_id for every cited slot	yes
source / origin	yes
version	yes
observed_at or updated_at	yes
locator for sections/tables/pages	yes when present
evidence_path	yes
suppressed count by reason	yes
hygiene_signal	yes
critique_summary	yes
7.6 Agent Memory Systems

WriteBackPayload is an opaque payload the caller passes to MemoryWriteBack.StoreContextPlan. For Mem0, Zep, Redis, Neo4j, or custom graph stores, callers implement the interface. CitewiseRAG ships only FileStore.

Memory write-back policy:

Write back only if HygieneSignal != red.
Write back only after the caller accepts the plan or the agent completes successfully.
Store evidence paths, not just final answer text.
On future loads, re-run access control before returning a prior plan.
8. Access Control and Provenance
8.1 Access Control Flow
CandidateSet received
  → validate node IDs
  → access gate nodes and edges
  → unauthorized nodes added to Suppressed(reason=access-control)
  → ranking only over allowed nodes
  → packing only over allowed nodes
  → memory write-back only stores allowed slots
  → memory load re-gates prior plans

Hard invariants:

No unauthorized node text may appear in RankedSet, ContextPlan, CritiqueSummary, or WriteBackPayload.
Unauthorized suppressed entries must be redacted.
Access is checked before scoring, before packing, before provenance expansion, and before memory load.
Access suppression is explicit, never silent.
8.2 Provenance Construction

Each selected slot must include:

SourceRef:
  node_id
  origin
  source
  url
  version
  observed_at
  updated_at
  locator
  community_id

SourceTrail:
  ordered chain of SourceHop values

SourceTrail construction rules:

If node was directly retrieved, first hop is {NodeID, "retrieved", Candidate.QueryRelevance}.
If node was included due to an edge from a candidate, append {NodeID, EdgeType, Edge.Confidence}.
If node was included to satisfy a required slot, append a synthetic hop with EdgeType="slot-required" and Confidence=Score.Total/100.
If node is a community summary, include the community-member-of path to representative member nodes when available.
If node is a decision, include decision-basis hops when available.
If source trail is missing for a required Agentic slot, set HygieneSignal=red.

A human reviewer must be able to answer:

Which nodes were included?
Why were they included?
Which nodes were excluded?
Was anything excluded by access control?
Which version of the source was used?
Where in the source document/table did the evidence come from?
Which graph path justified multi-hop inclusion?
9. Lost-in-the-Middle Mitigation

Lost-in-the-Middle findings show that relevant information placed near the beginning or end of long context is often used better than information buried in the middle. CitewiseRAG therefore orders context by role and query type rather than raw score alone.

9.1 Position Rules
Position	Content
Front 0–20%	Highest-authority required slots: permission, foundation, current policy, key definition.
Middle 20–80%	Overviews, supporting chunks, bridge nodes, detailed tables, secondary evidence.
Back 80–100%	Counterpoint, caveat, source-trail summary, final controlling decision, or last critical constraint.
9.2 Exact Ordering Algorithm
Split selected slots into front, middle, and back.
Put permission-record first for Agentic.
Put highest foundation second unless permission is absent.
Put procedure after permission/foundation in Agentic and Procedural.
Put overview before details for Exploratory, otherwise middle.
Put bridge in middle unless Comparative, where the best bridge goes front-adjacent.
Put counterpoint at the back for Adversarial, Comparative, and risk-bearing Agentic.
Put decision at the back if it is the final controlling decision; otherwise front-adjacent.
Stable-sort within each band by score descending, then node ID ascending.
If TokensUsed > tokenBudget, remove lowest-score middle support first, then optional overview, then optional bridge. Never remove required slots unless returning red.
10. Agent Memory and Write-Back

The write-back hook exists because repeated agent runs should not rebuild the same context path every time. The write-back stores the assembled context path, slot roles, scores, provenance, and source versions.

10.1 WriteBackPayload Shape
{
  "plan_hash": "sha256:abc...",
  "query_id": "q-123",
  "query_type": "Agentic",
  "topics": ["refunds", "permissions"],
  "evidence_path": [
    "perm-refund-policy-v3",
    "decision-refund-exception-2026-04",
    "procedure-refund-runbook-v2"
  ],
  "node_versions": {
    "perm-refund-policy-v3": "v3",
    "procedure-refund-runbook-v2": "v2"
  },
  "hygiene_signal": "green",
  "created_at": "2026-05-14T12:00:00Z"
}
10.2 Reuse Policy

A prior plan can be reused only if:

All required node versions still match.
No node is now superseded.
Caller access still permits all slots.
Query topic Jaccard similarity is >= 0.70.
QueryType matches, except Factual prior plans may seed Agentic retrieval but cannot satisfy Agentic permission slots alone.
10.3 Store Adapters

Core ships:

FileStore: JSONL, stdlib only

Caller-owned adapters:

RedisStore: caller-owned implementation using redis/go-redis
Neo4jStore: caller-owned implementation using neo4j-go-driver
Mem0Store: caller-owned HTTP/client adapter
ZepStore: caller-owned HTTP/client adapter

These adapters are intentionally outside core to preserve zero required infrastructure dependencies.

11. Non-Goals
Non-goal	Justification
No embedding generation	CitewiseRAG consumes QueryRelevance; embeddings are upstream.
No document chunking or ingestion pipeline	Ingestion is source-specific and often requires LLM/contextualization; CitewiseRAG assembles after retrieval.
No LLM calls	Determinism, auditability, and testability are core requirements.
No vector database management	Vector stores are retrieval infrastructure, not assembly infrastructure.
No community detection	CitewiseRAG accepts CommunityID; GraphRAG/LightRAG/RAPTOR-like systems compute communities upstream.
No Python interop	Go-only library and CLI. Python outputs are consumed via JSON/parquet files, not embedded Python.
No real-time graph mutation during a query	Assembly is read-only. Write-back happens after successful plan generation and may be async in the caller.
No replacement of existing CLI behavior	Backward compatibility is mandatory; new RAG functionality is additive.
12. Migration Guide
12.1 Existing Users

No existing command changes:

citewise roles --file backlog.json
citewise score --file backlog.json
citewise queue --file backlog.json
citewise explain --file backlog.json --id item-1
citewise hygiene --file backlog.json
citewise export --file backlog.json

These commands remain backed by pkg/citewise. The existing CLI parser and command list remain valid.

12.2 New RAG Library Users

New library usage:

candidateSet := ragnode.CandidateSet{...}

analysis, err := rag.Analyze(candidateSet)
ranked, err := rag.DefaultRanker().Rank(accessCtx, analysis, 6000)
plan := rag.DefaultPacker().Pack(analysis, packer.QueryAgentic, 6000, "confidential")
12.3 Backlog JSON Compatibility

Existing JSON files parse as before. When converted to RAGNode, defaults are:

Text = Item.Notes
ChunkType = document
TokenCount = estimate(Text)
Version = ""
Sensitivity = internal
ApprovedBy = nil
CommunityID = ""
ContextPrefix = ""
12.4 New JSON Compatibility

New RAG JSON files may include all RAGNode fields. Old CLI commands ignore unknown fields through the RAG adapter path, not by modifying old parser behavior.

13. Test Strategy
13.1 Unit Tests

Use table-driven tests for:

Package	Tests
ragnode	edge normalization, RAGNode.ToItem, token estimate fallback
access	clearance hierarchy, trusted approver rules, redaction
ranker	scoring weights, PPR convergence, authority priors, duplicate penalties
packer	slot filling per QueryType, token budget trimming, lost-in-middle ordering
hygiene	missing edge suggestions, hygiene score thresholds
router	deterministic decision tree coverage
memory	atomic write, load, similar prior plans, re-gating
13.2 Property Tests

Use testing/quick from the Go standard library:

Access invariant: no unauthorized node text appears in a plan.
Budget invariant: TokensUsed <= tokenBudget unless one required slot alone exceeds budget, in which case HygieneSignal=yellow or red.
Duplicate invariant: no duplicate cluster contributes more than one slot unless allowed by query type.
Determinism invariant: same input yields byte-identical JSON output.
Provenance invariant: every slot node ID appears in EvidencePath and has a SourceRef.
13.3 Integration Tests

Fixtures:

testdata/graphrag_minimal/
  documents.json
  text_units.json
  entities.json
  relationships.json
  community_reports.json

testdata/lightrag_minimal/
  local_results.json
  global_results.json

testdata/hybrid_rrf/
  bm25_results.json
  dense_results.json
  graph_results.json
  fused_candidates.json

Contracts:

GraphRAG community report imports as ChunkType=community-summary and role overview.
LightRAG local/global results preserve retrieval mode.
Hybrid RRF candidate score maps to QueryRelevance.
Cross-encoder handoff modifies relevance but not access gating.
Agentic fixture without permission record returns red.
13.4 HygieneSignal Smoke Test
Case A: Agentic query + permission + decision + procedure + source trail => green
Case B: Agentic query + permission + procedure but low-trust decision => yellow
Case C: Agentic query + no permission record => red
Case D: Temporal query + only superseded policy => red
Case E: Factual query + one foundation slot and no duplicates => green
13.5 Regression Tests for Existing CLI

Run:

go test ./...
go test ./pkg/citewise
go test ./pkg/ranker
go test ./pkg/packer

Existing tests must not be weakened. The old fixture expectations around roles, duplicate detection, readiness, and queue planning remain valid.

14. Dependency Recommendations

Core should remain stdlib-first. Dependencies below are either optional or test-only.

Package	Version verified	Verdict	Reason	Source
gonum.org/v1/gonum/graph/network	v0.17.0	Optional, not core	Provides PageRank, HITS, betweenness, closeness. Use only behind graphx build tag if graph algorithms outgrow custom PPR. Core PPR remains small and custom.	
github.com/parquet-go/parquet-go	v0.29.0	Use optional	Best fit for GraphRAG parquet import under graphrag_parquet build tag.	
github.com/apache/arrow-go/v18	v18.6.0	Reject for MVP	Powerful but heavier than needed for simple GraphRAG parquet table import. Reconsider only if nested parquet compatibility requires Arrow.	
github.com/blevesearch/bleve/v2	v2.6.0	Reject core; optional example only	It provides BM25, vector, and hybrid/RRF features, but retrieval is explicitly upstream.	
github.com/google/go-cmp/cmp	v0.7.0	Use test-only	Better semantic diffs for table-driven tests and golden output comparisons.	
github.com/neo4j/neo4j-go-driver/v6	v6.0.0	Reject core; caller adapter only	Useful for graph-store write-back, but core must not require Neo4j.	
github.com/redis/go-redis/v9	current repo	Reject core; caller adapter only	Useful for memory cache implementations, but default memory is file-backed JSONL.	
Any HTTP router framework	n/a	Reject	net/http is enough for cmd/serve.	
Any JSON streaming library	n/a	Reject	encoding/json.Decoder is enough.	
Any LLM SDK	n/a	Reject	No LLM calls inside CitewiseRAG.	
15. Open Questions
ApprovedBy semantics: Is ApprovedBy a list of trusted approvers of the source, or an allow-list of principals allowed to see the node? This spec treats it as source approval, while Sensitivity handles visibility.
Caller identity model: Should access use only clearance and groups, or should it support attribute-based rules such as region, customer account, department, and purpose?
Token counting: Should token counts be injected by the upstream model tokenizer, or is the len(text)/4 fallback acceptable for MVP?
GraphRAG schema version: Which Microsoft GraphRAG output version will be the first supported import fixture?
Table provenance: Should Locator.TableID, RowStart, and RowEnd be mandatory for metric-bearing nodes?
Memory acceptance signal: Should write-back occur when the plan is generated, when the LLM answer is accepted, or when the agent action succeeds?
Calibration set: What golden dataset should be used to tune the scoring constants?
Red-plan behavior: Should cmd/serve /pack return HTTP 200 with HygieneSignal=red, or HTTP 409 with a corrective retrieval payload?
PPR scale limits: What graph size should trigger optional Gonum or external graph-store delegation?
Suppression audit level: Should ordinary callers see counts only for access-control suppressions, while admins see node IDs?
Agentic fail-closed policy: Are there domains where Agentic mode may proceed without a permission record, or should this always be impossible?
Community summary trust: Should GraphRAG/LightRAG-generated summaries inherit trust from underlying sources, or have a separate generated-content trust penalty?
Cross-encoder leakage: Will upstream rerankers run in the same trust boundary, or does CitewiseRAG need a pre-rerank redaction utility?
CLI additions: Should new commands be added as citewise rag pack and citewise rag route, or should RAG remain library/HTTP only for the first release?
