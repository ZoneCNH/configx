# 发布

## 版本

v0.1.0

## 必需 Evidence

- `go test ./...`
- `go test -race ./...`
- `make boundary`
- `make security`
- `make contracts`
- `make evidence`
- `release/manifest/latest.json`

## 必需工具

- `golangci-lint`
- `govulncheck`

缺少任一工具时，`make ci` 必须失败。CI workflow 必须在运行 `make ci` 前安装这些工具。

## 发布规则

没有 Evidence 不得发布。
`release/manifest/latest.json` 是生成产物和 CI artifact，不提交到源码历史。
