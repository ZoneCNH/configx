# Evidence

2026-06-01 采集的完成 Evidence 需要区分 required gate 和 extended gate。

## Required Evidence

- `go test ./...`：通过。
- `go test -race ./...`：通过。
- `make boundary`：通过。
- `make security`：通过；`govulncheck` 和密钥扫描均已通过。
- `make contracts`：通过。
- `make evidence`：通过，并生成 `release/manifest/latest.json`。
- `make release-check`：通过。

## Extended Evidence

推荐在发布前强验证中记录：

- `make property`。
- `make golden`。
- `make fuzz-smoke`。
- `make ci-extended`。
- `make release-check-extended`。

最终声明必须使用：

DONE with evidence:
