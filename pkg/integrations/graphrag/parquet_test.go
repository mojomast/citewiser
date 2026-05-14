//go:build graphrag_parquet

package graphrag

import "testing"

func TestParquetBuildTagCompiles(t *testing.T) {
	if _, err := ParseParquetFiles(ParquetFiles{}); err != nil {
		t.Fatal(err)
	}
}
