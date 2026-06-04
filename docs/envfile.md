# Env file loading

Env file loading is allowed only through an explicit caller-provided path or reader. `configx` must not automatically search `.env`, `production.yaml`, `config.local.yaml`, or `/home/k8s/secrets/env/*`.

Use `NewEnvFileSource(path)` or `LoadEnvFile(ctx, path)` when the application already resolved the path. The path is provenance metadata, not a discovery hint.

Env file examples and tests must use temporary files or fake paths and must not read production secret directories.
