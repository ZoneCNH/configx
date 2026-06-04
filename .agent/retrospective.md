# configx v0.1.2 复盘

## 发布信息

- 版本：v0.1.2
- 日期：2026-06-04
- 模式：team execution + local release verification

## 改进项

- 显式 source、deterministic merge、typed decode、validation、redaction 和 sanitized report 已形成可测试公共 API。
- TOML/YAML 文件 source 已纳入明确加载路径，保持无自动发现。
- Release evidence 从单一 manifest 扩展为 manifest、checksum、gate report、redaction report 和 contract hashes 的组合证据。
- `make release-final-check` 已能验证证据 artifacts 存在、状态一致、checksum 匹配、下游采纳通过且 known risks 为空。

## 失败项

- 未发现 P0/P1 功能缺口。
- 收尾审计发现独立复盘 patch files 缺失；本轮补齐为 `.agent/patch_prompt.md`、`.agent/patch_harness.md` 和 `.agent/patch_rule.md`。
- Release branch、tag 和 GitHub Release 未远程发布；该动作需要外部发布授权。

## API 稳定性关注点

- `Decode` tag/default/required/secret 语义需要继续通过 contracts 和 examples 锁定。
- `Merge` strategy 扩展必须保持默认 last-wins，不得破坏现有行为。
- `SecretString` 和 `Redactor` 的字符串、JSON、错误输出安全面必须继续作为回归测试重点。

## 边界风险

- `configx` 不得引入 L2 provider、x.go、真实服务或生产 secret 路径读取。
- YAML/TOML parser 依赖保持为已批准轻量解析库，不扩大为远程配置生态。

## 安全发现

- 发布证据必须只包含 sanitized values、hashes、paths 和 gate facts。
- manifest 和 sidecar reports 均需要 secret scan；不能只扫描源码。

## Secret 处理发现

- DSN、token、password、cookie、private key、access key 和 userinfo URL 都需要按 key/value/type 多层规则脱敏。
- 错误和 release evidence 中只记录 key/source/type/status，不记录 raw value。

## 提示补丁

- 见 `.agent/patch_prompt.md`。

## Harness 补丁

- 见 `.agent/patch_harness.md`。

## 规则补丁

- 见 `.agent/patch_rule.md`。

## CI Gate 建议

- 加入 CodeQL。
- 保留 `govulncheck` 强制模式。
- 加入覆盖率阈值。
- 将 release evidence sidecar validation 作为 CI artifact gate。

## 新 Issue 候选

- ISSUE-CONFIGX-001：补充 public API hash gate。
- ISSUE-CONFIGX-002：补充 config schema hash gate。
- ISSUE-CONFIGX-003：在 CI 发布 job 中上传完整 release evidence artifacts。
