# Release Audit

Audit date: 2026-05-14

Verification commands run:

- `go test ./...`
- `go build ./...`
- `go test -tags graphrag_parquet ./...`
- `go test ./... && go test ./... && go test ./pkg/citewise ./pkg/ranker ./pkg/packer`
- `go run . roles --file testdata/citewise_backlog.json`
- `go run . score --file testdata/citewise_backlog.json`
- `go run . queue --file testdata/citewise_backlog.json`
- `go run . explain --file testdata/citewise_backlog.json --id foundation`
- `go run . hygiene --file testdata/citewise_backlog.json`
- `go run . export --file testdata/citewise_backlog.json`
- `go run . rag route --file testdata/api_examples/pack_request.json`
- `go run . rag rank --file testdata/api_examples/pack_request.json`
- `go run . rag pack --file testdata/api_examples/pack_request.json --token-budget 1200`
- `go run . rag hygiene --file testdata/api_examples/pack_request.json`
- `go run . rag route --file - < testdata/api_examples/pack_request.json`
- `go list -deps ./...`
- Optional parquet source/build-tag guardrail checked by `release_guardrail_test.go`

Dependency audit notes:

- `go list -deps ./...` shows required non-tagged builds use project packages plus Go standard library packages only.
- `net/http`, `context`, and `log/slog` appear through optional `cmd/serve`, not core RAG packages.
- No non-tagged dependency on LLM SDKs, vector databases, Redis, Neo4j, Bleve, web-router frameworks, Gonum, or Arrow was observed.
- Parquet-related external packages appear only when listing dependencies with `-tags graphrag_parquet` for `pkg/integrations/graphrag`.
- `go.mod` records parquet transitive modules because optional parquet support is implemented; the imports are isolated behind the `graphrag_parquet` build tag.
- `release_guardrail_test.go` now enforces the non-tagged dependency boundary and verifies parquet imports remain isolated in files guarded by `graphrag_parquet`; this keeps default tests runnable with the module's `go 1.24` directive even when the optional parquet module requires a newer patch toolchain.
- RAG CLI golden stdout fixtures under `testdata/rag_cli_golden/` lock the JSON output for `route`, `rank`, `pack`, and `hygiene`.
