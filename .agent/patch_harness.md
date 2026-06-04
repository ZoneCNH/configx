# Harness 补丁

## Release Evidence Gate

`make release-final-check` 必须验证以下 artifacts 非空存在：

- `release/manifest/latest.json`
- `release/manifest/latest.json.sha256`
- `release/evidence/gate-report.json`
- `release/evidence/redaction-report.json`
- `release/evidence/contract-hashes.json`

## 一致性校验

- manifest version、commit 和 tree state 必须与当前发布上下文一致。
- checksum sidecar 必须匹配 `release/manifest/latest.json`。
- gate report 中 required commands 必须全部为 `passed`。
- downstream adoption checks 必须全部为 `passed`。
- redaction checks 必须全部为 `passed`，并确认 no raw secret observed。
- contract hashes 必须覆盖所有必需 contract artifacts。
- evidence secret scan 必须覆盖 manifest 和全部 sidecar reports。

## 失败策略

- 缺少必需工具、artifact、hash、gate status 或 downstream evidence 时必须失败。
- 发现 boundary gate 定义的 forbidden discovery literals、自动发现语义或 raw secret-shaped value 时必须失败。
- 工作区不 clean 时必须失败，避免 release evidence 指向未提交源码。
