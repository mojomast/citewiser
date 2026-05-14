// Package ranker provides deterministic scoring helpers for retrieved RAG
// candidates.
//
// The helpers are intentionally deterministic and stdlib-only. Authority
// scoring follows the spec formula directly. Token budget fit uses
// RAGNode.EffectiveTokenCount, so missing upstream token counts fall back to
// ceil(len(Text)/4); density is normalized across the current analysis, with a
// bounded single-node fallback for callers that use the stateless helper.
// Personalized PageRank uses candidate QueryRelevance as restart seeds and the
// RAG edge ontology weights as transition weights.
//
// Version currentness is deterministic for the MVP: a node with SupersededBy is
// stale and gets 0, a node with a non-empty Version gets 1, and an unsuperseded
// node without a Version gets 0.5. Diversity starts at 1 and subtracts fixed
// source/community penalties for already-selected context, clamped to 0..1.
// DefaultRanker applies access control before scoring, redacts access
// suppressions to score metadata only, and returns stable output ordered by
// total descending then node ID ascending.
package ranker
