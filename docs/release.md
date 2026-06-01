# configx 发布证据（release evidence）

`configx` 的 release 或 completion claim 必须写成 `DONE with evidence:`，并包含最新 local 或 CI results。

## 必需 gate（Required gates）

Release 前需要满足：

- 格式化（formatting）与 vetting 通过
- 已修改 Go files 的 lint 通过
- 加载器（loaders）、merge precedence、decode、validation 与 sanitization 的 unit tests 通过
- 如果支持 concurrency，loader/result paths with shared readers 的 race tests 通过
- 配置 schema（config schema）、error envelope 与 release manifest 的 contract checks 通过
- 对 source、docs、tests 与 generated evidence 执行 secret scan
- 引入 dependencies 后执行 vulnerability scan

## 证据清单（Evidence manifest）

`release/manifest/template.json` 定义所需 release evidence shape。Generated manifests 只能包含 sanitized data，不得包含原始 config values。

必需 manifest fields 包括 module path、version、commit、tree state、checks、contract hashes、dependencies、tools、artifacts 与 known risks。

## 持续集成预期（CI expectation）

CI 应运行与 local release checks 相同的 gates。缺少 `golangci-lint` 或 `govulncheck` 等 required tools 时，应让相关 gate 失败，而不是 silent skip。
