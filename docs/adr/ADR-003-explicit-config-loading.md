# ADR-003：显式配置加载（无自动发现）

## 状态

已采纳

## 背景

部分配置库支持自动发现配置文件（如 Viper 的 `viper.AddConfigPath()` + `viper.ReadInConfig()` 自动搜索目录），或自动扫描带特定前缀的环境变量（如 `envconfig.Process("APP")` 自动绑定 `APP_*`）。

configx 需要决定是否支持这类"自动发现"机制。

## 决策

configx **要求所有配置源显式注册**，不支持自动发现。

```go
// 显式声明每个源——代码即文档
loader := NewLoader().
    AddSource(NewEnvFileSource("/etc/app/config.env")).
    AddSource(NewJSONFileSource("/etc/app/config.json")).
    AddSource(NewEnvSource("APP_", []string{"PORT", "HOST", "TOKEN"}))

// 不会自动搜索 /etc/app/ 下的其他文件
// 不会自动扫描所有 APP_* 环境变量（除非使用 NewAllEnvSource）
```

### 替代方案

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **自动搜索配置文件** | 在指定目录中自动查找 `.env`、`config.json` 等 | 启动简单，零配置 | 文件存在性不可预测、不同环境行为不一致、调试困难 |
| **环境变量前缀扫描** | 自动绑定 `APP_*` 前缀的所有环境变量 | 减少样板代码 | 意外绑定无关变量、安全风险（泄露敏感 env）、行为不可预测 |
| **约定优于配置** | 按约定路径（如 `./config.yaml`）自动加载 | 减少决策 | 路径硬编码、不同环境需要不同约定、容器化场景路径不固定 |

### 选择理由

1. **可预测性**：代码中显式列出所有源，运行时行为与代码完全一致，无"魔法"。
2. **安全性**：不会意外读取文件系统中的无关文件或绑定非预期的环境变量。在安全敏感环境中，自动扫描可能暴露不应读取的配置。
3. **可调试性**：配置来源在代码中一目了然。`SourceReport` 精确记录每个源的加载结果，不存在"它从哪来的？"问题。
4. **容器友好**：容器化部署中配置路径、环境变量前缀由部署脚本显式控制，自动发现反而增加不确定性。

## 影响

### 权衡

- 配置源较多时，`AddSource` 调用链较长（可通过辅助函数封装）。
- 新增配置文件需要修改代码并重新部署，不能"放个文件就生效"。
- 提供了 `NewAllEnvSource(prefix)` 作为"扫描指定前缀的所有环境变量"的显式选择，用户清楚知道自己在做什么。

### 后续工作

- `NewEnvSource` 默认只读取显式列出的 key 列表，`NewAllEnvSource` 显式选择扫描模式。
- 文件源（JSON、YAML、TOML、envfile）都需要明确的文件路径，不搜索目录。
- `SourceReport` 提供完整的源审计信息，包括名称、类型、路径、加载状态和值列表。

## 证据

- `pkg/configx/source_env.go`：`EnvSource` 的 `keys` 字段要求显式传入 key 列表；`NewAllEnvSource` 需要显式选择。
- `pkg/configx/source_file.go`：`EnvFileSource`、`JSONFileSource` 要求显式路径参数。
- `pkg/configx/structured_sources.go`：`TOMLFileSource`、`YAMLFileSource` 同样要求显式路径。
- `pkg/configx/core_test.go`：`TestNoImplicitConfigDiscovery` 验证空 Loader 不产生任何值。
- `pkg/configx/core_test.go`：`TestEnvSourceReadsOnlyExplicitKeys` 验证只读取声明的 key。
