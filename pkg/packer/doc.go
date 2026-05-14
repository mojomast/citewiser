// Package packer assembles ranked RAG nodes into deterministic context plans.
// It maps ranked evidence into query-type-specific slots, suppresses stale,
// duplicate, and access-denied nodes, builds provenance-bearing context slots,
// and applies lost-in-the-middle ordering and budget trimming. ContextPlan also
// exposes a structural PlanHash and suppression reason breakdown for audit logs.
package packer
