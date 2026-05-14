package graphrag

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/mojomast/citewiser/pkg/citewise"
	"github.com/mojomast/citewiser/pkg/ragnode"
)

type Export struct {
	QueryID          string            `json:"query_id,omitempty"`
	Query            string            `json:"query,omitempty"`
	Documents        []Document        `json:"documents,omitempty"`
	TextUnits        []TextUnit        `json:"text_units,omitempty"`
	Entities         []Entity          `json:"entities,omitempty"`
	Relationships    []Relationship    `json:"relationships,omitempty"`
	CommunityReports []CommunityReport `json:"community_reports,omitempty"`
	Communities      []Community       `json:"communities,omitempty"`
	Covariates       []Covariate       `json:"covariates,omitempty"`
}

type Document struct {
	ID              string `json:"id"`
	HumanReadableID string `json:"human_readable_id,omitempty"`
	Title           string `json:"title,omitempty"`
	Text            string `json:"text,omitempty"`
	Period          string `json:"period,omitempty"`
}

type TextUnit struct {
	ID              string `json:"id"`
	HumanReadableID string `json:"human_readable_id,omitempty"`
	DocumentID      string `json:"document_id,omitempty"`
	Text            string `json:"text,omitempty"`
	NTextTokens     int    `json:"n_tokens,omitempty"`
	CommunityID     string `json:"community_id,omitempty"`
	Period          string `json:"period,omitempty"`
}

type Entity struct {
	ID              string   `json:"id"`
	HumanReadableID string   `json:"human_readable_id,omitempty"`
	Title           string   `json:"title,omitempty"`
	Description     string   `json:"description,omitempty"`
	CommunityID     string   `json:"community_id,omitempty"`
	TextUnitIDs     []string `json:"text_unit_ids,omitempty"`
	Period          string   `json:"period,omitempty"`
}

type Relationship struct {
	ID          string  `json:"id,omitempty"`
	Source      string  `json:"source,omitempty"`
	Target      string  `json:"target,omitempty"`
	SourceID    string  `json:"source_id,omitempty"`
	TargetID    string  `json:"target_id,omitempty"`
	Description string  `json:"description,omitempty"`
	Weight      float64 `json:"weight,omitempty"`
}

type CommunityReport struct {
	ID              string  `json:"id"`
	HumanReadableID string  `json:"human_readable_id,omitempty"`
	Community       string  `json:"community,omitempty"`
	Title           string  `json:"title,omitempty"`
	Summary         string  `json:"summary,omitempty"`
	FullContent     string  `json:"full_content,omitempty"`
	Rank            float64 `json:"rank,omitempty"`
	Period          string  `json:"period,omitempty"`
}

type Community struct {
	ID              string   `json:"id"`
	HumanReadableID string   `json:"human_readable_id,omitempty"`
	Community       string   `json:"community,omitempty"`
	Title           string   `json:"title,omitempty"`
	EntityIDs       []string `json:"entity_ids,omitempty"`
	TextUnitIDs     []string `json:"text_unit_ids,omitempty"`
	ReportID        string   `json:"report_id,omitempty"`
}

type Covariate struct {
	ID              string   `json:"id"`
	HumanReadableID string   `json:"human_readable_id,omitempty"`
	SubjectID       string   `json:"subject_id,omitempty"`
	TextUnitID      string   `json:"text_unit_id,omitempty"`
	Claim           string   `json:"claim,omitempty"`
	Status          string   `json:"status,omitempty"`
	Period          string   `json:"period,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
}

func ParseJSON(r io.Reader) (ragnode.CandidateSet, error) {
	var export Export
	if err := json.NewDecoder(r).Decode(&export); err != nil {
		return ragnode.CandidateSet{}, err
	}
	return MapExport(export), nil
}

func MapExport(export Export) ragnode.CandidateSet {
	set := ragnode.CandidateSet{QueryID: export.QueryID, Query: export.Query}
	for _, doc := range export.Documents {
		set.Nodes = append(set.Nodes, node(doc.ID, doc.Title, doc.Text, ragnode.ChunkDocument, "", doc.HumanReadableID, doc.Period))
	}
	for _, unit := range export.TextUnits {
		n := node(unit.ID, unit.ID, unit.Text, ragnode.ChunkChunk, unit.CommunityID, unit.HumanReadableID, unit.Period)
		n.TokenCount = unit.NTextTokens
		n.Locator.DocumentID = unit.DocumentID
		set.Nodes = append(set.Nodes, n)
	}
	for _, entity := range export.Entities {
		set.Nodes = append(set.Nodes, node(entity.ID, entity.Title, entity.Description, ragnode.ChunkEntity, entity.CommunityID, entity.HumanReadableID, entity.Period))
	}
	for _, report := range export.CommunityReports {
		text := report.FullContent
		if text == "" {
			text = report.Summary
		}
		n := node(report.ID, report.Title, text, ragnode.ChunkCommunitySummary, report.Community, report.HumanReadableID, report.Period)
		n.Type = citewise.RoleOverview
		if n.Attributes == nil {
			n.Attributes = map[string]string{}
		}
		n.Attributes["community_report_rank"] = strconv.FormatFloat(report.Rank, 'f', -1, 64)
		set.Nodes = append(set.Nodes, n)
	}
	for _, covariate := range export.Covariates {
		n := node(covariate.ID, covariate.ID, covariate.Claim, ragnode.ChunkClaim, "", covariate.HumanReadableID, covariate.Period)
		set.Nodes = append(set.Nodes, n)
		for _, evidenceID := range covariate.EvidenceIDs {
			edgeType := ragnode.EdgeEvidenceFor
			if covariate.Status == "negative" || covariate.Status == "false" || covariate.Status == "disputed" {
				edgeType = ragnode.EdgeEvidenceAgainst
			}
			set.Edges = append(set.Edges, ragnode.Edge{SourceID: evidenceID, TargetID: covariate.ID, Type: edgeType, Confidence: 1})
		}
	}
	for _, rel := range export.Relationships {
		source := firstNonEmpty(rel.SourceID, rel.Source)
		target := firstNonEmpty(rel.TargetID, rel.Target)
		set.Edges = append(set.Edges, ragnode.Edge{SourceID: source, TargetID: target, Type: ragnode.EdgeRelatedTo, Confidence: normalizeWeight(rel.Weight), Note: rel.Description})
	}
	for _, community := range export.Communities {
		reportID := community.ReportID
		if reportID == "" {
			reportID = community.ID
		}
		for _, id := range community.EntityIDs {
			set.Edges = append(set.Edges, ragnode.Edge{SourceID: id, TargetID: reportID, Type: ragnode.EdgeCommunityMemberOf, Confidence: 0.8})
		}
		for _, id := range community.TextUnitIDs {
			set.Edges = append(set.Edges, ragnode.Edge{SourceID: id, TargetID: reportID, Type: ragnode.EdgeCommunityMemberOf, Confidence: 0.8})
		}
	}
	return set
}

func node(id, title, text string, chunkType ragnode.ChunkType, communityID, humanReadableID, period string) ragnode.RAGNode {
	return ragnode.RAGNode{Item: citewise.Item{ID: id, Title: title, Notes: text, Source: "graphrag", Trust: 0.75}, Text: text, ChunkType: chunkType, Sensitivity: ragnode.SensitivityInternal, Origin: "graphrag", CommunityID: communityID, ObservedAt: parsePeriod(period), Attributes: map[string]string{"human_readable_id": humanReadableID}}
}

func normalizeWeight(weight float64) float64 {
	if weight <= 0 {
		return 1
	}
	if weight > 1 {
		return 1
	}
	return weight
}

func parsePeriod(period string) time.Time {
	if period == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(layout, period); err == nil {
			return t
		}
	}
	return time.Time{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
