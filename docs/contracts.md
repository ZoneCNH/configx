# configx contracts

`configx` turns configuration loading into an explicit, testable, sanitizable runtime contract. `docs/goal.md` remains authoritative when this summary and the full goal differ.

## Public API contract

The public package is `configx`. The implementation should expose small, composable types rather than global state:

- `Source`: named configuration input with source metadata.
- `LoadEnv`, `LoadEnvFile(path)`, `LoadJSONFile(path)`, and `LoadMap(map[string]string)` style constructors or equivalents.
- `Loader`: caller-created loader that accepts ordered sources and returns a `LoadResult`.
- `LoadResult`: merged values plus source trace for every effective key.
- `Decode`: struct decoding with `config`, `default`, `required`, and `secret` tags.
- `Validator`: explicit validation hook for decoded configs.
- `SecretString`: a safe secret value type integrated with `foundationx` when available.
- `Sanitize`: stable redaction for logs, errors, tests, release evidence, and human-readable output.

## Source contract

Allowed sources are explicit and caller-owned:

- process environment requested by the caller
- env file at a path passed by the caller
- JSON file at a path passed by the caller
- in-memory map passed by the caller

Disallowed behavior:

- auto-discovering `.env`, `config.local.yaml`, or `production.yaml`
- reading `/home/k8s/secrets/env/*` unless the caller passes a concrete path
- retaining implicit defaults in package-level mutable state
- importing `x.go` or service driver packages

## Merge and trace contract

Merging must be deterministic. The source order is explicit, and the result records which source supplied the final value for each key. This trace is safe for troubleshooting only after secret values are sanitized.

## Validation and errors

Validation errors must be stable and classifiable. Errors must include enough field/source context to fix invalid configuration without including raw secret values.

## Secret contract

Secret-bearing fields are redacted by default in:

- `String` / `GoString` style representations
- error messages
- logs and structured diagnostic maps
- test output and golden files
- release manifests and evidence artifacts

Use `contracts/config.schema.json` to lock external config shape and `contracts/error.schema.json` to lock the public error envelope.
