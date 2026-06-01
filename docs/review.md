# configx template application review

Use this checklist when reviewing the applied baselib template against the `configx` goal.

## Boundary checks

- [ ] `go.mod` module is `github.com/bytechainx/configx`.
- [ ] Public package is named `configx`.
- [ ] No imports of `x.go` or service driver packages.
- [ ] No package-level mutable config, singleton client, or implicit initialization.
- [ ] No implicit reads of `.env`, `production.yaml`, `config.local.yaml`, or `/home/k8s/secrets/env/*`.

## API checks

- [ ] Sources are caller-created and explicitly ordered.
- [ ] Env file and JSON file loaders require caller-provided paths.
- [ ] Merge result records source trace for effective values.
- [ ] Decode supports `config`, `default`, `required`, and `secret` tags or documented equivalents.
- [ ] Validation errors identify fields without leaking secret values.

## Secret-safety checks

- [ ] Secret values are redacted in stringers, errors, logs, tests, examples, and release manifests.
- [ ] Secret tests include negative assertions that raw secret material is absent.
- [ ] Release evidence records sanitized contract hashes and tool output only.

## Documentation checks

- [ ] README points to `docs/goal.md` as authoritative.
- [ ] Examples show explicit caller-owned paths only.
- [ ] Release documentation requires evidence and secret scanning.
