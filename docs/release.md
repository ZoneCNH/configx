# configx release evidence

`configx` 的 release 或 completion claim 必须写成 `DONE with evidence:`，并包含最新 local 或 CI results。

## Required gates

Release 前需要满足：

- formatting 和 vetting 通过
- modified Go files 的 lint 通过
- loaders、merge precedence、decode、validation 和 sanitization 的 unit tests 通过
- 如果支持 concurrency，loader/result paths with shared readers 的 race tests 通过
- config schema、error envelope 和 release manifest 的 contract checks 通过
- 对 source、docs、tests 和 generated evidence 执行 secret scan
- 引入 dependencies 后执行 vulnerability scan

## Evidence manifest

`release/manifest/template.json` 定义所需 release evidence shape。Generated manifests 只能包含 sanitized data，不得包含 raw config values。

Required manifest fields 包括 module path、version、commit、tree state、checks、contract hashes、dependencies、tools、artifacts 和 known risks。

## CI expectation

CI 应运行与 local release checks 相同的 gates。缺少 `golangci-lint` 或 `govulncheck` 等 required tools 时，应让相关 gate 失败，而不是 silent skip。
