# 模板生成（集成测试用）

> ⚠️ **身份声明**：`configx` 是 L1 运行时配置库（concrete library），不是模板源。
> 模板生成的标准源是 `xlib-standard`。本地 `render_template.sh` 仅用于 CI 集成测试验证，
> 确保 configx 作为 L1 模块可以被 `xlib-standard` 的 Generator 正确渲染。
> 计划在 configx v0.3 foundationx 迁移时将此基础设施移回 xlib-standard。

## 用途

`scripts/render_template.sh` 是 `xlib-standard` Generator 的本地副本，用于 CI 集成测试中验证模板渲染正确性。

## 示例

```bash
scripts/render_template.sh \
  --module-name kernel \
  --module-path github.com/ZoneCNH/kernel \
  --package-name kernel \
  --out ../kernel
```

`--out` 必须指向源码仓库之外的不存在或为空目录，避免覆盖已有仓库内容。

## 渲染范围

- `configx` 替换为 `--module-name`。
- `github.com/ZoneCNH/configx` 替换为 `--module-path`。
- `configx`、`pkg/configx` 和 `configx` imports 替换为 `--package-name`。
- 文档、Go 代码、JSON contract、shell 脚本、Makefile 和 CI 配置同步更新。

脚本只复制 Git 跟踪文件，并显式排除生成的 release evidence 与根目录 `xlib-standard.lock`。因此 `.git`、`.omc`、`.omx`、`.worktree`、`.agent/inbox`、临时目录、本地覆盖率目录和被忽略的本地分析文件不会进入下游渲染结果。生成后的库必须自己运行 release gate 生成新的发布证据 artifact。

`--enable-governance` 是完整 `xlib-standard` governance pack 的显式采用开关。当前 `configx` 尚未包含 goalcli、Docker Toolchain Runtime、governance makefile 和 hook/ruleset pack；因此该开关会在 pack 缺失时失败，避免把 partial adoption 误标成 full adoption。

## 验证

生成后至少运行：

```bash
GOWORK=off make release-check
```

模板自身的 `make integration` 会渲染三个临时下游库：

- `kernel`：L0 kernel 目标，用于证明最小基础库命名可以生成。
- `configx`：L1 自身目标，用于证明同名渲染不会破坏路径或 package。
- `redisx`：L2 downstream smoke 目标，用于证明非配置类基础库命名仍可生成。

每个临时库都会运行以下验证：

- `scripts/check_rendered_template.sh`：确认 `go.mod` module path、`pkg/<package>` 目录、旧模板目录、旧 module path、旧 smoke 目标、占位符和 `configx` 标识。
- `GOWORK=off go mod tidy` 后确认 `go.mod` / `go.sum` 不漂移。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`

这组验证用于防止生成脚本、包路径、imports、contract gate、boundary gate 和生成后证据回归。

## 生成后发布证据

生成后的库会继承 `internal/tools/releasemanifest`。该工具会生成并校验 `release/manifest/latest.json`，其中包括当前 HEAD、tree SHA、源码摘要、contract SHA256、依赖清单和工具版本。发布前应使用：

```bash
GOWORK=off make release-final-check
```

`release-final-check` 要求所有 gate 状态为 `passed`，并要求 git 工作区为 `clean`。如果只是开发中自测，`make release-check` 已足够；它允许工作区显示 `dirty`，但仍会验证 manifest 和当前源码内容一致。

## 边界

生成后的基础库仍必须保持独立，不能依赖 `github.com/bytechainx/x.go`、`github.com/ZoneCNH/x.go` 或任何 `x.go/internal/*` 包。
