// Package ragnode defines the GraphRAG-aware data model used by CitewiseRAG.
//
// The package is the adapter boundary between legacy pkg/citewise backlog data
// and the richer RAG candidate sets used by the current context-assembly
// pipeline. Old backlog JSON is first parsed by pkg/citewise and then converted
// here with explicit defaults; new RAG JSON is parsed by this package so legacy
// CLI parsing behavior remains unchanged.
package ragnode
