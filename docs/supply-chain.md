# 供应链与证据

## 目标

本模板的发布证据不是普通构建日志，而是可重复校验的发布事实清单。它必须回答三个问题：

- 当前发布对应哪个 Go module、commit 和 git tree。
- 当前源码、contract 文件和依赖清单是否与 manifest 一致。
- 必需 gate 是否全部以 `passed` 状态完成。

## 清单生成

`make evidence` 调用 `scripts/generate_manifest.sh`，最终由 `internal/tools/releasemanifest` 生成 `release/manifest/latest.json`。生成内容包括：

- `commit` 和 `tree_sha`：来自当前 git HEAD。
- `source_digest` 和 `tracked_file_count`：来自 `git ls-files` 中所有受跟踪文件的路径和内容摘要。
- `contracts`：核心 contract 文件的 SHA256 指纹。
- `dependencies`：`go list -m -json all` 的模块清单。
- `tools`：Go、`golangci-lint` 和 `govulncheck` 的版本或可用状态。
- `checks`：`fmt`、`vet`、`lint`、测试、race、boundary、secret scan、security、contract 和 integration gate 状态。

`release/manifest/latest.json` 是生成产物，不提交源码历史；`release/manifest/template.json` 只保留字段模板。

## 清单校验

`make release-evidence-check` 会重新读取当前仓库事实，并校验：

- 清单（manifest）的 module、commit、tree SHA、源码摘要和受跟踪文件数量与当前仓库一致。
- 契约（contract）指纹和依赖清单与当前 Go module 解析结果一致。
- 必需 check 均存在，且在 release gate 中必须为 `passed`。
- 产物（artifact）列表包含 `release/manifest/latest.json`。

`make release-final-check` 在上述校验之外要求 `tree_state=clean`。正式发布、打 tag 或交付给下游基础库前必须使用该入口。

## CI 产物

GitHub Actions 运行 `GOWORK=off make release-check`，并上传 `release/manifest/latest.json` 作为 `release-manifest` artifact。CI 中上传的 artifact 是发布证据的外部留痕；本地生成的 `latest.json` 只用于验证和排障。

## 下游模板安全线

`make integration` 会按 `xlib-standard` 当前 smoke matrix 渲染 `kernel`、`configx` 和 `redisx` 三个临时下游库，检查旧模板标识是否清空，并在下游库内生成、校验发布证据。这保证模板替换逻辑、contract gate、boundary gate 和证据工具不会只在模板仓库自身成立。

当前仓库的标准状态记录在根目录 `xlib-standard.lock`。该 lock 表示 L1 partial adoption；goalcli governance surface 和 Docker Toolchain Runtime 尚未纳入本仓库默认 gate，后续如启用必须同时补齐对应 hooks、rulesets、makefile 和 harness 文件。
