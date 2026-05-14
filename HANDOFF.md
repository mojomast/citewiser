# CitewiseRAG Handoff

This file is the first thing a new agent should read. It summarizes the current development state and gives explicit next instructions. Update this file at the end of every completed or paused work slice, alongside `DEVPLAN.md` progress blocks and any docs touched by the slice.

## Required Startup Sequence

1. Read `HANDOFF.md` first.
2. Read `spec.md` for product truth. Current ranker refs: `spec.md` around lines 296-402 and 743-789.
3. Read `DEVPLAN.md` for task state and sequencing. Resume workflow: around lines 5-16. Documentation discipline: around lines 30-40. Progress marker protocol: around lines 55-88. Next tasks: around lines 374-427.
4. Check git before editing: `git status --short`, `git diff`, `git log --oneline -5`.
5. Identify the earliest unblocked `[ ]` task in `DEVPLAN.md`. Do not take over `[/]` work unless stale or explicitly approved.
6. Claim exactly one task by changing its checkbox to `[/]` and adding a new `PROGRESS` block immediately under that task.
7. Implement in a circular slice: code, tests, docs, verification, `DEVPLAN.md`, `HANDOFF.md`, then commit if the user asked for normal development/commits or the dev plan task requires it.

## Current State

Completed task commits:

- `T00.1` bootstrap Go module: `853b335`
- `T00.2` Citewise compatibility anchor: `a184869`
- `T01.1` RAGNode types: `e2b249f`
- `T01.2` edge ontology normalization: `8a80294`
- `T01.3` conversion and analysis builder: `eaa365a`
- `T02.1` access controller: `c43759e`
- `T02.2` provenance builders/redaction: `cfde974`

Latest ledger-only hash updates exist after several task commits; use `git log --oneline -10` for exact current HEAD.

The working tree was clean when this handoff was created.

## Next Work

Earliest unblocked task: `T03.1 Implement Authority, Token Budget, Diversity Scorers`.

Primary files to create/update:

- `pkg/ranker/doc.go`
- `pkg/ranker/authority.go`
- `pkg/ranker/token_budget.go`
- `pkg/ranker/diversity.go`
- `pkg/ranker/*_test.go`
- `DEVPLAN.md`
- `HANDOFF.md`

Relevant `DEVPLAN.md` refs:

- `T03.1`: around lines 374-390.
- `T03.2`: around lines 391-409, can be worked independently after `T01.3`.
- `T03.3`: around lines 410-427, must wait for `T02.1`, `T03.1`, and `T03.2`.
- Parallelization map: around lines 687-707.

Relevant `spec.md` refs:

- Ranker types/interfaces: around lines 296-360.
- `AuthorityScore` formula and chunk priors: around lines 361-383.
- `TokenBudgetFit` formula: around lines 384-394.
- PPR context for later `T03.2`: around lines 396-402.
- Default score formula for later `T03.3`: around lines 745-762.
- Query-type modifiers for later `T03.3`: around lines 777-789.
- Unit-test strategy: around lines 1137-1148.
- Dependency guidance: around lines 1203-1217. Do not add Gonum or other core dependencies for MVP ranker work.

Concrete `T03.1` instructions:

1. Claim `T03.1` in `DEVPLAN.md` with a `PROGRESS` block.
2. Add package docs explaining deterministic scorer helpers and the token-estimate fallback rationale.
3. Implement authority scoring from `spec.md` lines 361-383:
   - Clamp to `0..1`.
   - `0.35 * Trust`.
   - `0.20 * normalizedIncomingCites` using incoming `cites` edges, normalized by the max incoming cite count in the analysis.
   - `0.20 * approvedByScore`, likely `1` when `ApprovedBy` is non-empty, else `0`.
   - `0.15 * chunkTypeAuthority` using the exact prior table.
   - `0.10 * versionCurrentness`, likely `1` when `Version` is non-empty and node is not superseded, lower/zero for superseded or missing version. If interpretation is unclear, document the chosen deterministic MVP rule in package docs and tests.
4. Implement token budget fit from `spec.md` lines 384-392:
   - Use `RAGNode.EffectiveTokenCount()` so `ceil(len(Text)/4)` applies when `TokenCount == 0`.
   - Compute token density as `AuthorityScore / max(1, log2(TokenCount + 2))`.
   - Normalize density deterministically across nodes in the current analysis, or document/test a bounded single-node fallback if implemented as a stateless helper.
   - Apply length-fit thresholds exactly: `<= 0.08`, `<= 0.15`, `<= 0.25`, otherwise `0.25`.
5. Implement diversity scoring by source and community, deterministic and bounded `0..1`.
6. Add exact numeric table tests covering clamp behavior, chunk priors, token estimate fallback, and diversity source/community penalties.
7. Run `go test ./pkg/ranker` and `go test ./...`.
8. Update `DEVPLAN.md` `T03.1` progress with changed paths, tests, docs, notes, and commit hash after commit.
9. Update this `HANDOFF.md` so the next agent sees the new current state and next task.

Parallel option after claiming only one task yourself:

- A subagent may research `T03.2` PPR requirements without writing files while you implement `T03.1`.
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
