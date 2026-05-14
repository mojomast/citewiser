# CitewiseRAG Handoff

This file is the first thing a new agent should read. It summarizes the current development state and gives explicit next instructions. Update this file at the end of every completed or paused work slice, alongside `DEVPLAN.md` progress blocks and any docs touched by the slice.

## Required Startup Sequence

1. Read `HANDOFF.md` first.
2. Read `spec.md` for product truth. Current ranker refs: `spec.md` around lines 296-402 and 743-789.
3. Read `DEVPLAN.md` for task state and sequencing. Resume workflow: around lines 5-16. Documentation discipline: around lines 30-41. Progress marker protocol: around lines 56-89. Next tasks: around lines 392-428.
4. Check git before editing: `git status --short`, `git diff`, `git log --oneline -5`.
5. Identify the earliest unblocked `[ ]` task in `DEVPLAN.md`. Do not take over `[/]` work unless stale or explicitly approved.
6. Claim exactly one task by changing its checkbox to `[/]` and adding a new `PROGRESS` block immediately under that task.
7. Implement in a circular slice: code, tests, docs, verification, `DEVPLAN.md`, `HANDOFF.md`, then commit if the user asked for normal development/commits or the dev plan task requires it.

## GitHub Remote And Push

Canonical GitHub repository: `https://github.com/mojomast/citewiser`

Current local repository may not have a named remote configured. Before pushing, check with `git remote -v`.

Push instructions:

1. Commit all completed slice changes first. Do not push uncommitted work.
2. Verify the branch and status with `git branch --show-current` and `git status --short`.
3. If a remote named `origin` already points to `https://github.com/mojomast/citewiser.git`, push with `git push origin HEAD:main`.
4. If no remote is configured, push without changing git config: `git push https://github.com/mojomast/citewiser.git HEAD:main`.
5. If a remote points somewhere else, do not overwrite it without user approval; use the explicit URL push form above.
6. Never force-push unless the user explicitly requests it. Never force-push to `main` without warning first.

## Current State

Completed task commits:

- `T00.1` bootstrap Go module: `853b335`
- `T00.2` Citewise compatibility anchor: `a184869`
- `T01.1` RAGNode types: `e2b249f`
- `T01.2` edge ontology normalization: `8a80294`
- `T01.3` conversion and analysis builder: `eaa365a`
- `T02.1` access controller: `c43759e`
- `T02.2` provenance builders/redaction: `cfde974`
- `T03.1` authority/token-budget/diversity scorers: `2a66c2b`
- `T03.2` personalized PageRank: `2a66c2b`
- `T03.3` ranker scoring/suppression: `2a66c2b`
- `T04.1` slot policies/context plan: `2a66c2b`
- `T04.2` lost-in-the-middle ordering/budget trimming: `2a66c2b`
- `T04.3` hygiene analyzer: `148a510`
- `T04.4` deterministic query router: `148a510`
- `T05.1` GraphRAG JSON mapper: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.
- `T05.2` optional GraphRAG parquet reader: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.
- `T05.3` LightRAG/hybrid/reranker handoff mappers: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.
- `T05.4` file-backed memory store: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.
- `T06.1` top-level RAG pipeline: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.
- `T06.2` optional HTTP server: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.
- `T06.3` optional stdio JSON integration: implemented and verified in working tree, not committed because no commit was explicitly requested in this slice.

Latest ledger-only hash updates exist after several task commits; use `git log --oneline -10` for exact current HEAD.

The working tree has intentional uncommitted T05.1-T06.3 changes when this handoff was updated: README/docs updates, go.mod/go.sum parquet dependency updates, GraphRAG/LightRAG/hybrid integration code and tests, GraphRAG fixture data, memory store code/tests, rag pipeline code/tests, and cmd/serve HTTP/stdio code/tests.

## Next Work

Earliest unblocked task after accepting the current working tree changes: `T07.1 Implement Property And Invariant Tests`.

Primary files to create/update:

- package test files across `pkg/*`
- `DEVPLAN.md`
- `HANDOFF.md`
- `README.md` if new test/golden behavior is documented

Relevant `DEVPLAN.md` refs:

- `T07.1`: around lines 732-746, depends on T02.1, T03.3, T04.1, and T05.4.
- `T07.2`: follows T07.1 and depends on all M01-M06 tasks.
- Parallelization map: around lines 687-707.

Relevant `spec.md` refs:

- Property/invariant test strategy: around lines 1137-1148.
- Access-control invariants: around lines 969-984.
- Lost-in-the-middle/budget invariants: around lines 1021-1040.
- Unit-test strategy: around lines 1137-1148.
- Dependency guidance: around lines 1203-1217. Keep core stdlib-only unless a task explicitly changes that rule.

Concrete `T07.1` instructions:

1. Claim `T07.1` in `DEVPLAN.md` with a new `PROGRESS` block.
2. Add property/invariant tests using `testing/quick` where it fits naturally.
3. Focus on unauthorized text absence, budget bounds, duplicate limits, deterministic JSON, and provenance coverage.
4. Run `go test ./...` repeatedly enough to catch nondeterminism.
5. Update `DEVPLAN.md`, `HANDOFF.md`, and README/docs if public test/golden behavior changes.

T03.1 design decisions made in the current uncommitted slice:

- Version currentness is `1` for non-superseded versioned nodes, `0.5` for non-superseded nodes missing `Version`, and `0` for superseded nodes.
- Diversity starts at `1`, subtracts `0.35` for repeated `Source`, subtracts `0.25` for repeated `CommunityID`, and clamps to `0..1`.
- `TokenBudgetFit` normalizes density across `analysis.Nodes`; `TokenBudgetFitSingle` clamps the single-node density as a bounded fallback.

T03.2/T03.3 design decisions made in the current uncommitted slice:

- PPR uses `Candidate.QueryRelevance` restart seeds, `ragnode.EdgeWeight` transition weights, alpha `0.15`, max `100` iterations, tolerance `1e-8`, and zero-map fallback for fewer than two usable edges.
- `DefaultRanker` ranks candidate nodes when candidates are present; if no candidates are present, it ranks all analysis nodes for test/adapter convenience.
- Query-type modifiers are read from `access.Context.Attributes["query_type"]` until `pkg/packer` query-type constants are implemented.
- Access suppression returns only score metadata and does not include redacted node copies.

T04.1/T04.2 design decisions made in the current uncommitted slice:

- `DefaultPacker` uses `DefaultRanker` and passes `callerClearance` through to access control; it sets `allow_unapproved_agentic_nodes=true` only so packer can enforce Agentic approval failures explicitly and return red plans.
- Agentic controlling nodes without `ApprovedBy` are suppressed at packing time and missing required Agentic slots produce red hygiene.
- Ordering prioritizes Agentic permission, foundation, then procedure before score sorting; budget trimming removes middle support, then optional overview, then optional bridge, and leaves required over-budget slots with red hygiene.

T04.3/T04.4 design decisions made in the current uncommitted slice:

- Hygiene red reports include deterministic retrieval targets derived from missing bridges and missing-edge suggestions when degraded plans are disallowed.
- Title similarity uses token Jaccard; topic overlap uses intersection over smaller topic set for deterministic heuristic matching.
- Router applies the spec decision tree in exact order and returns only the first matched rule reason; Agentic wins before temporal if both are present.

T05.1/T05.2/T05.3 design decisions made in the current uncommitted slice:

- GraphRAG JSON imports are dependency-free and accept one export object containing GraphRAG tables.
- Optional parquet support is isolated behind `graphrag_parquet`; without the tag, `ParseParquetFiles` returns `ErrParquetSupportDisabled`.
- `github.com/parquet-go/parquet-go` is only imported by build-tagged files, but its module requirements are recorded in `go.mod`/`go.sum`; `go test ./...` without tags passed.
- LightRAG and hybrid mappers preserve upstream relevance metadata as input features and do not alter downstream access, ranking, diversity, or packing policy.

T05.4/T06.1/T06.2/T06.3 design decisions made in the current uncommitted slice:

- `memory.FileStore` keeps the spec `MemoryWriteBack` interface; caller access/current node state are configured on the store for re-gating and reuse checks.
- Memory write-back payloads are populated at store time, red plans return `ErrRedPlan`, and appends use `O_APPEND` plus `Sync` guarded by the store mutex.
- `rag.Pipeline` exposes `rag.Analyze`, default constructors, metadata derivation, and typed errors; red plans return `ErrRedCorrectiveSignal` unless degraded plans are allowed.
- `cmd/serve` uses stdlib HTTP only; `/pack` returns HTTP 200 for red plans so callers can inspect the red plan and corrective details.
- `go run ./cmd/serve stdio` accepts `operation` plus `request` JSON and reuses the HTTP DTOs/pipeline behavior.

Verification from the current slice:

- `go test ./pkg/ranker` passed.
- `go test ./pkg/packer` passed.
- `go test ./pkg/hygiene` passed.
- `go test ./pkg/router` passed.
- `go test ./pkg/integrations/graphrag` passed.
- `go test -tags graphrag_parquet ./pkg/integrations/graphrag` passed.
- `go test ./pkg/integrations/lightrag ./pkg/integrations/hybrid` passed.
- `go test ./pkg/memory` passed.
- `go test ./pkg/rag` passed.
- `go test ./cmd/serve` passed.
- `go test ./pkg/memory ./pkg/rag ./cmd/serve` passed.
- `go test ./...` passed.

Parallel option after claiming only one task yourself:

- A subagent may research `T07.2` golden fixture/docs requirements without writing files while you implement `T07.1`.
- Do not have subagents edit `DEVPLAN.md` or code concurrently unless each claims a separate unblocked task and the file boundaries are safe.

## Slice Completion Checklist

At the end of every slice, do all applicable items:

1. Run task-specific tests from `DEVPLAN.md`.
2. Run `go test ./...` unless the task explicitly narrows verification and full tests are impossible.
3. Update docs for any public package behavior, public API, command, fixture format, memory rule, HTTP surface, or integration mapper introduced.
4. Update `DEVPLAN.md` task marker and newest `PROGRESS` block with:
   - `Status`
   - `Evidence` with paths changed and exact tests run
   - `Spec refs`
   - `Docs`
   - `Notes`
   - `Commit`
5. Update `HANDOFF.md` with:
   - New completed task commit hash.
   - Current working tree status if relevant.
   - Exact next unblocked task.
   - Approximate line refs in `DEVPLAN.md` and `spec.md` for the next task.
   - Any design decisions or ambiguities discovered.
6. Commit task-sized changes if operating under the dev-plan workflow. Never amend or rewrite history unless explicitly requested.
7. Leave unrelated user changes untouched.

## Docs Rules

- `spec.md` is product truth. Do not weaken it to match incomplete code.
- `DEVPLAN.md` is the task ledger. Keep anchors and task IDs intact.
- `HANDOFF.md` is the operational next-step summary. Keep it concise but explicit enough that a new agent can resume from it alone.
- Public packages need useful `doc.go` comments when behavior becomes usable.
- Public exported functions/types need comments when names do not fully explain semantics.
- Do not add aspirational docs for unimplemented features unless marked as a future task with the relevant spec/task ID.
- Keep `README.md` present and current. Update it whenever a slice changes implemented package behavior, public APIs, CLI/API surfaces, schemas, fixtures, examples, operational workflows, dependency guidance, or setup/test instructions; otherwise note in `DEVPLAN.md` why no README update was required.

## Subagent Instructions

Use subagents for read-only research and independent work only when safe:

- Good parallel research: ask an explore subagent to inspect `spec.md`/`DEVPLAN.md` for formulas, edge cases, or test requirements.
- Good independent implementation only after dependencies are done: `T03.1` and `T03.2` can be built independently after `T01.3`; `T03.3` must wait for both plus `T02.1`.
- Each writing subagent must claim exactly one task in `DEVPLAN.md`, update its own `PROGRESS` block, and keep changes scoped.
- Do not let two agents edit the same files unless the boundaries are obvious and safe.
- Main agent should reconcile docs, tests, `DEVPLAN.md`, and `HANDOFF.md` before final response.

## Current Architecture Notes

- Core is stdlib-only so far.
- Existing `pkg/citewise` CLI behavior is imported from upstream and must remain unchanged.
- `pkg/ragnode` owns RAG types, edge ontology, old backlog conversion, RAG JSON parsing, and deterministic analysis construction.
- `pkg/access` owns hard access decisions and strict node redaction.
- `pkg/provenance` owns source refs/trails and redaction helpers.
- Access gates must run before scoring, packing, provenance expansion, and memory load.
