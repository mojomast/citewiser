// Package hygiene analyzes graph readiness, missing-edge suggestions, and
// corrective signals for RAG context assembly. It wraps existing Citewise
// duplicate/orphan/stale/missing-bridge checks where available and adds
// deterministic RAG-specific scoring and edge suggestions. Analyzer is the
// named interface for downstream consumers that need to mock or replace hygiene.
package hygiene
