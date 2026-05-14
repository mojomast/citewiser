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

Placeholder packages remain for future milestones: `pkg/packer`, `pkg/hygiene`, `pkg/router`, `pkg/memory`, `pkg/rag`, and upstream integration packages.

## Development

Run the full test suite with:

```sh
go test ./...
```

Run package-specific tests while developing a slice, for example:

```sh
go test ./pkg/ranker
```

The core module is stdlib-first. Do not add infrastructure, LLM, vector database, Redis, Neo4j, web-router, or graph-algorithm dependencies to core packages unless a later task explicitly changes that rule.

## Documentation Updates

Update this README whenever a completed task changes one of these user-visible facts:

- Implemented package behavior or public APIs.
- CLI commands, HTTP endpoints, JSON schemas, fixtures, examples, or operational workflows.
- Access-control, provenance, memory, ranking, packing, routing, or integration rules that callers must understand.
- Setup, test, release, or dependency guidance.

Do not document aspirational behavior as implemented. If a feature is planned but incomplete, refer to its task ID in `DEVPLAN.md` instead.

At the end of each development slice, update `DEVPLAN.md` and `HANDOFF.md` with whether README changes were needed and why.
