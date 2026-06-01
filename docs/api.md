# configx API

`configx` exposes explicit configuration loading primitives plus the standard base-library contracts carried forward from the applied template.

## Explicit loading

- `NewLoader(opts ...LoaderOption) *Loader` creates an isolated loader. A loader has no process-global state and performs no discovery until sources are added.
- `(*Loader).AddSource(Source) *Loader` appends a caller-provided source.
- `(*Loader).Load(context.Context) (LoadResult, error)` loads each source in order. Later values override earlier values; the previous `Value` is marked `Overridden`.
- `WithFailFast(bool)` controls whether source errors stop loading immediately.

## Sources

- `NewEnvSource(prefix string, keys []string, opts ...SourceOption)` reads only the named keys after applying the prefix. This is the safe default for environment use.
- `NewAllEnvSource(prefix string, opts ...SourceOption)` reads all matching environment variables and is intentionally opt-in.
- `NewEnvFileSource(path string, opts ...SourceOption)` reads a caller-provided dotenv-style file path.
- `NewJSONFileSource(path string, opts ...SourceOption)` reads a caller-provided JSON file path and flattens nested keys with dots.
- `NewMapSource` and `NewSecretMapSource` support tests and embedded defaults.

Every source reports `Name`, `Kind`, optional `Path`, loaded keys, and sanitized errors through `SourceReport`.

## Decode and validation

`Decode(result, &target)` fills exported struct fields from `config` tags. Supported tags:

- `config:"KEY"`: key name in the `LoadResult`.
- `default:"value"`: default used when the key is missing.
- `required:"true"`: missing key is a validation error.
- `config:"-"`: skip the field.

Supported field types include strings, booleans, signed and unsigned integers, floats, `time.Duration`, `SecretString`, and types implementing `encoding.TextUnmarshaler`. If the target implements `Validate() error`, `Decode` runs it after field assignment.

## Sanitization

`LoadResult.Sanitize()` returns a `SanitizedResult` with secret values redacted as `***`. Keys containing secret, password, passwd, token, access_key, or secret_key are treated as secrets; `NewSecretMapSource` can mark additional keys explicitly. `SecretString.String()` and text marshaling are redacted.

## Baseline contracts

The repository also preserves baseline contracts from the template:

- `Config`, `Validate`, and `Sanitize` for minimal explicit config validation.
- `New`, `Close`, and `HealthCheck` for lifecycle and health contract tests.
- `Error`, `ErrorKind`, `NewError`, `WrapError`, and `IsKind` for stable typed errors.
- `Metrics` hooks and names locked by `contracts/metrics.md`.
- `Version` and `ModuleName` for release evidence.

The package must not import `x.go`, create global config state, or add driver dependencies.
