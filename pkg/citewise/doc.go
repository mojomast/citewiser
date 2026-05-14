// Package citewise implements the legacy reading-backlog triage engine.
//
// It is retained as a compatibility anchor for the citewise CLI and for test fixtures.
// New development should target pkg/ragnode, pkg/ranker, pkg/packer, and pkg/rag instead.
//
// The engine scores backlog items on five dimensions:
//   - GoalFit:       keyword overlap between item topics/notes and the active reading goal.
//   - Centrality:    weighted in/out-degree in the prerequisite/citation graph.
//   - Readiness:     difficulty alignment with the reader's familiarity level and whether
//     prerequisite items have been read.
//   - Freshness:     recency of the publication year, with a classic-content exception.
//   - EnergyTimeFit: match between the item's length/energy level and the reader's budget.
//
// Scoring weights are named constants (wGoalFit, wCentrality, etc.) in engine.go and
// represent policy defaults. They are not user-configurable in the current release.
package citewise
