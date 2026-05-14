//go:build graphrag_parquet

package graphrag

import (
	"io"
	"os"

	parquet "github.com/parquet-go/parquet-go"
)

type ParquetFiles struct {
	Documents        string
	TextUnits        string
	Entities         string
	Relationships    string
	CommunityReports string
	Communities      string
	Covariates       string
}

func ParseParquetFiles(files ParquetFiles) (Export, error) {
	var export Export
	if files.Documents != "" {
		rows, err := readParquetRows(files.Documents)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.Documents = append(export.Documents, Document{ID: getString(row, "id"), HumanReadableID: getString(row, "human_readable_id"), Title: getString(row, "title"), Text: getString(row, "text"), Period: getString(row, "period")})
		}
	}
	if files.TextUnits != "" {
		rows, err := readParquetRows(files.TextUnits)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.TextUnits = append(export.TextUnits, TextUnit{ID: getString(row, "id"), HumanReadableID: getString(row, "human_readable_id"), DocumentID: getString(row, "document_id"), Text: getString(row, "text"), NTextTokens: int(getFloat(row, "n_tokens")), CommunityID: getString(row, "community_id"), Period: getString(row, "period")})
		}
	}
	if files.Entities != "" {
		rows, err := readParquetRows(files.Entities)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.Entities = append(export.Entities, Entity{ID: getString(row, "id"), HumanReadableID: getString(row, "human_readable_id"), Title: getString(row, "title"), Description: getString(row, "description"), CommunityID: getString(row, "community_id"), Period: getString(row, "period")})
		}
	}
	if files.Relationships != "" {
		rows, err := readParquetRows(files.Relationships)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.Relationships = append(export.Relationships, Relationship{ID: getString(row, "id"), Source: getString(row, "source"), Target: getString(row, "target"), SourceID: getString(row, "source_id"), TargetID: getString(row, "target_id"), Description: getString(row, "description"), Weight: getFloat(row, "weight")})
		}
	}
	if files.CommunityReports != "" {
		rows, err := readParquetRows(files.CommunityReports)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.CommunityReports = append(export.CommunityReports, CommunityReport{ID: getString(row, "id"), HumanReadableID: getString(row, "human_readable_id"), Community: getString(row, "community"), Title: getString(row, "title"), Summary: getString(row, "summary"), FullContent: getString(row, "full_content"), Rank: getFloat(row, "rank"), Period: getString(row, "period")})
		}
	}
	if files.Communities != "" {
		rows, err := readParquetRows(files.Communities)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.Communities = append(export.Communities, Community{ID: getString(row, "id"), HumanReadableID: getString(row, "human_readable_id"), Community: getString(row, "community"), Title: getString(row, "title"), EntityIDs: getStringSlice(row, "entity_ids"), TextUnitIDs: getStringSlice(row, "text_unit_ids"), ReportID: getString(row, "report_id")})
		}
	}
	if files.Covariates != "" {
		rows, err := readParquetRows(files.Covariates)
		if err != nil {
			return Export{}, err
		}
		for _, row := range rows {
			export.Covariates = append(export.Covariates, Covariate{ID: getString(row, "id"), HumanReadableID: getString(row, "human_readable_id"), SubjectID: getString(row, "subject_id"), TextUnitID: getString(row, "text_unit_id"), Claim: getString(row, "claim"), Status: getString(row, "status"), Period: getString(row, "period"), EvidenceIDs: getStringSlice(row, "evidence_ids")})
		}
	}
	return export, nil
}

func readParquetRows(path string) ([]map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := parquet.NewGenericReader[map[string]any](f)
	defer r.Close()
	var out []map[string]any
	buf := make([]map[string]any, 64)
	for {
		n, err := r.Read(buf)
		out = append(out, buf[:n]...)
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func getString(row map[string]any, key string) string {
	if v, ok := row[key].(string); ok {
		return v
	}
	if b, ok := row[key].([]byte); ok {
		return string(b)
	}
	return ""
}

func getFloat(row map[string]any, key string) float64 {
	switch v := row[key].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

func getStringSlice(row map[string]any, key string) []string {
	switch v := row[key].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
