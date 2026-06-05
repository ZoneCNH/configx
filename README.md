# configx

`configx` 是面向 ZoneCNH 服务与库的显式、轻依赖 Go 配置基础库。它提供环境变量、env 文件、JSON、TOML、YAML 文件与内存映射的类型化加载器，支持可预测的 source 合并、结构体解码、校验 hooks，以及面向日志、health 输出和发布证据的安全脱敏。

该库有意保持显式：每个 source 和路径都由调用方选择。它不会自动发现配置文件，不会创建全局配置状态，不会注册单例，不会导入 driver packages，也不会依赖任何 `x.go` module。

## 目标

- 只从调用方提供的 sources 加载配置。
- 使用 last-wins 语义以可预测方式合并 sources。
- 使用 `config`、`default` 和 `required` tags 解码到调用方拥有的 structs。
- 在 values 被记录到日志或序列化之前，标记并脱敏类似 secret 的 keys。
- 保留 errors、health、metrics、tests、CI 和发布证据所需的稳定基础库契约。

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

safe := result.Sanitize() // secret 值已脱敏
```

## 公共 API 范围

- `NewLoader`、`Loader.AddSource`、`Loader.Load`：构建并运行显式 source 管线。
- `NewEnvSource`、`NewAllEnvSource`、`NewEnvFileSource`、`NewJSONFileSource`、`NewTOMLFileSource`、`NewYAMLFileSource`、`NewMapSource`：具体 source 适配器。
- `LoadEnvFile`、`LoadJSONFile`、`LoadTOMLFile`、`LoadYAMLFile`、`LoadMap`：单 source convenience loaders。
- `LoadResult`、`Value`、`SourceReport`、`SanitizedResult`：检查已加载 values 与 source 证据。
- `Decode`：从 `LoadResult` 填充调用方 structs。
- `SecretString`、`NewSecretString`、`IsSecretKey`：由 `foundationx` compatibility 支持的 secret 处理 helpers。
- `Config`、`New`、`Close`、`HealthCheck`、`Error`、`Metrics`：从 xlib-standard 继承的基础库契约。

## 非目标

- 不做隐式配置发现。
- 不提供进程级可变配置单例。
- 不引入隐藏驱动依赖。
- 不在脱敏输出中出现原始 secret 值。
- 不依赖 `github.com/bytechainx/x.go`、`github.com/ZoneCNH/x.go` 或内部 `x.go` packages。

## 命令

如果当前 checkout 位于上层 `go.work` 之下，使用 `GOWORK=off` 运行 validation，以证明 module independence。

```bash
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off ./scripts/check_secrets.sh
```

`make lint` 需要 `golangci-lint`；`make security` 需要 `govulncheck` 和本地密钥扫描工具。CI 应显式安装这些工具。

## 文档

- [目标](docs/goal.md)：权威产品目标和验收标准。
- [API](docs/api.md)：公共配置 API 与契约。
- [配置](docs/config.md)：source、merge、decode、validation 和 sanitization 规则。
- [foundationx 兼容性](docs/foundationx-compatibility.md)：本地兼容边界与升级规则。
- [测试](docs/testing.md)：unit、contract、race、boundary、security 和发布证据 Gate。
- [发布](docs/release.md)：release manifest 与证据要求。
- [ADR](docs/adr/)：架构决策记录（merge 策略、无全局状态、显式加载）。
- [项目分析](docs/project-analysis-20260604.md)：结构分析与评分报告（v0.1.3 满分）。
