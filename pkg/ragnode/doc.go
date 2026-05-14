// Package ragnode defines the GraphRAG-aware data model used by CitewiseRAG.
//
// The package is the adapter boundary between the existing pkg/citewise backlog
// model and richer RAG candidate sets. Old backlog JSON is first parsed by
// pkg/citewise and then converted here with explicit defaults; new RAG JSON is
// parsed by this package so existing CLI parsing behavior remains unchanged.
package ragnode
