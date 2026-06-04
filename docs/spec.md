# configx 规格（SPEC v1.0）

## 需求

- 为可复用基础库提供独立 Go module。
- 提供 `Config`、`Validate`、`Sanitize`、`Client`、`New`、`Option`、`HealthCheck`、错误模型、指标钩子和版本元数据。
- `Validate`、`New`、`Close` 和 `HealthCheck` 必须返回或记录可分类的生产语义，包括 typed error、幂等关闭、上下文取消和健康状态。
- 提供 Harness Gate 脚本、生成脚本、CI 工作流、contracts、examples、发布证据 artifact、release 和复盘模板。

## 验收标准

- `GOWORK=off go test ./...` 和 `GOWORK=off go test -race ./...` 通过。
- `GOWORK=off make release-check` 通过，并以 `CHECK_STATUS=passed` 生成未提交的 `release/manifest/latest.json` 发布证据 artifact。
- `contracts/config.schema.json` 与 `Config` 字段映射保持一致，`timeout_ms` 映射到 `Config.Timeout`。
- `contracts/error.schema.json`、`contracts/health.schema.json` 和 `contracts/metrics.md` 与公共常量保持一致。
- `scripts/render_template.sh` 可以生成非自引用的 ZoneCNH 基础库形态并通过 `GOWORK=off go test ./...`。
- 模块不得依赖 `github.com/bytechainx/x.go` 或 `github.com/ZoneCNH/x.go`。
- 模块不得隐式读取生产密钥或自动发现 `.env`、`production.yaml`、`config.local.yaml`、`/home/k8s/secrets/env/*`。
- `docs/current-state.md` 必须记录当前 MVA surface、边界约束和最新可复现证据命令。

## 非目标

- 不包含业务模型、生产连接默认值和隐藏全局客户端。

## 可追踪性

- 目标：`GOAL-20260601-001`
- 模板占位符：`configx`、`github.com/ZoneCNH/configx`、`configx`
