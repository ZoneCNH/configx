# configx 发布证据

`configx` 的 release 或完成声明必须写成 `DONE with evidence:`，并包含最新本地或 CI 结果。

## 必需 gate

发布前需要满足：

- 格式化与 vetting 通过
- 已修改 Go files 的 lint 通过
- 加载器、merge precedence、decode、validation 与 sanitization 的 unit tests 通过
- 如果支持并发，带共享 reader 的 loader/result 路径 race tests 通过
- 配置 schema、error envelope、current-state/security/redaction docs 与 release manifest 的 contract checks 通过
- 对 source、docs、tests 与 generated evidence 执行 secret scan
- 引入 dependencies 后执行 vulnerability scan

## 证据清单

`release/manifest/template.json` 定义所需发布证据结构。生成的 manifests 只能包含脱敏数据，不得包含原始 config values；`docs/current-state.md` 记录生成证据前应复跑的本地 gate。

必需 manifest fields 包括 module path、version、commit、tree state、checks、contract hashes、dependencies、tools、artifacts 与已知风险。

## 持续集成预期

CI 应运行与本地发布检查相同的 gates。缺少 `golangci-lint` 或 `govulncheck` 等必需工具时，应让相关 gate 失败，而不是静默跳过。
