# 发布

## 版本

v0.1.2

## 必需 Evidence

- `GOWORK=off go test ./...`
- `GOWORK=off make ci`
- `GOWORK=off make ci-extended`
- `XLIB_CONTEXT=release_verify GOWORK=off make release-check`
- `XLIB_CONTEXT=release_verify GOWORK=off make release-final-check`
- `release/manifest/latest.json`
- `release/manifest/latest.json.sha256`
- `release/evidence/gate-report.json`
- `release/evidence/redaction-report.json`
- `release/evidence/contract-hashes.json`

## 必需工具

- `golangci-lint`
- `govulncheck`

缺少任一工具时，`make ci` 必须失败。CI workflow 必须在运行 `make ci` 前安装这些工具。

## 发布规则

没有 Evidence 不得发布。
发布证据 artifacts 是生成产物和 CI artifact，不提交到源码历史。
发布 manifest 必须记录最终提交、clean tree、全部 gate 通过、下游采纳证据和空 known risks。
远程 branch、tag 和 GitHub Release 属于外部发布动作，执行前必须获得明确授权。
