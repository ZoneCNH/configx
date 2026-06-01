# configx API

`configx` 暴露显式配置加载原语，并保留已应用模板继承的标准基础库 contract。

## 显式加载

- `NewLoader(opts ...LoaderOption) *Loader` 创建隔离的 loader。loader 不持有进程级全局状态，也不会在加入 source 前执行 discovery。
- `(*Loader).AddSource(Source) *Loader` 追加调用方提供的 source。
- `(*Loader).Load(context.Context) (LoadResult, error)` 按顺序加载每个 source。后面的值覆盖前面的值；先前的 `Value` 会标记为 `Overridden`。
- `WithFailFast(bool)` 控制 source error 是否立即停止加载。

## Sources

- `NewEnvSource(prefix string, keys []string, opts ...SourceOption)` 在应用 prefix 后只读取指定 key。这是 environment 使用场景的安全默认方式。
- `NewAllEnvSource(prefix string, opts ...SourceOption)` 读取所有匹配的 environment variables，并且必须显式 opt-in。
- `NewEnvFileSource(path string, opts ...SourceOption)` 读取调用方提供的 dotenv-style file path。
- `NewJSONFileSource(path string, opts ...SourceOption)` 读取调用方提供的 JSON file path，并用点号展开嵌套 key。
- `NewMapSource` 和 `NewSecretMapSource` 支持 tests 与 embedded defaults。

每个 source 都通过 `SourceReport` 报告 `Name`、`Kind`、可选 `Path`、已加载 key 和 sanitized errors。

## Decode 与 validation

`Decode(result, &target)` 根据 `config` tags 填充导出的 struct fields。支持的 tags：

- `config:"KEY"`：`LoadResult` 中的 key name。
- `default:"value"`：key 缺失时使用的 default。
- `required:"true"`：key 缺失时产生 validation error。
- `config:"-"`：跳过该 field。

支持的 field types 包括 strings、booleans、有符号和无符号 integers、floats、`time.Duration`、`SecretString`，以及实现 `encoding.TextUnmarshaler` 的类型。如果 target 实现 `Validate() error`，`Decode` 会在字段赋值后运行它。

## Sanitization

`LoadResult.Sanitize()` 返回 `SanitizedResult`，其中 secret values 会脱敏为 `***`。包含 secret、password、passwd、token、access_key 或 secret_key 的 keys 会被视为 secrets；`NewSecretMapSource` 可以显式标记额外 key。`SecretString.String()` 与 text marshaling 均返回脱敏输出。

## Baseline contracts

仓库也保留模板中的 baseline contracts：

- `Config`、`Validate` 和 `Sanitize` 用于最小显式 config validation。
- `New`、`Close` 和 `HealthCheck` 用于 lifecycle 与 health contract tests。
- `Error`、`ErrorKind`、`NewError`、`WrapError` 和 `IsKind` 用于稳定的 typed errors。
- `Metrics` hooks 和名称由 `contracts/metrics.md` 锁定。
- `Version` 和 `ModuleName` 用于 release evidence。

包不得 import `x.go`，不得创建全局 config state，也不得添加 driver dependencies。
