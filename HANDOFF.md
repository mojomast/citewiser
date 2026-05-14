# CitewiseRAG Handoff

This file is the first thing a new agent should read. It summarizes the current development state and gives explicit next instructions. Update this file at the end of every completed or paused work slice, alongside `DEVPLAN.md` progress blocks and any docs touched by the slice.

## Required Startup Sequence

1. Read `HANDOFF.md` first.
2. Read `spec.md` for product truth. Current implementation refs: memory/API/CLI around lines 616-647, 686-740, and 1097-1134; test/dependency guidance around lines 1136-1217.
3. Read `DEVPLAN.md` for task state and sequencing. Resume workflow: around lines 5-16. Documentation discipline: around lines 30-41. Progress marker protocol: around lines 56-89. Completed post-release tasks are near the end of the work-packet section.
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
- `T05.1` GraphRAG JSON mapper: `23b26db`
- `T05.2` optional GraphRAG parquet reader: `23b26db`
- `T05.3` LightRAG/hybrid/reranker handoff mappers: `23b26db`
- `T05.4` file-backed memory store: `23b26db`
- `T06.1` top-level RAG pipeline: `23b26db`
- `T06.2` optional HTTP server: `23b26db`
- `T06.3` optional stdio JSON integration: `23b26db`
- `T07.1` property and invariant tests: `95ecb14`
- `T07.2` golden fixtures, docs, and examples: `95ecb14`
- `T07.3` release gate and compatibility audit: `95ecb14`
- `T08.1` dependency boundary guardrails: `bdf6690`
- `T08.2` API and stdio JSON fixtures: `bdf6690`
- `T08.3` additive RAG CLI commands: `bdf6690`
- `T09.1` RAG CLI rank and hygiene commands: `8a31d95`
- `T09.2` RAG CLI stdin input support: `8a31d95`
- `T09.3` RAG CLI golden stdout fixtures: `8a31d95`
- `T10.1` access policy defaults: `3f4acab`
- `T10.2` token counting contract: `3f4acab`
- `T10.3` table provenance preservation: `3f4acab`
- `T10.4` pre-rerank redaction helper: `3f4acab`
- `T11.1` module path and legacy Citewise hardening: `aafd54a`
- `T12.1` GovOne library surface hardening: not committed in current slice

Latest ledger-only hash updates exist after several task commits; use `git log --oneline -10` for exact current HEAD.

The working tree contains completed and verified T12.1 GovOne library surface changes that are not committed because the user has not explicitly requested a commit.

## Next Work

Earliest unblocked task after accepting the current working tree changes: none. All planned tasks through `T12.1` are implemented and verified.

Primary files to create/update:

- No remaining planned implementation tasks after T12.1.
- If work continues, use `spec.md` open questions and release feedback to add new `DEVPLAN.md` tasks before coding.

Relevant `DEVPLAN.md` refs:

- `T12.1`: complete in `DEVPLAN.md`.
- Open decisions remain at the end of `DEVPLAN.md` for post-MVP/product follow-up: D03, D04, D08, D09, D10, and D13 are still unresolved defaults.

Relevant `spec.md` refs:

- Test strategy: around lines 1136-1201.
- Dependency guidance: around lines 1203-1217.
- Open questions for future product decisions: around lines 1218-1232. T10 resolved ApprovedBy semantics, caller identity model, token counting, table provenance, suppression audit level, Agentic fail-closed policy, and cross-encoder leakage.

Concrete next instructions:

1. If new work is requested, add new anchored tasks to `DEVPLAN.md` before implementation.
2. Resolve open product questions in `spec.md`/`DEVPLAN.md` before changing behavior that depends on them.

T12.1 design decisions made in the current uncommitted slice:

- `hygiene.Analyzer` is now the named mockable hygiene interface used by `rag.Pipeline`.
- `memory.Store` extends existing context-plan write-back with portable session `Load`, `Save`, and `Delete`; `FileStore` satisfies it while preserving the old `MemoryWriteBack` interface.
- `access.ContextFromGovOne` maps GovOne roles to CitewiseRAG clearance and sets tenant groups, trusted approvers, and `tenant_id` attributes.
- `RAGNode.TenantID` is enforced by `DefaultController` before clearance checks.
- `ContextPlan.PlanHash` hashes structural identity only, and `SuppressedByReason` is populated by `DefaultPacker.Pack`.

T11.1 design decisions made in the current uncommitted slice:

- Canonical module path is now `github.com/mojomast/citewiser`; all internal imports and docs were updated from the old compatibility name.
- `go.mod` now uses `go 1.24` because patch versions do not belong in the `go` directive.
- Legacy `pkg/citewise` hardening is intentionally compatibility-preserving except for additive `QueuePlan.BudgetExceeded` signaling and read-item filtering in role output.

T10.1/T10.2/T10.3/T10.4 design decisions made in the current uncommitted slice:

- Access policy defaults are explicit: `ApprovedBy` is source approval, ordinary suppression audit is redacted ID/reason/detail only, broader ABAC remains caller-owned, and Agentic remains fail-closed without a permitted permission record.
- `access.AttrAllowUnapprovedAgenticNodes` replaces raw string usage for the internal validation escape hatch; production Agentic packing still suppresses unapproved controlling nodes.
- Token counting prefers upstream `TokenCount`; `ragnode.UsesEstimatedTokenCount` identifies fallback usage, and ranker rationale marks `ceil(len(Text)/4)` estimates.
- Table locator fields remain optional but are preserved through `provenance.BuildSourceRef` and packed slot source refs when supplied.
- `hybrid.RedactForReranker` lets callers drop unauthorized nodes, candidates, and edges before a reranker handoff when their trust boundary requires it; downstream access gates still re-run.

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

T07.1/T07.2/T07.3 design decisions made in the current uncommitted slice:

- Property tests are focused on stable invariants: unauthorized text absence, budget/determinism/provenance, and duplicate cluster limits.
- Golden context-plan fixtures are static byte-for-byte JSON fixtures for Agentic green, Agentic red, Temporal stale red, and Factual green shapes.
- `RELEASE_AUDIT.md` records release-gate commands and dependency-audit notes; non-tagged builds remain stdlib plus project packages, while parquet dependencies are isolated behind `graphrag_parquet`.

T08.1/T08.2/T08.3 design decisions made in the current uncommitted slice:

- Dependency boundary checks now run as a root Go test using `go list -deps`; non-tagged builds must not include parquet or rejected infra/LLM/router packages, while parquet is expected under `graphrag_parquet`.
- Public API/stdio request examples live under `testdata/api_examples` and are exercised by `cmd/serve` tests.
- Additive CLI commands live under `citewise rag route` and `citewise rag pack`; legacy commands still dispatch through `pkg/citewise` unchanged.

T09.1/T09.2/T09.3 design decisions made in the current uncommitted slice:

- `citewise rag rank` emits the same access-gated ranked set shape used by the library ranker.
- `citewise rag hygiene` emits the hygiene report for a RAG candidate set.
- `pkg/ragcli.RunWithInput` supports `--file -` for stdin without changing legacy `pkg/citewise` command behavior.
- `testdata/rag_cli_golden` locks byte-for-byte stdout for `rag route`, `rag rank`, `rag pack`, and `rag hygiene`.

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
- `go test ./pkg/rag` passed after invariant/example/golden additions.
- `go test ./...` passed repeatedly.
- `go test -tags graphrag_parquet ./...` passed.
- Existing CLI commands `roles`, `score`, `queue`, `explain`, `hygiene`, and `export` passed against `testdata/citewise_backlog.json`.
- `go list -deps ./...` and `go list -deps -tags graphrag_parquet ./pkg/integrations/graphrag` were audited and recorded in `RELEASE_AUDIT.md`.
- `go test .` passed after dependency guardrails.
- `go test ./cmd/serve` passed after API/stdio fixture tests.
- `go test ./...` passed after RAG CLI additions.
- `go run . roles --file testdata/citewise_backlog.json` passed.
- `go run . rag route --file testdata/api_examples/pack_request.json` passed.
- `go run . rag pack --file testdata/api_examples/pack_request.json --token-budget 1200` passed.
- `go run . rag rank --file testdata/api_examples/pack_request.json` passed.
- `go run . rag hygiene --file testdata/api_examples/pack_request.json` passed.
- `go run . rag route --file - < testdata/api_examples/pack_request.json` passed.
- `go test ./...` passed.
- `go test ./pkg/access` passed for T10.1.
- `go test ./pkg/ragnode ./pkg/ranker` passed for T10.2.
- `go test ./pkg/provenance ./pkg/packer` passed for T10.3.
- `go test ./pkg/integrations/hybrid` passed for T10.4.
- `go test ./...` passed after each T10 slice.
- `go build ./...` passed for T11.1.
- `go test ./...` passed for T11.1.

Parallel option after claiming only one task yourself:

- No obvious safe parallel implementation remains until new tasks are added.
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

- Core runtime packages are stdlib-first; optional GraphRAG parquet imports are isolated behind the `graphrag_parquet` build tag and guarded by release tests.
- Existing `pkg/citewise` CLI behavior is imported from upstream and must remain unchanged.
- `pkg/ragnode` owns RAG types, edge ontology, old backlog conversion, RAG JSON parsing, and deterministic analysis construction.
- `pkg/access` owns hard access decisions and strict node redaction.
- `pkg/provenance` owns source refs/trails and redaction helpers.
- Access gates must run before scoring, packing, provenance expansion, and memory load.
