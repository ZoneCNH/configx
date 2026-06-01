# configx 模板应用复盘

评审已应用的 baselib template 与 `configx` goal 是否一致时，使用本检查清单。

## 2026-06-01 模板与 foundation 复盘记录

- `docs/goal.md` 保留为权威来源；模板应用必须更新支撑文档与 harness checks，但不得覆盖它。
- `internal/foundationx` 是窄兼容模块，不是 `/tmp/configx-foundationx` 的完整副本；参见 `docs/foundationx-compatibility.md`。
- `configx` 保持显式加载语义：只有在调用方传入具体 source/path 时，才读取 env files、JSON files 与 `/home/k8s/secrets/env/*` paths。
- 边界校验必须在发布证据被接受前拒绝生成的 `x.go` files、`x.go` dependencies 与 infrastructure driver dependencies。
- 密钥证据只能使用脱敏 values；不得在 validation logs、examples、manifests 或 documentation output 中调用 `Reveal()`。

## 边界检查

- [ ] `go.mod` module 是 `github.com/ZoneCNH/configx`。
- [ ] 公共 package 名称是 `configx`。
- [ ] 不导入 `x.go` 或 service driver packages。
- [ ] 不存在 package-level mutable config、singleton client 或隐式初始化。
- [ ] 不隐式读取 `.env`、`production.yaml`、`config.local.yaml` 或 `/home/k8s/secrets/env/*`。

## 接口检查

- [ ] Sources 由调用方创建并显式排序。
- [ ] Env file 和 JSON file loaders 要求调用方提供路径。
- [ ] Merge result 记录 effective values 的 source trace。
- [ ] Decode 支持 `config`、`default`、`required` 和 `secret` tags，或记录等价能力。
- [ ] 校验错误能标识 fields，且不泄漏 secret values。

## 密钥安全检查

- [ ] Secret 值在 stringers、errors、logs、tests、examples 与 release manifests 中脱敏。
- [ ] Secret tests 包含原始 secret material 不存在的 negative assertions。
- [ ] 发布证据只记录 sanitized contract hashes 与 tool output。

## 文档检查

- [ ] README 指向 `docs/goal.md` 作为权威来源。
- [ ] Examples 只展示 explicit 调用方持有的 paths。
- [ ] Release documentation 要求 evidence 与 secret scanning。
