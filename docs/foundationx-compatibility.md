# foundationx 兼容边界（Compatibility boundary）

`configx` 通过 `internal/foundationx` 中的 local replacement module 依赖 `github.com/bytechainx/foundationx`。这让 config library 保持可构建，同时保留调用方使用的 tested foundation compatibility surface。

## 兼容范围（Compatibility scope）

local module 只刻意镜像 `configx` 当前测试锁定的 foundation API：

- `SecretString`、`NewSecretString`、`Reveal`、redacted string/text/JSON formatting、`Sanitize` 和 `IsZero`
- `Error`、`ErrorKind`、`NewError`、`WrapError`、`WithRetryable`、`IsKind` 和 `AsFoundationError`
- `Clock`、`RealClock` 和 `FixedClock`
- `RetryPolicy` 和 `DefaultRetryPolicy`
- `HealthStatus`、`HealthChecker` 和 health status constants
- `Starter`、`Closer` 和 `Lifecycle`
- `VersionInfo`、`NewVersionInfo` 和 `String`

`configx` public API 主要 re-export `SecretString` 并把 typed errors 映射到 foundationx。由于 `go.mod` 使用 local replace，contract tests 也会锁定上面的 supporting helpers，避免调用方在当前 module 边界内遇到 foundationx drift。未列出的 foundationx API 不属于兼容范围，除非先补充 contract tests。

## 不可变边界（Non-negotiable boundaries）

- `docs/goal.md` 保持 authoritative contract，template application work 不得重写它。
- 配置加载保持显式：调用方创建 loaders，并传入每个 source 或 path。library 不得 auto-discover `.env`、`production.yaml`、`config.local.yaml` 或 `/home/k8s/secrets/env/*`。
- 模块（module）不得包含生成的 `x.go` files，也不得 import `x.go` 或 Redis、Kafka、PostgreSQL、TDengine、object-storage SDKs 等 infrastructure driver packages。
- 验证 evidence、examples、release manifests 与 documentation 只能使用 sanitized secret output。

## 升级规则（Upgrade rule）

将 local module 替换为 upstream foundationx release 前，必须先用 `GOWORK=off go test ./...` 以及 boundary、contract、secret scanners 证明 `SecretString` redaction、JSON/text marshaling、typed errors、clock、retry、health、lifecycle 和 version helpers 保持兼容。
