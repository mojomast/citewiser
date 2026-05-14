// Package access implements hard access gates for RAG nodes.
//
// Access decisions are binary filters, not ranking inputs. Unauthorized nodes
// must be redacted before they can appear in suppression audit data, and
// controlling Agentic node types require source approval unless the caller
// explicitly sets allow_unapproved_agentic_nodes=true.
package access
