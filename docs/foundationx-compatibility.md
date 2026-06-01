# foundationx compatibility boundary

`configx` depends on `github.com/bytechainx/foundationx` through the local
replacement module in `internal/foundationx`. This keeps the config library
buildable while preserving the public `foundationx.SecretString` contract used
by callers.

## Compatibility scope

The local module intentionally mirrors only the foundation API required by
`configx`:

- `SecretString`, `NewSecretString`, `Reveal`, redacted string/text/JSON
  formatting, `Sanitize`, and `IsZero`
- minimal typed errors used by configx error wrapping

Reference modules such as `/tmp/configx-foundationx` also include health,
lifecycle, retry, clock, and version contracts. Those APIs are not re-exported
or used by `configx` unless a configx feature imports them and adds matching
tests. This prevents template drift from pulling infrastructure behavior into a
base configuration library.

## Non-negotiable boundaries

- `docs/goal.md` remains the authoritative contract and must not be rewritten by
  template application work.
- Config loading stays explicit: callers create loaders and pass every source or
  path. The library must not auto-discover `.env`, `production.yaml`,
  `config.local.yaml`, or `/home/k8s/secrets/env/*`.
- The module must not contain generated `x.go` files and must not import `x.go`
  or infrastructure driver packages such as Redis, Kafka, PostgreSQL, TDengine,
  or object-storage SDKs.
- Validation evidence, examples, release manifests, and documentation must use
  sanitized secret output only.

## Upgrade rule

When replacing the local module with an upstream foundationx release, first
prove that `SecretString` redaction, `Sanitize`, `IsZero`, JSON/text marshaling,
and configx error behavior remain unchanged with `GOWORK=off go test ./...` and
the boundary, contract, and secret scanners.
