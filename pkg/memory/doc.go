// Package memory persists accepted context plans as JSONL and re-gates them
// before reuse. The file-backed store writes only non-red plans, stores evidence
// paths and node versions, and uses deterministic topic Jaccard for similarity.
// External consumers can implement Store to replace FileStore with their own
// persistence layer.
package memory
