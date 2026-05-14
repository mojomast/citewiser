//go:build !graphrag_parquet

package graphrag

import "errors"

var ErrParquetSupportDisabled = errors.New("graphrag parquet support requires the graphrag_parquet build tag")

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
	return Export{}, ErrParquetSupportDisabled
}
