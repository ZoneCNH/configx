# configx current state evidence

Date: 2026-06-04

This snapshot records the implemented `configx` v0.1 MVA surface and the boundaries that must remain true before a release/completion claim.

## Implemented MVA surface

- **Explicit sources only**: callers construct sources with `NewMapSource`, `NewSecretMapSource`, `NewEnvSource`, `NewAllEnvSource`, `NewEnvFileSource`, `NewJSONFileSource`, `NewYAMLFileSource`, `NewTOMLFileSource`, or the matching single-source loader helpers.
- **No implicit discovery**: the library has no default search path for `.env`, `production.yaml`, `config.local.yaml`, `/home/k8s/secrets/env/*`, or generated `x.go` config.
- **Ordered merge**: `Loader.AddSource` preserves caller order; later source values become effective and earlier values are marked overridden in the result trace.
- **Report-only metadata**: `SourceReport`, value traces, sanitized results, and release manifests record names, kinds, paths, key names, hashes, and check statuses, never raw secret values.
- **Decode support**: `Decode` handles exported nested structs, strings, bools, integers, unsigned integers, floats, `time.Duration`, `SecretString`, `encoding.TextUnmarshaler`, `config`, `default`, `required`, and `config:"-"` tags, then invokes `Validate() error` when present.
- **Redaction primitives**: `SecretString` string/text/JSON output is redacted; `LoadResult.Sanitize()` and `SanitizedResult` are the supported output path for logs, reports, tests, examples, and release evidence.
- **Harness gates**: `make boundary`, `make contracts`, `make golden`, `make fuzz-smoke`, `make evidence`, `make release-check`, and `make release-preflight` are present as release/completion gates.

## Boundary contract

`configx` is a small L1 library. It must not become an L2 provider/client layer. The committed boundary scripts reject:

- generated `x.go` files or `github.com/bytechainx/x.go` / `github.com/ZoneCNH/x.go` dependencies;
- automatic reads of implicit config locations in `pkg`, `internal`, `contracts`, or `examples`;
- cloud KMS/Vault/Consul/Etcd/Nacos and infrastructure driver dependencies;
- business-domain terms that would couple the foundation library to downstream trading or runtime products.

## Current evidence commands

Use these commands from the repository root with `GOWORK=off`:

```sh
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off make golden
GOWORK=off make fuzz-smoke
GOWORK=off make evidence
```

For release evidence, generate `release/manifest/latest.json` as an uncommitted artifact only after the required gates have run with explicit `*_STATUS=passed` inputs.

## Known v0.1 naming notes

The implementation exposes `LoadResult` and `SourceReport` as the public result/report types. If future plans introduce aliases named `Result` or `MergeReport`, they must preserve the same no-raw-secret and explicit-source semantics before replacing the current names in docs or contracts.
