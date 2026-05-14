// Package access implements hard access gates for RAG nodes.
//
// Access decisions are binary filters, not ranking inputs. Unauthorized nodes
// must be redacted before they can appear in suppression audit data, and
// controlling Agentic node types require source approval unless an internal
// validation path explicitly sets AttrAllowUnapprovedAgenticNodes=true.
// ApprovedBy is source approval, not a viewer allow-list; the MVP visibility
// model uses clearance and trusted approver checks and intentionally leaves
// broader ABAC dimensions to caller-owned policy layers.
package access
