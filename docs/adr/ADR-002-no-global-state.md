# ADR-002：无全局可变状态

## 状态

已采纳

## 背景

许多配置库使用全局单例或 `init()` 函数来管理配置状态（如 Viper 的 `viper.Set()`/`viper.Get()`，envconfig 的 `init()` 自动绑定）。

configx 需要决定是否采用类似模式。

## 决策

configx **不使用任何全局可变状态**。所有配置操作通过显式创建的 `Loader` 和 `Client` 实例完成。

```go
// 显式创建，无全局副作用
loader := NewLoader().
    AddSource(NewEnvSource("APP_", keys))
result, err := loader.Load(ctx)

// 通过 Client 管理生命周期
client, err := NewClient(ctx, result, WithMetrics(m))
defer client.Close()
```

### 替代方案

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **全局单例** | `var defaultLoader = NewLoader()` + 包级函数 | 调用方便，一行代码 | 测试污染、并发竞争、无法隔离多个配置实例 |
| **init() 自动加载** | 包导入时自动读取环境变量 | 零配置启动 | 加载时机不可控、无法注入 context、测试困难 |
| **sync.Once 懒加载** | 首次访问时初始化全局实例 | 延迟初始化 | 仍然存在全局状态、测试隔离困难、无法动态重载 |

### 选择理由

1. **测试隔离**：每个测试创建独立的 Loader/Client，`t.Setenv()` 互不干扰，无需 `TestMain` 清理全局状态。
2. **并发安全**：无共享可变状态，Loader 的 `Load()` 是纯函数（输入源列表 → 输出结果），天然并发安全。
3. **显式依赖**：函数签名 `func NewClient(ctx, result, opts)` 清楚表达了依赖关系，不依赖隐式全局初始化。
4. **多实例支持**：同一进程中可以有多个独立的配置客户端（如微服务网关同时管理多个上游配置）。

## 影响

### 权衡

- 调用方需要多写几行代码来创建和传递 Loader/Client 实例。
- 没有"全局便捷函数"，简单脚本场景略有冗余。
- 提供了 `LoadEnv()`、`LoadJSONFile()` 等包级便捷函数，内部创建临时 Loader，不保留全局状态。

### 后续工作

- `Client` 使用 `sync.RWMutex` 保护内部状态（`client.go`），确保多 goroutine 安全访问。
- 便捷函数（`LoadEnv`、`LoadJSONFile` 等）为简单场景提供一行式调用，平衡便利性与无全局状态原则。

## 证据

- `pkg/configx/loader.go`：`NewLoader()` 返回新实例，无包级变量。
- `pkg/configx/client.go`：`Client` 通过 `NewClient()` 显式创建，`mu sync.RWMutex` 保护状态。
- `pkg/configx/options.go`：函数式选项模式，无全局默认值。
- `pkg/configx/core_test.go`：每个测试独立创建 Loader，无共享状态。
