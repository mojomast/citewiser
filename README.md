# CitewiseRAG

CitewiseRAG is a deterministic Go knowledge layer for assembling retrieved RAG candidates into access-controlled, provenance-aware context. It extends the existing Citewise CLI compatibility surface while adding GraphRAG-oriented packages for node modeling, access gates, provenance, and ranking helpers.

The product truth is `spec.md`. The task ledger and resume workflow live in `DEVPLAN.md`. New development sessions should start with `HANDOFF.md`.

## Current Scope

Implemented so far:

- Existing Citewise CLI behavior in `main.go` and `pkg/citewise`.
- RAG node, candidate, edge ontology, conversion, and analysis helpers in `pkg/ragnode`.
- Hard access-control gates and redaction in `pkg/access`.
- Provenance source-ref and source-trail builders in `pkg/provenance`.
- Ranker authority, token-budget fit, diversity, personalized PageRank, access-gated ranking, deterministic scoring, and suppression helpers in `pkg/ranker`.
- Context plan packing in `pkg/packer`, including query-type slot policies, provenance-bearing slots, hygiene signals, lost-in-the-middle ordering, and budget trimming.
- Graph hygiene analysis in `pkg/hygiene`, including missing-edge suggestions, hygiene scoring, corrective signals, and retrieval targets for red plans.
- Deterministic query routing in `pkg/router`, including query type, retrieval mode, budget hints, and rule-match reasons.
- Upstream handoff mappers in `pkg/integrations/graphrag`, `pkg/integrations/lightrag`, and `pkg/integrations/hybrid` for GraphRAG JSON/parquet-tag imports, LightRAG local/global results, hybrid RRF, and cross-encoder relevance handoffs.
- File-backed memory in `pkg/memory`, using JSONL write-back for non-red accepted plans, stable plan hashes, access re-gating on load, and deterministic topic-Jaccard reuse.
- Top-level orchestration in `pkg/rag`, including `rag.Analyze`, default constructors, typed pipeline errors, and end-to-end candidate-to-plan execution.
- Optional `cmd/serve` HTTP and stdio JSON surfaces for routing, ranking, packing, and hygiene.

Future milestones focus on hardening, golden fixtures, release gates, and optional post-MVP CLI additions.

## Development

Run the full test suite with:

```sh
go test ./...
```

Run package-specific tests while developing a slice, for example:

```sh
go test ./pkg/ranker
```

Optional GraphRAG parquet support is isolated behind the `graphrag_parquet` build tag:

```sh
go test -tags graphrag_parquet ./pkg/integrations/graphrag
```

The core module is stdlib-first. Do not add infrastructure, LLM, vector database, Redis, Neo4j, web-router, or graph-algorithm dependencies to core packages unless a later task explicitly changes that rule.

## Memory

`pkg/memory.FileStore` writes one JSON object per line to `citewiserag_memory.jsonl` by default. Callers should invoke `StoreContextPlan` only after a plan is accepted by the caller or an agent completes successfully; red plans are rejected. Loads and reuse checks re-run access control against the store's current caller context and optional current node map, reject superseded or version-mismatched slots, and require topic Jaccard similarity of at least `0.70`.

## Server

The optional server compiles separately from the library:

```sh
go run ./cmd/serve
```

It exposes `GET /health` and `POST /router`, `/rank`, `/pack`, and `/hygiene`. Red `/pack` plans return HTTP 200 with `hygiene_signal: "red"` so clients can inspect corrective detail rather than treating red plans as transport failures.

For stdio agent integrations, run:

```sh
go run ./cmd/serve stdio
```

The stdio request shape is `{"operation":"router|rank|pack|hygiene","request":{...}}`; the response shape is `{"ok":true,"response":{...}}` or `{"ok":false,"error":"..."}`.

## Documentation Updates

Update this README whenever a completed task changes one of these user-visible facts:

- Implemented package behavior or public APIs.
- CLI commands, HTTP endpoints, JSON schemas, fixtures, examples, or operational workflows.
- Access-control, provenance, memory, ranking, packing, routing, or integration rules that callers must understand.
- Setup, test, release, or dependency guidance.

Do not document aspirational behavior as implemented. If a feature is planned but incomplete, refer to its task ID in `DEVPLAN.md` instead.

At the end of each development slice, update `DEVPLAN.md` and `HANDOFF.md` with whether README changes were needed and why.
