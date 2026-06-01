# 复盘

## 改进项

- 基础库创建从手工变成模板化。
- 后续 `foundationx`、`postgresx`、`kafkax`、`redisx` 可复用目录、脚本、CI、文档和 Evidence。

## 失败项

- 最终验证运行中没有必需 Gate 失败。
- `govulncheck` 已从可选检查收紧为强制 Security Gate；缺失时必须失败，不能跳过。

## 提示补丁

- 后续创建基础库时必须从 `configx` 复制。
- 所有基础库必须保留 Boundary Gate 和 Secret Gate。

## Harness 补丁

- 后续加入 public API hash gate。
- 后续加入 config schema hash gate。

## 规则补丁

- 禁止基础库依赖 `x.go`。
- 禁止基础库承载业务语义。
- 禁止无 Evidence 声称 `DONE`。

## CI Gate 建议

- 加入 CodeQL。
- 保留 govulncheck 强制模式。
- 加入覆盖率阈值。

## 新 Issue 候选

- ISSUE-FOUNDATIONX-001 从 `configx` 生成 `foundationx`。
- ISSUE-POSTGRESX-001 从 `configx` 生成 `postgresx`。
