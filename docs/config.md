# Configuration contract

`configx` is explicit by design. Callers provide every source and path, then decode the resulting values into their own typed configuration structs.

## Source rules

1. No implicit config discovery: the library never searches default paths such as `.env`, `config.json`, home directories, or working directories.
2. No global state: loaders are ordinary values created with `NewLoader`.
3. Source order is deterministic: later sources override earlier sources.
4. Every load records source evidence in `SourceReport` without exposing secret values.

## Environment variables

Use `NewEnvSource(prefix, keys)` for production paths. It reads only the requested keys. `NewAllEnvSource` is available for explicit bulk reads, but callers must opt in to that broader behavior.

## Files

`NewEnvFileSource(path)` and `NewJSONFileSource(path)` require a caller-provided path. They do not infer path names or walk parent directories.

## Decoding

`Decode` supports `config`, `configx`, `default`, and `required` struct tags. The `config` tag accepts comma-separated options such as `config:"DB_PASSWORD,required,secret"` and dotted keys can fall back to uppercase env-style names such as `DB_PASSWORD`. Validation errors use `ErrorKindValidation`; source and parse failures use the existing typed error model.

## Secrets

Secret-like keys are detected by name and redacted in `SanitizedResult`. `SecretString` stores caller-provided secret text while redacting string, Go-syntax, text marshaling, and JSON marshaling output. Use `Reveal` only at the final integration boundary that actually needs the secret.
