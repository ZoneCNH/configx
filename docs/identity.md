# configx 身份

## 我是谁

`configx` 是 FoundationX 的 **L1 运行时配置库**。提供显式配置加载、合并、校验、脱敏和 provenance 追踪。

> ⚠️ **身份声明**：configx 是 concrete L1 runtime library，不是模板源。模板生成属于 xlib-standard。本地 `render_template.sh` 仅用于 CI 集成测试。

## 我做什么

| 能力 | 职责 |
|------|------|
| Reader/Config/Option | 显式配置加载和合并 |
| multi-source | 多源配置（文件/环境/map/命令行） |
| schema 校验 | 配置结构体验证 |
| 环境变量覆盖 | 优先级合并 |
| provenance | 每个 key 的来源、优先级追踪 |
| Sanitize | 敏感字段脱敏 |
| Watch | 配置变更监听 |

## 我不做什么

| 不是 | 原因 |
|------|------|
| **不是 secret manager** | 密钥存储由外部 vault 管理 |
| **不是全局配置中心** | 无隐式全局状态 |
| **不是业务配置结构体** | 业务配置由各模块自己定义 |
| **不是模板源** | 模板生成属于 xlib-standard |

## 宪法合规

| 条款 | 遵循方式 |
|------|----------|
| §3.3 | L1 运行时，仅依赖 kernel |
| §4.3 | 提供 Config 结构体 + Validate() + Sanitize() |
| §6.4 | 敏感字段脱敏，使用 observex.Redactor |
