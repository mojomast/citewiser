# CitewiseRAG Autonomous Dev Plan

This plan converts `spec.md` into agent-sized implementation work. Agents must update the markers in this file as they progress and must not remove anchors. Treat this file as the handoff ledger: a new agent should be able to read only `spec.md`, `DEVPLAN.md`, recent git history, and current tests to resume safely.

## Agent Resume Workflow

Every agent session starts here:

- Read `spec.md` first for product truth, then read this file for implementation sequencing and current status.
- Check git state before editing: `git status --short`, `git diff`, and recent commits with `git log --oneline -5`.
- Identify the earliest unblocked `[ ]` task whose dependencies are `[x]`; if a task is already `[/]`, do not take it unless it is clearly stale and the user approves takeover.
- Read the task scope, dependencies, implementation bullets, verification bullets, and related spec sections before changing code.
- Claim exactly one task by changing its checkbox to `[/]` and adding a `PROGRESS` block below that task.
- Keep work task-sized. If a task is too broad, split it by adding anchored subtasks before implementation.
- If implementation reveals that the plan does not cover the spec, update this file in the same branch before continuing code work.
- End every session by updating the task `PROGRESS` block, even if the task remains incomplete.

## Completion Checklist

A task is not complete until all items below are true:

- Code implements the referenced `spec.md` requirements without weakening existing behavior.
- Tests required by the task pass, or failures are documented in the `PROGRESS` block with exact commands and reasons.
- Documentation is updated with any new package behavior, public API, command, fixture format, or operational rule introduced by the task.
- Generated fixtures, golden files, and examples are committed with the code that needs them.
- The task checkbox is changed to `[x]` only after verification succeeds.
- The `PROGRESS` block lists paths changed, tests run, commit hash if committed, and known follow-ups.
- A task-sized git commit exists unless the user explicitly requested no commit or the repository is not initialized.

## Documentation Discipline

Documentation must be built as the system is built, not deferred to the end:

- Public packages need concise package docs or README coverage when their behavior becomes usable.
- Public functions, structs, and interfaces need comments when exported API semantics are not obvious from names.
- Integration mappers must document accepted input shape, required fields, optional fields, and redaction/access expectations.
- HTTP and stdio surfaces must document request/response JSON and red-plan behavior before being marked complete.
- Memory behavior must document when write-back is allowed, what is stored, re-gating rules, and reuse rejection reasons.
- If docs and implementation disagree, fix both before marking the task complete.
- Do not add aspirational docs for features not implemented; use `TODO(spec section, task ID)` only when a later task owns the work.

## Git Discipline

Use git as the durable circular-development boundary:

- Make one logical commit per completed task or coherent subtask.
- Do not commit unrelated user changes. If unrelated dirty files exist, leave them untouched and mention them in the `PROGRESS` notes only if they affect verification.
- Before committing, review `git status --short` and `git diff` to ensure only intended files are staged.
- Commit code, tests, docs, fixtures, and the `DEVPLAN.md` progress update together when they belong to the same task.
- Use concise imperative commit messages prefixed with the task ID, for example `T03.2 implement personalized pagerank`.
- Do not amend, rebase, force-push, or rewrite history unless the user explicitly asks.
- If hooks or tests fail, do not mark the task `[x]`; fix the issue or mark `[!]` with evidence.
- After committing, record the short commit hash in the task `PROGRESS` block.

## Progress Marker Protocol

Use these exact status markers in task checkboxes:

- `[ ]` not started
- `[/]` in progress
- `[x]` complete
- `[!]` blocked
- `[?]` needs human decision

Agent update format:

```md
<!-- PROGRESS:AGENT_ID:TASK_ID:YYYY-MM-DDTHH:MMZ -->
Status: [/]
Owner: agent-name-or-session
Evidence: paths changed, tests run, key output
Spec refs: spec.md sections implemented or checked
Docs: docs/examples/README/package comments updated, or "not needed: reason"
Notes: blockers, follow-ups, or "none"
Commit: short hash or "not committed"
<!-- /PROGRESS -->
```

Rules for subagents:

- Claim exactly one task at a time by changing its checkbox to `[/]` and adding a `PROGRESS` block below it.
- Do not edit another agent's active task unless its marker is `[!]`, `[?]`, or the user explicitly asks you to take over.
- Keep changes scoped to the files listed in the task unless the task says otherwise.
- Before marking `[x]`, run the verification commands listed for the task and record the result.
- If you discover a spec conflict, mark the task `[?]`, document the question, and continue with independent tasks.
- Never weaken existing CLI behavior or existing tests.
- Keep the newest `PROGRESS` block closest to the task heading; older progress blocks may remain below for audit history.
- If a task is split, keep the original anchor and add new `<!-- TASK:... -->` anchors so future agents can search and resume.

## Build Principles

- Core must be deterministic, Go-first, and stdlib-first.
- CitewiseRAG assembles context after upstream retrieval; it must not implement embeddings, vector search, chunking, LLM extraction, LLM reranking, or community detection.
- Access control is a hard gate before scoring, packing, provenance expansion, and memory load.
- Existing `pkg/citewise` CLI behavior is the compatibility anchor.
- New packages should expose small interfaces and pure functions so agents can work independently.
- Output JSON must be deterministic: stable ordering, stable float rounding where specified, stable suppression order.

## Reuse And Dependency Decisions

Research-backed decisions for avoiding unnecessary custom work:

- Microsoft GraphRAG: consume its JSON/parquet output tables and query-mode concepts; do not reimplement indexing, entity extraction, Leiden clustering, community summaries, or DRIFT loops. Source: Microsoft GraphRAG docs describe indexing text units, entities, relationships, claims, community hierarchy/summaries, and global/local/DRIFT/basic query modes.
- LightRAG: consume local/global results and retrieval modes; do not embed or port LightRAG core. Source: LightRAG project exposes dual-level graph-aware retrieval, server/API usage, reranker support, and citation/retrieved-context features.
- `github.com/parquet-go/parquet-go`: use only for optional GraphRAG parquet imports behind `graphrag_parquet`; core JSON import remains dependency-free. Source: parquet-go is a Go module for high-performance parquet read/write, Go 1.22+, Apache-2.0, pre-v1 compatibility caveat.
- `gonum.org/v1/gonum/graph/network`: do not use in MVP core; consider behind future `graphx` build tag if graph algorithms outgrow custom PPR. Source: Gonum provides PageRank, HITS, betweenness, closeness, etc., but pulling it into core is unnecessary for the small personalized PPR required here.
- `github.com/google/go-cmp/cmp`: allow as test-only for golden/table diff readability.
- Reject core dependencies on Bleve, Redis, Neo4j, web routers, JSON streaming libraries, and LLM SDKs because retrieval, external stores, HTTP routing, JSON decoding, and LLM calls are outside core scope.

## Milestones

<!-- ANCHOR:M00-REPO-BOOTSTRAP -->
### M00 Repo Bootstrap

Goal: create a compilable Go module and preserve existing Citewise compatibility surface.

Exit gate: `go test ./...` runs and either passes or fails only for explicitly documented missing implementation tests.

<!-- ANCHOR:M01-DATA-MODEL -->
### M01 Data Model And Conversion

Goal: implement `pkg/ragnode` types, edge ontology normalization, analysis construction, and conversion to/from `pkg/citewise`.

Exit gate: ragnode unit tests cover conversion, token estimates, edge aliases, and deterministic analysis maps.

<!-- ANCHOR:M02-ACCESS-PROVENANCE -->
### M02 Access And Provenance

Goal: implement hard access gates and safe provenance builders/redaction.

Exit gate: tests prove unauthorized text/source trails cannot appear in ranked sets, plans, critique summaries, or memory loads.

<!-- ANCHOR:M03-RANKER -->
### M03 Ranker

Goal: implement deterministic scoring, authority, PPR, token budget fit, diversity, rationale, and suppressions.

Exit gate: ranker tests cover exact scoring factors, PPR convergence/fallback, query-type modifiers, duplicate/stale/low-trust penalties.

<!-- ANCHOR:M04-PACKER-HYGIENE-ROUTER -->
### M04 Packer, Hygiene, Router

Goal: produce context plans with slot policies, lost-in-the-middle ordering, corrective signals, and deterministic routing.

Exit gate: smoke tests cover green/yellow/red cases and all router decision branches.

<!-- ANCHOR:M05-INTEGRATIONS-MEMORY -->
### M05 Integrations And Memory

Goal: import upstream candidate formats and persist/reuse prior plans safely.

Exit gate: integration fixtures for GraphRAG JSON, LightRAG, hybrid RRF, optional parquet, and JSONL memory pass.

<!-- ANCHOR:M06-API-PIPELINE -->
### M06 Pipeline And HTTP API

Goal: expose top-level `pkg/rag` orchestration and optional `cmd/serve` endpoints.

Exit gate: `/health`, `/router`, `/rank`, `/pack`, `/hygiene` compile and pass handler tests using stdlib `net/http/httptest`.

<!-- ANCHOR:M07-HARDENING-RELEASE -->
### M07 Hardening And Release Readiness

Goal: property tests, golden outputs, docs, examples, and compatibility checks.

Exit gate: `go test ./...`, existing CLI regression tests, race-safe memory tests, and deterministic golden tests pass.

## Work Packets

<!-- TASK:T00.1-BOOTSTRAP-MODULE -->
### [x] T00.1 Bootstrap Go Module

<!-- PROGRESS:opencode:T00.1-BOOTSTRAP-MODULE:2026-05-14T17:39Z -->
Status: [x]
Owner: opencode
Evidence: added go.mod, .gitignore, package directory skeletons, and testdata placeholders; `go test ./...` passed
Spec refs: spec.md section 3.1 checked
Docs: package doc placeholders added for initial packages only
Notes: module path github.com/mojomast/citewiseussy; Go go1.24.4 linux/amd64; git was already initialized with no commits
Commit: 853b335
<!-- /PROGRESS -->

Scope: `go.mod`, initial directories, placeholder package docs only where needed.

Dependencies: none.

Implementation:

- Initialize module as `github.com/mojomast/citewiseussy` unless the existing repository declares a different module.
- If the workspace is not already a git repository, initialize git before implementation work so task commits can be recorded.
- Add a minimal `.gitignore` for Go build/test artifacts and local memory files such as `citewiserag_memory.jsonl`.
- Create package directories from `spec.md` section 3.1.
- Keep placeholders minimal; do not add fake behavior that hides missing implementation.

Verification:

- Run `go test ./...`.
- Record module path and Go version.
- Record whether git was initialized and the initial commit hash if created.

<!-- TASK:T00.2-CITEWISE-COMPAT -->
### [x] T00.2 Establish Citewise Compatibility Anchor

<!-- PROGRESS:opencode:T00.2-CITEWISE-COMPAT:2026-05-14T17:52Z -->
Status: [x]
Owner: opencode
Evidence: imported upstream github.com/mojomast/citewiseussy main.go, pkg/citewise CLI/types/engine, and regression tests; `go test ./pkg/citewise` passed; `go test ./...` passed
Spec refs: spec.md sections 3.1, 3.2, 11, and 12.3 checked
Docs: upstream CLI help and existing README-level docs preserved by behavior; no new public API beyond imported compatibility anchor
Notes: existing code sourced from temporary clone of mojomast/citewiseussy after user guidance; old commands roles, score, queue, explain, hygiene, export preserved
Commit: a184869
<!-- /PROGRESS -->

Scope: `pkg/citewise`, `main.go`, existing CLI tests if present.

Dependencies: T00.1.

Implementation:

- If existing `github.com/mojomast/citewiseussy` code is present, preserve it unchanged except for mechanical relocation if required.
- If only `spec.md` is present, create the smallest compatible `pkg/citewise` surface needed by RAG packages: `Item`, `Edge`, `Goal`, `Analysis`, classifier hooks, centrality, duplicate/hygiene/readiness stubs backed by tests.
- Do not change old command names: `roles`, `score`, `queue`, `explain`, `hygiene`, `export`.

Verification:

- Run `go test ./pkg/citewise`.
- Run old CLI fixture tests if fixtures exist.

<!-- TASK:T01.1-RAGNODE-TYPES -->
### [x] T01.1 Implement RAGNode Types

<!-- PROGRESS:opencode:T01.1-RAGNODE-TYPES:2026-05-14T17:58Z -->
Status: [x]
Owner: opencode
Evidence: added pkg/ragnode node/candidate/analysis types and tests; `go test ./pkg/ragnode` passed; `go test ./...` passed
Spec refs: spec.md section 3.2 implemented for T01.1 scope
Docs: exported type names match spec; no additional docs needed yet because behavior is limited to data shapes and token estimation
Notes: plain Edge struct added for CandidateSet/RAGAnalysis compilation; normalization remains T01.2
Commit: e2b249f
<!-- /PROGRESS -->

Scope: `pkg/ragnode/node.go`, `candidate.go`, `analysis.go`.

Dependencies: T00.2.

Implementation:

- Implement all structs and constants from `spec.md` section 3.2.
- Embed `citewise.Item` in `RAGNode`; do not fork old item fields.
- Add token estimation helper: `ceil(len(Text)/4)` when `TokenCount == 0`.
- Keep helper exported only if needed by ranker/packer tests.

Verification:

- Tests for JSON tags, zero-value safety, and token estimate behavior.

<!-- TASK:T01.2-EDGE-ONTOLOGY -->
### [x] T01.2 Implement Edge Ontology Normalization

<!-- PROGRESS:opencode:T01.2-EDGE-ONTOLOGY:2026-05-14T18:02Z -->
Status: [x]
Owner: opencode
Evidence: added edge ontology constants, alias normalization, weights, raw type note preservation, and tests; `go test ./pkg/ragnode` passed; `go test ./...` passed
Spec refs: spec.md section 6 implemented
Docs: not needed: public constants/functions mirror spec edge ontology
Notes: unknown edge types normalize to related-to and can be detected with EdgeTypeKnown/NormalizeEdge.Known for later ranker rationale
Commit: 8a80294
<!-- /PROGRESS -->

Scope: `pkg/ragnode/edge.go`.

Dependencies: T01.1.

Implementation:

- Normalize edge types to lowercase kebab-case.
- Implement alias table from `spec.md` section 6.
- Preserve raw type in `Note` or `Attributes["raw_type"]` if an attributes field is later added.
- Unknown edge types must normalize to `related-to` for weighting and produce warning rationale data for ranker.

Verification:

- Table tests for every canonical edge type and alias.
- Unknown-type test verifies fallback behavior.

<!-- TASK:T01.3-CONVERSION-ANALYSIS -->
### [x] T01.3 Implement Conversion And RAGAnalysis Builder

<!-- PROGRESS:opencode:T01.3-CONVERSION-ANALYSIS:2026-05-14T18:09Z -->
Status: [x]
Owner: opencode
Evidence: added pkg/ragnode convert.go, package docs, and conversion/analysis tests; `go test ./pkg/ragnode` passed; `go test ./...` passed
Spec refs: spec.md sections 3.2, 6, 12.3, 12.4 implemented for T01.3 scope
Docs: pkg/ragnode/doc.go documents old backlog conversion, new RAG JSON parsing, and CLI parser separation
Notes: none
Commit: not committed
<!-- /PROGRESS -->

Scope: `pkg/ragnode/convert.go`, `analysis.go`, RAG JSON adapter tests.

Dependencies: T01.1, T01.2, T00.2.

Implementation:

- `RAGNode.ToItem()` must produce a valid `citewise.Item`.
- When converting existing backlog JSON to `RAGNode`, apply section 12.3 defaults exactly: `Text=Item.Notes`, `ChunkType=document`, estimated `TokenCount`, empty `Version`, `Sensitivity=internal`, nil `ApprovedBy`, empty `CommunityID`, and empty `ContextPrefix`.
- New RAG JSON files may include RAG-only fields; old CLI behavior must remain unchanged and ignore unknown RAG fields through the adapter path rather than by weakening the old parser.
- `Edge.ToCitewiseEdge()` must normalize edge type and preserve confidence.
- Build `RAGAnalysis` maps with deterministic ordering where slices are emitted.
- Validate candidate node IDs and return actionable errors.

Verification:

- Tests for old backlog defaults from `spec.md` section 12.3.
- Tests that new RAG JSON fields round-trip through the RAG adapter without changing old CLI parsing expectations.
- Tests for missing candidate node ID errors.

<!-- TASK:T02.1-ACCESS-CONTROLLER -->
### [ ] T02.1 Implement Access Controller

Scope: `pkg/access/access.go`, `approved.go`.

Dependencies: T01.1.

Implementation:

- Implement clearance ordinal rules exactly.
- Enforce `ApprovedBy` trusted approver rule for confidential/restricted nodes.
- Enforce agentic approval rule for permission-record, decision, and procedure unless `allow_unapproved_agentic_nodes=true`.
- `RedactNode` must remove `Text`, `Title`, `URL`, `Source`, `SourceTrail`, and any sensitive fields not required for audit.

Verification:

- Table tests for all clearance combinations.
- Tests for trusted approvers, unapproved agentic nodes, edge access, and redaction.

<!-- TASK:T02.2-PROVENANCE -->
### [ ] T02.2 Implement Provenance Builders

Scope: `pkg/provenance/trail.go`, `redact.go`.

Dependencies: T01.1, T02.1.

Implementation:

- Build `SourceRef` from node metadata.
- Build `SourceTrail` using retrieved, edge, slot-required, community-member-of, and decision-basis rules.
- Redact unauthorized provenance before it leaves package boundaries.

Verification:

- Tests for direct retrieval hop, required slot synthetic hop, community summary path, decision basis path, and redaction.

<!-- TASK:T03.1-AUTHORITY-TOKEN-DIVERSITY -->
### [ ] T03.1 Implement Authority, Token Budget, Diversity Scorers

Scope: `pkg/ranker/authority.go`, `token_budget.go`, `diversity.go`.

Dependencies: T01.3.

Implementation:

- Authority formula from section 3.4 with chunk type priors.
- TokenBudgetFit formula from section 3.4, including normalized token density.
- Diversity score by source and community; deterministic and bounded 0..1.

Verification:

- Exact numeric table tests with clamp behavior and token estimate rationale.

<!-- TASK:T03.2-PPR -->
### [ ] T03.2 Implement Custom Personalized PageRank

Scope: `pkg/ranker/ppr.go`.

Dependencies: T01.2, T01.3.

Implementation:

- Implement custom PPR with alpha `0.15`, max iterations `100`, tolerance `1e-8`.
- Seed from upstream `Candidate.QueryRelevance`.
- Use edge transition weights from ontology table.
- If graph has fewer than two usable edges, return zeroes and let ranker use weighted centrality only.
- Do not add Gonum in MVP core.

Verification:

- Tests for convergence, disconnected graph, dangling nodes, fewer-than-two-edge fallback, deterministic output.

<!-- TASK:T03.3-RANKER -->
### [ ] T03.3 Implement Ranker Scoring And Suppression

Scope: `pkg/ranker/ranker.go`, `scorer.go`, `explain.go`.

Dependencies: T02.1, T03.1, T03.2.

Implementation:

- Apply access gating before scoring.
- Compute score formula from section 4.1 and modifiers from section 4.3.
- Include rationale for inclusion/exclusion, token estimate fallback, unknown edge fallback, penalties, and access suppressions.
- Return `RankedSet` with stable ordering by total descending then node ID ascending.

Verification:

- Tests for formula weights, access suppression redaction, modifier application, duplicate/stale/low-trust penalties, stable sorting.

<!-- TASK:T04.1-PACKER-SLOTS -->
### [ ] T04.1 Implement Slot Policies And ContextPlan

Scope: `pkg/packer/querytype.go`, `plan.go`, `slots.go`, `packer.go`.

Dependencies: T03.3, T02.2.

Implementation:

- Implement QueryType, SlotType, SlotPosition, HygieneSignal, ContextPlan structs.
- Fill slots according to section 5.1.
- Enforce duplicate, stale, required slot, low score, and Agentic fail-closed rules.
- Enforce Agentic hard failures from section 5.2: no permitted permission record, unapproved permission record, missing required procedure, missing source trail for any required slot, required node suppressed by access control, or stale-only policy.
- Never include access-suppressed nodes.

Verification:

- Table tests for every QueryType slot requirement and every Agentic hard failure.
- Smoke tests A-E from section 13.4.

<!-- TASK:T04.2-LOST-IN-MIDDLE -->
### [ ] T04.2 Implement Lost-In-The-Middle Ordering And Budget Trimming

Scope: `pkg/packer/ordering.go`, `packer.go`.

Dependencies: T04.1.

Implementation:

- Implement exact ordering algorithm from section 9.2.
- Stable-sort inside front/middle/back bands by score descending then node ID ascending.
- Trim lowest-score middle support first, then optional overview, then optional bridge.
- Never remove required slots unless returning red.

Verification:

- Tests for ordering bands, tie-breaking, budget trimming order, required-slot over-budget behavior.

<!-- TASK:T04.3-HYGIENE -->
### [ ] T04.3 Implement Hygiene Analyzer

Scope: `pkg/hygiene/hygiene.go`, `suggestions.go`, `signal.go`.

Dependencies: T01.3, T04.1.

Implementation:

- Wrap existing `citewise.HygieneReport` if available.
- Implement missing-edge heuristics from section 3.6.
- Implement HygieneScore and CorrectiveSignal thresholds.
- Return retrieval target suggestions for red signals if the caller disallows degraded plans.

Verification:

- Tests for each heuristic and threshold boundary.

<!-- TASK:T04.4-ROUTER -->
### [ ] T04.4 Implement Deterministic Query Router

Scope: `pkg/router/router.go`, `heuristics.go`, `metadata.go`.

Dependencies: T04.1.

Implementation:

- Implement decision tree from section 3.7 in exact order.
- Use deterministic tokenization and case normalization.
- Return reasons for every matched rule.

Verification:

- Branch coverage tests for every rule and fallback.

<!-- TASK:T05.1-GRAPHRAG-JSON -->
### [ ] T05.1 Implement GraphRAG JSON Mapper

Scope: `pkg/integrations/graphrag/mapper.go`, `testdata/graphrag_minimal`.

Dependencies: T01.3.

Implementation:

- Map documents, text_units, entities, relationships, community_reports, communities, and covariates from section 7.1.
- Accept JSON exports by default.
- Store `human_readable_id`, period, and community report rank as specified.

Verification:

- Fixture test: community report imports as `community-summary` and role overview.

<!-- TASK:T05.2-GRAPHRAG-PARQUET-OPTIONAL -->
### [ ] T05.2 Implement Optional GraphRAG Parquet Reader

Scope: `pkg/integrations/graphrag/parquet.go`.

Dependencies: T05.1.

Implementation:

- Add build tag `graphrag_parquet`.
- Use `github.com/parquet-go/parquet-go` only in files guarded by that build tag.
- Keep JSON mapper dependency-free.

Verification:

- Run `go test ./...` without tag and verify no parquet dependency is required by core.
- Run `go test -tags graphrag_parquet ./pkg/integrations/graphrag`.

<!-- TASK:T05.3-LIGHTRAG-HYBRID -->
### [ ] T05.3 Implement LightRAG, Hybrid, And Reranker Handoff Mappers

Scope: `pkg/integrations/lightrag/mapper.go`, `pkg/integrations/hybrid/schema.go`, cross-encoder handoff fixtures.

Dependencies: T01.3.

Implementation:

- Map LightRAG local, global, and relationship/path results from section 7.2.
- Map hybrid BM25/dense/graph RRF schema from section 7.3.
- Accept cross-encoder/ColBERT-style reranker handoff from section 7.4 as relevance input only; preserve `RerankerScore` and method scores without allowing reranking to override slot diversity, counterpoint inclusion, access gates, or context ordering.
- Preserve `RetrievalMode`, method scores, ranks, weights, and normalized `QueryRelevance`.
- Document that upstream rerankers must run after coarse ACL and must not receive unauthorized text; CitewiseRAG still reapplies hard access gating.

Verification:

- Fixture tests for LightRAG mode preservation, RRF query relevance mapping, and cross-encoder relevance handoff.

<!-- TASK:T05.4-MEMORY -->
### [ ] T05.4 Implement File-Backed Memory Store

Scope: `pkg/memory/interface.go`, `file_store.go`, `redact.go`.

Dependencies: T02.1, T04.1.

Implementation:

- Store one JSON object per line in `citewiserag_memory.jsonl`.
- Compute `PlanHash = sha256(QueryType + EvidencePath + node versions)`.
- Populate `WriteBackPayload` with the section 10.1 shape: `plan_hash`, `query_id`, `query_type`, `topics`, `evidence_path`, `node_versions`, `hygiene_signal`, and `created_at`.
- Enforce write-back policy from section 7.6: never write red plans, and only write after the caller accepts the plan or the agent completes successfully.
- Implement atomic append with `O_APPEND` and `Sync` or temp-file swap where appropriate.
- Re-run access gating on load before returning any slot.
- Implement deterministic topic Jaccard for `SimilarPriorPlans` and enforce reuse policy from section 10.2: matching required node versions, no superseded nodes, current caller access, topic Jaccard >= 0.70, and compatible `QueryType` rules.

Verification:

- Tests for store/load, hash stability, write-back payload shape, red-plan rejection, similar prior plans, reuse rejection reasons, redacted load, and concurrent append if feasible.

<!-- TASK:T06.1-RAG-PIPELINE -->
### [ ] T06.1 Implement Top-Level RAG Pipeline

Scope: `pkg/rag/pipeline.go`, `interfaces.go`.

Dependencies: T03.3, T04.1, T04.3, T04.4, T05.4.

Implementation:

- Orchestrate validate -> analyze -> access -> classify -> rank -> hygiene -> pack.
- Expose default constructors for ranker, packer, router, hygiene analyzer, and memory store.
- Expose the section 12.2 library flow: `rag.Analyze(candidateSet)`, `rag.DefaultRanker().Rank(...)`, and `rag.DefaultPacker().Pack(...)`.
- Return typed errors for invalid candidates, access-denied-only results, and red corrective signal.

Verification:

- End-to-end tests from candidate set to context plan.

<!-- TASK:T06.2-HTTP-SERVER -->
### [ ] T06.2 Implement Optional HTTP Server

Scope: `cmd/serve/main.go`, `handlers.go`, `schema.go`.

Dependencies: T06.1.

Implementation:

- Use only `net/http`, `encoding/json`, `context`, and `log/slog` unless a strong need is documented.
- Implement `/health`, `/router`, `/rank`, `/pack`, `/hygiene`.
- Implement request/response DTOs from section 3.9, including `/rank` candidate intake and `/pack` `ContextPlan` response shape.
- `/health` must return service status and graph hygiene signal.
- Decide red plan response behavior from open question; default to HTTP 200 with `HygieneSignal=red` and corrective detail unless user decides otherwise.
- Do not make server required for library builds.

Verification:

- Handler tests with `httptest` for success, bad JSON, access suppression, and red Agentic plan.

<!-- TASK:T06.3-STDIO-INTEGRATION -->
### [ ] T06.3 Implement Optional Stdio JSON Integration

Scope: `cmd/serve` or a small sibling command if HTTP and stdio need separate entrypoints.

Dependencies: T06.1.

Implementation:

- Provide a stdio JSON request/response path for the same `/router`, `/rank`, `/pack`, and `/hygiene` operations if the release target requires non-HTTP agent integration.
- Reuse the same DTOs and pipeline as HTTP; do not create a second behavior surface.
- Keep stdio optional and dependency-free.

Verification:

- Golden stdin/stdout tests for one route request and one pack request.

<!-- TASK:T07.1-PROPERTY-TESTS -->
### [ ] T07.1 Implement Property And Invariant Tests

Scope: package test files across `pkg/*`.

Dependencies: T02.1, T03.3, T04.1, T05.4.

Implementation:

- Use `testing/quick` for invariants in section 13.2.
- Focus on unauthorized text absence, budget bounds, duplicate limits, deterministic JSON, and provenance coverage.

Verification:

- Run `go test ./...` repeatedly enough to catch nondeterminism.

<!-- TASK:T07.2-GOLDEN-AND-DOCS -->
### [ ] T07.2 Add Golden Fixtures, Docs, And Examples

Scope: `testdata/*`, `README.md` or docs file, example tests.

Dependencies: all M01-M06 tasks.

Implementation:

- Add minimal examples for `rag.Analyze`, `DefaultRanker().Rank`, and `DefaultPacker().Pack`.
- Add golden JSON for an Agentic green plan, Agentic red plan, Temporal stale red plan, and Factual green plan.
- Document non-goals, upstream handoff expectations, and downstream agent provenance obligations from section 7.5, including required answer/action-log fields.
- Document memory write-back acceptance timing, reuse constraints, and caller-owned Redis/Neo4j/Mem0/Zep adapter boundaries.

Verification:

- Golden tests pass byte-for-byte.
- Examples compile under `go test ./...`.

<!-- TASK:T07.3-RELEASE-GATE -->
### [ ] T07.3 Release Gate And Compatibility Audit

Scope: whole repository.

Dependencies: all tasks.

Implementation:

- Run full test suite.
- Run existing CLI commands against old fixtures.
- Verify no forbidden core dependencies were introduced.
- Verify optional parquet dependency is build-tag isolated.
- Verify no LLM, vector DB, Redis, Neo4j, or web router dependency is in core.

Verification:

- `go test ./...`
- `go test -tags graphrag_parquet ./...` if optional parquet implemented.
- Record `go list -deps ./...` audit notes.

## Parallelization Map

Safe early parallel work after T00.1/T00.2:

- Agent A: T01.1, T01.2, T01.3
- Agent B: T02.1, then T02.2
- Agent C: T04.4 router, because it only depends on query type constants once T04.1 constants exist or can be stubbed briefly
- Agent D: T05.1/T05.3 mapper fixture design, using `ragnode` structs once T01.1 lands

Safe parallel work after M01/M02:

- Ranker scorers T03.1 and PPR T03.2 can be built independently.
- Hygiene T04.3 and router T04.4 can proceed independently.
- Memory T05.4 can proceed once access and plan structs exist.

Serial dependencies:

- T03.3 must wait for T03.1, T03.2, and T02.1.
- T04.1 must wait for T03.3 and provenance.
- T06.1 must wait for ranker, packer, hygiene, router, and memory.
- T07.3 must wait for all implementation tasks.

## Open Decisions To Resolve Early

<!-- DECISION:D01-APPROVEDBY -->
### [?] D01 ApprovedBy Semantics

Default: treat `ApprovedBy` as source approval, not viewer allow-list. `Sensitivity` and future ABAC rules handle visibility.

Needed before: T02.1, T03.1.

<!-- DECISION:D02-ACCESS-ATTRIBUTES -->
### [?] D02 Caller Attribute Rules

Default: MVP supports clearance, groups, trusted approvers, and raw attributes only for specified flags. Do not implement region/account/department ABAC until required.

Needed before: T02.1.

<!-- DECISION:D03-GRAPHRAG-VERSION -->
### [?] D03 First GraphRAG Schema Version

Default: support current documented JSON exports by field names in fixtures; make mapper tolerant of missing optional fields.

Needed before: T05.1.

<!-- DECISION:D04-RED-PACK-HTTP -->
### [?] D04 Red Plan HTTP Behavior

Default: `/pack` returns HTTP 200 with `HygieneSignal=red` and corrective detail. Use HTTP 409 only if API clients need transport-level failure.

Needed before: T06.2.

<!-- DECISION:D05-CLI-RAG-COMMANDS -->
### [?] D05 RAG CLI Commands

Default: MVP exposes library and optional HTTP only. Add `citewise rag pack` and `citewise rag route` only after compatibility tests are locked.

Needed before: post-MVP unless user requests CLI.

<!-- DECISION:D06-TOKEN-COUNTING -->
### [?] D06 Token Counting Source

Default: upstream tokenizer-provided counts are preferred; `ceil(len(Text)/4)` remains the deterministic MVP fallback and must be marked in scoring rationale.

Needed before: T01.1, T03.1.

<!-- DECISION:D07-TABLE-PROVENANCE -->
### [?] D07 Table Provenance Requirements

Default: `Locator.TableID`, `RowStart`, and `RowEnd` are optional unless present in upstream data; metric-bearing node tests should preserve them when supplied.

Needed before: T02.2, T07.2.

<!-- DECISION:D08-MEMORY-ACCEPTANCE -->
### [?] D08 Memory Acceptance Signal

Default: FileStore exposes storage primitives, but automatic write-back only occurs after explicit caller acceptance or successful agent completion; generation alone is insufficient.

Needed before: T05.4, T06.1.

<!-- DECISION:D09-CALIBRATION-SET -->
### [?] D09 Scoring Calibration Dataset

Default: implement specified constants first and add golden fixtures; tune only after a user-provided calibration set exists.

Needed before: T07.3.

<!-- DECISION:D10-PPR-SCALE-LIMITS -->
### [?] D10 PPR Scale Limits

Default: MVP uses custom in-core PPR for all tests; document observed performance, and consider optional `graphx`/Gonum only if scale tests show a need.

Needed before: T03.2, T07.3.

<!-- DECISION:D11-SUPPRESSION-AUDIT -->
### [?] D11 Suppression Audit Level

Default: ordinary responses expose redacted access-control node IDs, reasons, and details only; no text, title, URL, or source trail. Admin-specific expanded audit is out of MVP scope.

Needed before: T02.1, T03.3, T04.1.

<!-- DECISION:D12-AGENTIC-FAIL-CLOSED -->
### [?] D12 Agentic Fail-Closed Exceptions

Default: Agentic mode may not proceed without a permitted permission record unless a future domain-specific policy explicitly changes the spec.

Needed before: T04.1, T06.2.

<!-- DECISION:D13-COMMUNITY-SUMMARY-TRUST -->
### [?] D13 Community Summary Trust

Default: community summaries use the specified chunk-type authority prior and may inherit source trust when upstream provides it; no extra generated-content penalty unless calibration shows one is needed.

Needed before: T03.1, T05.1.

<!-- DECISION:D14-CROSS-ENCODER-LEAKAGE -->
### [?] D14 Cross-Encoder Leakage Boundary

Default: upstream rerankers are responsible for coarse ACL before reranking; CitewiseRAG re-gates every node and documents this boundary. Add a pre-rerank redaction utility only if callers require it.

Needed before: T05.3, T07.2.

## Definition Of Done

- Existing CLI behavior remains unchanged.
- Core packages compile without optional dependencies.
- Access-control invariant is tested at rank, pack, provenance, critique, and memory boundaries.
- Context plans are deterministic and auditable.
- Agentic mode fails closed on missing permission, unapproved controlling nodes, missing procedure, stale-only policy, or missing source trail for required slots.
- GraphRAG, LightRAG, hybrid RRF, and cross-encoder outputs are accepted as handoff schemas rather than reimplemented.
- `go test ./...` passes.
