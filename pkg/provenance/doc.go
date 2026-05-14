// Package provenance builds source references and source trails for selected
// context nodes.
//
// Builders are deterministic and access-agnostic. Callers must redact trails
// with the access controller before data crosses package or response
// boundaries.
package provenance
