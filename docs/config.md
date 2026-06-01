# Configuration contract

`configx` 在设计上保持显式。调用方提供每个 source 和 path，然后把加载后的 values decode 到自己的 typed configuration structs。

## Source rules

1. 不做 implicit config discovery：library 永不搜索 `.env`、`config.json`、home directories 或 working directories 等 default paths。
2. 不使用 global state：loaders 是通过 `NewLoader` 创建的普通值。
3. Source order 确定：后面的 sources 覆盖前面的 sources。
4. 每次 load 都在 `SourceReport` 中记录 source evidence，且不暴露 secret values。

## Environment variables

生产路径使用 `NewEnvSource(prefix, keys)`。它只读取请求的 keys。`NewAllEnvSource` 可用于显式 bulk reads，但调用方必须 opt in 到更宽的行为。

## Files

`NewEnvFileSource(path)` 和 `NewJSONFileSource(path)` 要求调用方提供 path。它们不会推断 path names，也不会遍历 parent directories。

## Decoding

`Decode` 支持 `config`、`configx`、`default` 和 `required` struct tags。`config` tag 接受 comma-separated options，例如 `config:"DB_PASSWORD,required,secret"`；dotted keys 可以回退到 `DB_PASSWORD` 这类 uppercase env-style names。Validation errors 使用 `ErrorKindValidation`；source 和 parse failures 使用现有 typed error model。

## Secrets

类似 secret 的 keys 会按名称检测，并在 `SanitizedResult` 中 redacted。`SecretString` 保存调用方提供的 secret text，同时对 string、Go-syntax、text marshaling 和 JSON marshaling 输出执行 redaction。只应在真正需要 secret 的最终 integration boundary 使用 `Reveal`。
