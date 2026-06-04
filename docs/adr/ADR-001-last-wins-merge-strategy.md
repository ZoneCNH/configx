# ADR-001：Last-Wins 合并策略

## 状态

已采纳

## 背景

configx 支持从多个配置源加载配置（环境变量、.env 文件、JSON/YAML/TOML 文件、内存 Map），当多个源定义了同一个 key 时，需要一种明确的冲突解决策略。

核心问题：**当 key 冲突时，哪个源的值生效？**

## 决策

采用 **Last-Wins**（后注册者胜出）作为默认合并策略。

```go
loader := NewLoader().
    AddSource(NewMapSource("defaults", defaults)).   // 先加载：低优先级
    AddSource(NewJSONFileSource("config.json")).      // 后加载：覆盖 defaults
    AddSource(NewEnvSource("APP_", keys))             // 最后加载：最高优先级
```

后添加的源覆盖先前的源，`SourceReport.Overridden` 字段标记被覆盖的值。

### 替代方案

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **First-Wins**（先注册者胜出） | 第一个加载的源生效 | 保护默认值不被意外覆盖 | 需要把"最重要"的源放最前面，违反直觉；无法用环境变量覆盖文件配置 |
| **Deep-Merge**（深度合并） | 对嵌套 Map 递归合并 | 保留部分覆盖 | 值是扁平字符串，没有"嵌套"语义；合并规则复杂、难以预测 |
| **Error-On-Conflict**（冲突报错） | 发现重复 key 立即报错 | 强制显式处理冲突 | 与环境变量覆盖文件配置的常见模式冲突，使用摩擦大 |

### 选择理由

1. **符合直觉**：命令行参数覆盖配置文件、环境变量覆盖默认值——这是 UNIX 哲学和 12-Factor App 的标准模式。
2. **显式排序**：源的优先级由 `AddSource` 调用顺序决定，代码即文档。
3. **实现简单**：遍历源列表依次写入，无需递归合并或复杂冲突检测。
4. **可观测**：`Overridden` 标记和 `SourceReport` 让覆盖关系可审计。

## 影响

### 权衡

- 用户必须理解源的注册顺序即优先级顺序，文档需要明确说明。
- 被覆盖的值默认不保留（仅标记 `Overridden`），需要 `Sanitize()` 或报告才能看到。
- 不支持"部分覆盖嵌套对象"的场景（但 configx 的扁平 Map 模型天然不涉及此问题）。

### 后续工作

- 已提供 `WithMergeStrategy(MergeFirstWins)` 和 `WithMergeStrategy(MergeErrorOnConflict)` 作为备选策略，满足特殊场景需求。
- `SourceReport` 记录每个源的加载结果，便于调试覆盖链。

## 证据

- `pkg/configx/merge.go`：`mergeValue()` 函数实现，`MergeStrategy` 类型定义。
- `pkg/configx/core_test.go`：`TestLoaderMergesSourcesLastWinsAndSanitizesSecrets`、`TestLoaderMergeStrategyFirstWins`、`TestLoaderMergeStrategyErrorOnConflict`。
