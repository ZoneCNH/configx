# configx

`configx` is an explicit, dependency-light Go configuration base library for ByteChainX services and libraries. It provides typed loaders for environment variables, env files, JSON files, and in-memory maps; deterministic source merging; struct decoding; validation hooks; and safe sanitization for logs, health output, and release evidence.

The library is intentionally explicit: callers choose every source and path. It does not discover config files automatically, create global configuration state, register singletons, import driver packages, or depend on any `x.go` module.

## Goals

- Load configuration only from caller-provided sources.
- Merge sources predictably with last-wins semantics.
- Decode into caller-owned structs using `config`, `default`, and `required` tags.
- Mark and redact secret-like keys before values are logged or serialized.
- Preserve stable base-library contracts for errors, health, metrics, tests, CI, and release evidence.

## Quick start

```go
loader := configx.NewLoader().
    AddSource(configx.NewMapSource("defaults", map[string]string{
        "APP_NAME": "service",
        "PORT":     "8080",
    })).
    AddSource(configx.NewEnvSource("APP_", []string{"NAME", "PORT", "API_TOKEN"}))

result, err := loader.Load(context.Background())
if err != nil {
    return err
}

var cfg struct {
    Name  string               `config:"NAME" required:"true"`
    Port  int                  `config:"PORT" default:"8080"`
    Token configx.SecretString `config:"API_TOKEN"`
}
if err := configx.Decode(result, &cfg); err != nil {
    return err
}

safe := result.Sanitize() // secret values are redacted
```

## Public API areas

- `NewLoader`, `Loader.AddSource`, `Loader.Load`: build and run explicit source pipelines.
- `NewEnvSource`, `NewAllEnvSource`, `NewEnvFileSource`, `NewJSONFileSource`, `NewMapSource`: concrete source adapters.
- `LoadResult`, `Value`, `SourceReport`, `SanitizedResult`: inspect loaded values and source evidence.
- `Decode`: populate caller structs from a `LoadResult`.
- `SecretString`, `NewSecretString`, `IsSecretKey`: secret handling helpers backed by `foundationx` compatibility.
- `Config`, `New`, `Close`, `HealthCheck`, `Error`, `Metrics`: baseline library contracts retained from the baselib template.

## Non-goals

- No implicit config discovery.
- No process-wide mutable configuration singleton.
- No hidden driver dependencies.
- No secret values in sanitized output.
- No dependency on `github.com/bytechainx/x.go`, `github.com/ZoneCNH/x.go`, or internal `x.go` packages.

## Commands

If this checkout is under a parent `go.work`, run validation with `GOWORK=off` to prove module independence.

```bash
GOWORK=off go test ./...
GOWORK=off go vet ./...
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off ./scripts/check_secrets.sh
```

`make lint` requires `golangci-lint`; `make security` requires `govulncheck` plus the local secret scanner. CI is expected to install those tools explicitly.

## Documentation

- [Goal](docs/goal.md): authoritative product and acceptance criteria.
- [API](docs/api.md): public configuration API and contracts.
- [Config](docs/config.md): source, merge, decode, validation, and sanitization rules.
- [Testing](docs/testing.md): unit, contract, race, boundary, security, and release evidence gates.
- [Release](docs/release.md): release manifest and evidence requirements.
