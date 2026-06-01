# configx template application review

评审已应用的 baselib template 与 `configx` goal 是否一致时，使用本 checklist。

## 2026-06-01 template/foundation review notes

- `docs/goal.md` 保留为 source of truth；template application 必须更新 supporting docs 和 harness checks，但不得覆盖它。
- `internal/foundationx` 是窄 compatibility module，不是 `/tmp/configx-foundationx` 的完整副本；参见 `docs/foundationx-compatibility.md`。
- `configx` 保持 explicit loading semantics：只有在调用方传入具体 source/path 时，才读取 env files、JSON files 和 `/home/k8s/secrets/env/*` paths。
- Boundary validation 必须在 release evidence 被接受前拒绝生成的 `x.go` files、`x.go` dependencies 和 infrastructure driver dependencies。
- Secret evidence 只能使用 redacted values；不得在 validation logs、examples、manifests 或 documentation output 中调用 `Reveal()`。

## Boundary checks

- [ ] `go.mod` module 是 `github.com/bytechainx/configx`。
- [ ] Public package 名称是 `configx`。
- [ ] 不 import `x.go` 或 service driver packages。
- [ ] 不存在 package-level mutable config、singleton client 或 implicit initialization。
- [ ] 不 implicit read `.env`、`production.yaml`、`config.local.yaml` 或 `/home/k8s/secrets/env/*`。

## API checks

- [ ] Sources 由调用方创建并显式排序。
- [ ] Env file 和 JSON file loaders 要求调用方提供 paths。
- [ ] Merge result 记录 effective values 的 source trace。
- [ ] Decode 支持 `config`、`default`、`required` 和 `secret` tags，或记录等价能力。
- [ ] Validation errors 能标识 fields，且不泄漏 secret values。

## Secret-safety checks

- [ ] Secret values 在 stringers、errors、logs、tests、examples 和 release manifests 中 redacted。
- [ ] Secret tests 包含 raw secret material 不存在的 negative assertions。
- [ ] Release evidence 只记录 sanitized contract hashes 和 tool output。

## Documentation checks

- [ ] README 指向 `docs/goal.md` 作为 authoritative source。
- [ ] Examples 只展示 explicit caller-owned paths。
- [ ] Release documentation 要求 evidence 和 secret scanning。
