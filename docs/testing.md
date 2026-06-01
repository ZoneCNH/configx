# 测试模板

## 占位符

- `configx`
- `configx`

## 测试策略

本模板遵循 [测试策略母版](test-strategy.md)。默认强制 SDD、ATDD、TDD、Contract、Boundary、Security、Integration Smoke 和 Evidence；默认增强 Property、Fuzz Smoke、Golden、Compatibility 和 Observability；Chaos、Mutation、Long Soak 和 Full E2E 只由派生库按 profile 启用。

## 测试模式矩阵

| 模式 | 是否默认强制 | Gate | 说明 |
|---|---:|---|---|
| SDD | 是 | `docs/spec.md` | 规格先行 |
| ATDD | 是 | `docs/testing.md` | 验收标准先行 |
| TDD / Unit | 是 | `make test` | 核心逻辑测试 |
| Race | 是 | `make race` | 并发安全 |
| Contract | 是 | `make contracts` | schema、metrics、errors |
| Boundary | 是 | `make boundary` | 模块边界 |
| Security | 是 | `make security` | `govulncheck` 和 secret scan |
| Integration Smoke | 是 | `make integration` | 模板渲染后可运行 |
| Evidence | 是 | `make evidence` / `make release-check` | release manifest 与 gate 结果 |
| Property | 推荐 | `make property` | 不变量测试 |
| Fuzz Smoke | 推荐 | `make fuzz-smoke` | 边界输入测试 |
| Golden | 推荐 | `make golden` | 稳定输出回归 |
| Compatibility | 推荐 | `make contracts` | 公共契约兼容性 |
| Observability | 推荐 | `make contracts` / `make test` | metrics、health、logs |
| Chaos | 按库启用 | profile-specific | 存储和消息库 |
| Mutation | 按库启用 | critical-only | 高风险逻辑 |
| Full BDD | 不默认 | docs only | 基础库不强制 |
| Full DDD | 不作为测试模式 | boundary rule | 只保留边界思想 |

## 必需 Gate

本地执行 gate 前必须可用：

- `golangci-lint`
- `govulncheck`

缺少上述工具时，`make lint` 或 `make security` 必须失败。

- `make fmt`
- `make vet`
- `make lint`
- `make test`
- `make race`
- `make boundary`
- `make security`
- `make contracts`
- `make integration`
- `make evidence`

## 扩展 Gate

扩展 gate 推荐在发布前、公共 API 变更、contract 变更、schema 变更、metrics 变更和安全敏感变更时运行：

- `make property`
- `make fuzz-smoke`
- `make golden`
- `make ci-extended`
- `make release-check-extended`

`make ci` 必须保持轻量，扩展 gate 不进入默认 `make ci`。

## 必需覆盖范围

- `go test ./...` 必须覆盖公共包、`internal/`、`contracts/`、`testkit/` 和 `examples/`。
- 配置校验。
- 配置脱敏。
- 类型化 error kind 和 wrapped cause。
- 客户端创建、取消 context、过期 context。
- 幂等关闭、zero-value client、取消 context。
- 健康与非健康状态检查。
- 健康检查 JSON 字段 contract。
- 生命周期 metrics 和健康 metrics。
- `contracts/` 与公共常量同步。
- `contracts/config.schema.json` 与 `Config` 字段映射同步。
- `scripts/render_template.sh` 生成的临时 ZoneCNH 基础库可以通过 `GOWORK=off go test ./...`。
- `Config.Sanitize` 的 secret 不变量必须由 property test 覆盖。
- `Config` 边界输入必须由 fuzz-smoke 覆盖。
- `HealthStatus` JSON 公共输出必须由 golden test 锁定。

## 示例与 testkit Smoke

- `examples/basic` 必须输出当前 module name。
- `examples/config` 必须输出脱敏后的 secret 值。
- `examples/health` 必须输出 `healthy`。
- `testkit` 必须验证 `Config("fixture")` 生成可通过 `Validate` 的测试配置。
- `testkit.RequireNoError` 必须接受 `nil`，作为生成库测试断言的最小契约。
- `testkit.RequireGolden` 必须比较稳定公共输出，并在 mismatch 时输出 expected 和 actual 上下文。

生成的基础库必须保持测试独立于 `x.go`。
