# configx 契约（contracts）

`configx` 将配置加载变成显式、可测试、可 sanitizable 的 runtime contract。当本摘要与完整目标不一致时，`docs/goal.md` 仍是权威来源。

## 公共 API 契约（contract）

公共 package 是 `configx`。实现应暴露小型、可组合的 types，而不是 global state：

- `Source`：带 source metadata 的命名配置输入。
- `LoadEnv`、`LoadEnvFile(path)`、`LoadJSONFile(path)` 和 `LoadMap(map[string]string)` 风格的 constructors 或等价 API。
- `Loader`：由调用方创建、接收有序 sources 并返回 `LoadResult` 的 loader。
- `LoadResult`：合并后的 values，以及每个 effective key 的 source trace 记录。
- `Decode`：使用 `config`、`default`、`required` 和 `secret` tags 的 struct decoding。
- `Validator`：decoded configs 的显式 validation hook。
- `SecretString`：可安全表示 secret value 的类型，可在可用时与 `foundationx` 集成。
- `Sanitize`：面向 logs、errors、tests、release evidence 与可读输出的稳定 redaction。

## 来源契约（Source contract）

允许的 sources 必须显式且由调用方拥有：

- 调用方请求的 process environment
- 调用方传入 path 指向的 env file
- 调用方传入 path 指向的 JSON file
- 调用方传入的 in-memory map

禁止的行为：

- 自动发现 `.env`、`config.local.yaml` 或 `production.yaml`
- 在调用方未传入具体 path 时读取 `/home/k8s/secrets/env/*`
- 在 package-level mutable state 中保留隐式 defaults
- 导入 `x.go` 或 service driver packages

## 合并与追踪契约（Merge 与 trace contract）

Merging 必须是 deterministic。Source order 是显式的，result 会记录每个 key 最终值来自哪个 source。该 trace 只有在 secret values 完成 sanitization 后才能用于排障。

## 校验与错误（Validation 与 errors）

Validation errors 必须稳定且可分类。Errors 必须包含足够的 field/source context，帮助修复无效配置，同时不得包含原始 secret values。

## 密钥契约（Secret contract）

带 secret 的字段默认在以下输出中 redacted：

- `String` / `GoString` 风格表示
- 错误消息（error messages）
- 日志（logs）与 structured diagnostic maps
- 测试输出（test output）与 golden files
- 发布 manifests（release manifests）与 evidence artifacts

使用 `contracts/config.schema.json` 锁定外部 config shape，使用 `contracts/error.schema.json` 锁定 public error envelope。
