# configx

`configx` 是面向 ByteChainX services 和 libraries 的显式、dependency-light Go configuration base library。它提供 environment variables、env files、JSON files 和 in-memory maps 的 typed loaders，支持确定性的 source merging、struct decoding、validation hooks，以及面向 logs、health output 和 release evidence 的安全 sanitization。

该 library 有意保持显式：每个 source 和 path 都由调用方选择。它不会自动发现 config files，不会创建 global configuration state，不会注册 singletons，不会 import driver packages，也不会依赖任何 `x.go` module。

## 目标

- 只从调用方提供的 sources 加载 configuration。
- 使用 last-wins 语义以可预测方式 merge sources。
- 使用 `config`、`default` 和 `required` tags decode 到调用方拥有的 structs。
- 在 values 被记录到 log 或 serialized 之前标记并 redact 类似 secret 的 keys。
- 保留 errors、health、metrics、tests、CI 和 release evidence 的稳定 base-library contracts。

## 快速开始

```go
loader := configx.NewLoader().
    AddSource(configx.NewMapSource("defaults", map[string]string{
        "APP_NAME": "service",
        "PORT":     "8080",
    })).
    AddSource(configx.NewEnvSource("APP_", []string{"NAME", "PORT", "API_TOKEN"}))

result, err := loader.Load(context.Background())
if err != nil {
    return err
}

var cfg struct {
    Name  string               `config:"NAME" required:"true"`
    Port  int                  `config:"PORT" default:"8080"`
    Token configx.SecretString `config:"API_TOKEN"`
}
if err := configx.Decode(result, &cfg); err != nil {
    return err
}

safe := result.Sanitize() // secret values are redacted
```

## 公共 API 范围

- `NewLoader`、`Loader.AddSource`、`Loader.Load`：构建并运行显式 source pipelines。
- `NewEnvSource`、`NewAllEnvSource`、`NewEnvFileSource`、`NewJSONFileSource`、`NewMapSource`：具体 source adapters。
- `LoadResult`、`Value`、`SourceReport`、`SanitizedResult`：检查已加载 values 与 source evidence。
- `Decode`：从 `LoadResult` 填充调用方 structs。
- `SecretString`、`NewSecretString`、`IsSecretKey`：由 `foundationx` compatibility 支持的 secret handling helpers。
- `Config`、`New`、`Close`、`HealthCheck`、`Error`、`Metrics`：从 baselib template 保留的 baseline library contracts。

## 非目标

- 不做 implicit config discovery。
- 不提供 process-wide mutable configuration singleton。
- 不引入 hidden driver dependencies。
- 不在 sanitized output 中出现 secret values。
- 不依赖 `github.com/bytechainx/x.go`、`github.com/ZoneCNH/x.go` 或 internal `x.go` packages。

## 命令

如果当前 checkout 位于上层 `go.work` 之下，使用 `GOWORK=off` 运行 validation，以证明 module independence。

```bash
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off ./scripts/check_secrets.sh
```

`make lint` 需要 `golangci-lint`；`make security` 需要 `govulncheck` 和本地 secret scanner。CI 应显式安装这些 tools。

## 文档

- [Goal](docs/goal.md)：权威 product 目标和 acceptance criteria。
- [API](docs/api.md)：公共 configuration API 与 contracts。
- [Config](docs/config.md)：source、merge、decode、validation 和 sanitization 规则。
- [foundationx compatibility](docs/foundationx-compatibility.md)：local compatibility boundary 与 upgrade rule。
- [Testing](docs/testing.md)：unit、contract、race、boundary、security 和 release evidence gates。
- [Release](docs/release.md)：release manifest 与 evidence 要求。
