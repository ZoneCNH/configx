# foundationx compatibility boundary（兼容边界）

`configx` 通过 `internal/foundationx` 中的 local replacement module 依赖 `github.com/bytechainx/foundationx`。这让 config library 保持可构建，同时保留调用方使用的 public `foundationx.SecretString` contract。

## Compatibility scope

local module 只刻意镜像 `configx` 所需的 foundation API：

- `SecretString`、`NewSecretString`、`Reveal`、redacted string/text/JSON formatting、`Sanitize` 和 `IsZero`
- configx error wrapping 使用的最小 typed errors

`/tmp/configx-foundationx` 等 reference modules 还包含 health、lifecycle、retry、clock 和 version contracts。除非某个 configx feature import 这些 API 并添加匹配 tests，否则这些 API 不会由 `configx` re-export 或使用。这样可防止 template drift 把 infrastructure behavior 拉入基础配置库。

## Non-negotiable boundaries

- `docs/goal.md` 保持 authoritative contract，template application work 不得重写它。
- Config loading 保持显式：调用方创建 loaders，并传入每个 source 或 path。library 不得 auto-discover `.env`、`production.yaml`、`config.local.yaml` 或 `/home/k8s/secrets/env/*`。
- module 不得包含生成的 `x.go` files，也不得 import `x.go` 或 Redis、Kafka、PostgreSQL、TDengine、object-storage SDKs 等 infrastructure driver packages。
- Validation evidence、examples、release manifests 与 documentation 只能使用 sanitized secret output。

## Upgrade rule

将 local module 替换为 upstream foundationx release 前，必须先用 `GOWORK=off go test ./...` 以及 boundary、contract、secret scanners 证明 `SecretString` redaction、`Sanitize`、`IsZero`、JSON/text marshaling 与 configx error behavior 保持不变。
