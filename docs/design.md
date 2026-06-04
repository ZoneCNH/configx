# configx 设计 v1.0

## 架构

生成的库是独立 Go module。公共 API 位于 `pkg/configx`，内部辅助代码位于 `internal/`，契约位于 `contracts/`，运行证据位于 `release/manifest/`。`scripts/render_template.sh` 是模板到具体基础库的唯一内置渲染入口。

## 公共 API

模板暴露 `Config`、`SanitizedConfig`、`Client`、`New`、`Close`、`Option`、`HealthCheck`、`Error`、`NewError`、`WrapError`、`IsKind`、`Metrics`、`NoopMetrics`、指标常量、`ModuleName` 和 `Version`。

## 配置

调用方必须显式传入配置。生成的库不得隐式读取 `x.go` 生产密钥路径、`.env`、`production.yaml`、`config.local.yaml` 或 `/home/k8s/secrets/env/*`。`Validate` 使用稳定校验错误表达缺失字段和负数 timeout，`Sanitize`、`LoadResult.Sanitize()` 与 `SecretString` 只返回可安全记录的脱敏视图。`contracts/config.schema.json` 使用外部字段 `timeout_ms`，并通过契约回归测试锁定到 `Config.Timeout`。

## 错误模型

错误使用稳定的 `ErrorKind` 枚举，并通过 `Unwrap` 支持错误包装。上下文超时归类为 `timeout` 且可重试；上下文取消归类为 `unavailable` 且不可重试。

## 健康检查

持有资源的客户端暴露 `HealthCheck(context.Context)`，并返回 `healthy`、`degraded` 或 `unhealthy`。返回结构使用 `name`、`status`、`message`、`checked_at`、`latency_ms` 和 `metadata` JSON 字段；nil client、零值 client、已关闭 client、nil context 和已取消 context 都必须返回 `unhealthy`。

## 指标

指标通过钩子注入，默认使用无操作实现。模板锁定 client 生命周期、错误、健康检查、请求、重试和 inflight 指标名称，具体列表以 `contracts/metrics.md` 和 `pkg/configx` 指标常量为准。

## 测试

模板要求为配置校验、脱敏、客户端生命周期、健康检查和内部辅助代码提供单元测试与竞态测试。

## 安全与边界

`docs/security.md` 和 `docs/redaction.md` 是安全边界的读者入口。边界脚本拒绝 `x.go`、隐式配置发现、L2 provider/cloud/driver dependencies 和业务术语漂移；发布证据只能包含脱敏 manifest 与 gate 状态。

## 发布

发布前必须通过 Harness Gate，并生成 `release/manifest/latest.json`。`latest.json` 是发布证据 artifact，不提交到源码历史；仓库只提交 `release/manifest/template.json`。`make release-check` 会先运行 CI 和 integration gate，再以 `CHECK_STATUS=passed` 生成 manifest；manifest 记录实际执行 gate 的 `commit`、`generated_by`、`go_version` 和 `tree_state`。integration gate 会渲染临时下游基础库并运行测试，防止模板替换链路回归。
