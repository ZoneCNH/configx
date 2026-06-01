# 配置契约

`configx` 在设计上保持显式。调用方提供每个 source 和路径，然后把加载后的 values 解码到自己的类型化配置结构体。

## 来源规则

1. 不做隐式配置发现：库永不搜索 `.env`、`config.json`、home directory、working directory 或其他默认路径。
2. 不使用全局状态：loaders 是通过 `NewLoader` 创建的普通值。
3. Source 顺序确定：后面的 sources 覆盖前面的 sources。
4. 每次 load 都在 `SourceReport` 中记录 source 证据，且不暴露 secret 值。

## 环境变量

生产路径使用 `NewEnvSource(prefix, keys)`。它只读取请求的 keys。`NewAllEnvSource` 可用于显式批量读取，但调用方必须 opt in 到更宽的行为。

## 文件

`NewEnvFileSource(path)` 和 `NewJSONFileSource(path)` 要求调用方提供路径。它们不会推断路径名称，也不会遍历父目录。

## 解码

`Decode` 支持 `config`、`configx`、`default` 和 `required` 结构体 tag。`config` tag 接受逗号分隔 options，例如 `config:"DB_PASSWORD,required,secret"`；dotted key 可以回退到 `DB_PASSWORD` 这类大写 env-style name。校验错误使用 `ErrorKindValidation`；source 与解析失败使用现有类型化错误模型。

## 密钥

类似 secret 的 keys 会按名称检测，并在 `SanitizedResult` 中脱敏。`SecretString` 保存调用方提供的密钥文本，同时对 string、Go-syntax、text marshaling 和 JSON marshaling 输出执行脱敏。只应在真正需要 secret 的最终集成边界使用 `Reveal`。
