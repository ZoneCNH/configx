# configx template application review

Use this checklist when reviewing the applied baselib template against the `configx` goal.

## 2026-06-01 template/foundation review notes

- `docs/goal.md` is preserved as the source of truth; template application must
  update supporting docs and harness checks without overwriting it.
- `internal/foundationx` is a narrow compatibility module, not a full copy of
  `/tmp/configx-foundationx`; see `docs/foundationx-compatibility.md`.
- `configx` keeps explicit loading semantics: env files, JSON files, and
  `/home/k8s/secrets/env/*` paths are read only when the caller passes a
  concrete source/path.
- Boundary validation must reject generated `x.go` files, `x.go` dependencies,
  and infrastructure driver dependencies before release evidence is accepted.
- Secret evidence must use redacted values only; do not call `Reveal()` in
  validation logs, examples, manifests, or documentation output.

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
