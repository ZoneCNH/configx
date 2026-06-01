# configx examples

These examples describe the intended caller-owned usage. They avoid package globals and implicit discovery.

## Explicit env file path

```go
ctx := context.Background()
loader := configx.NewLoader()

result, err := loader.AddSource(configx.NewEnvFileSource("/home/k8s/secrets/env/postgres.env")).Load(ctx)
if err != nil {
    return err
}

var cfg PostgresConfig
if err := configx.Decode(result, &cfg); err != nil {
    return err
}
```

The path is supplied by the application. `configx` must not search this directory on its own.

## Map override precedence

```go
loader := configx.NewLoader()
result, err := loader.
    AddSource(configx.NewMapSource("defaults", defaults)).
    AddSource(configx.NewMapSource("runtime", runtimeOverrides)).
    Load(ctx)
```

Later sources have higher precedence when the implementation chooses this ordering convention; the merge trace must make the chosen value and source visible.

## Sanitized diagnostics

```go
safe := configx.Sanitize(result)
logger.Info("config loaded", "config", safe)
```

Secrets must render as redacted placeholders, never as raw credential material.
