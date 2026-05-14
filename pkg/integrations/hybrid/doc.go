// Package hybrid accepts hybrid BM25/dense/graph RRF and reranker handoff
// schemas. It preserves upstream relevance details as candidate features;
// CitewiseRAG still applies access gates, ranking policy, diversity, and
// context ordering downstream. RedactForReranker is available for deployments
// that need a coarse ACL pass before sending handoff text to an upstream
// cross-encoder inside a stricter trust boundary.
package hybrid
