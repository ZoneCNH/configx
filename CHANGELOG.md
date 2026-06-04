# 变更日志

## v0.1.3 - 2026-06-04

### 重构

- 拆分 `core.go`（667 行）为 8 个职责文件：`loader.go`、`source.go`、`source_env.go`、`source_file.go`、`source_map.go`、`merge.go`、`secret.go`、`result.go`。
- 内联 `internal/sanitize` 和 `internal/validation` 到 `pkg/configx` 作为私有函数（单函数包反模式修复）。
- 最大文件行数从 667 降至 265。

### 新增

- 测试覆盖率从 75.8% 提升至 97.1%，新增 93 个测试函数。
- 新增 6 个 benchmark 测试（`core_bench_test.go`）。
- 新增 3 个 ADR 文档：last-wins merge strategy、no global state、explicit config loading。
- 新增 `examples/error-handling/` 示例，覆盖 5 种错误处理模式。
- 增强 golangci-lint 配置，新增 5 个 linter：`errcheck`、`gosec`、`unconvert`、`unparam`、`misspell`。

### 修复

- 修复 7 个 gosec 告警（文件路径清理 `filepath.Clean`、目录权限 0o750、文件权限 0o600）。

### 验证

- `GOWORK=off go test ./...` 全部通过（8 packages）。
- `GOWORK=off go vet ./...` 零告警。
- `golangci-lint run ./...` 零告警。
- 覆盖率 97.1%。

## 未发布

### 新增

- 新增显式 TOML 与 YAML/YML 文件 source，并提供 `LoadTOMLFile`、`LoadYAMLFile` convenience loaders。

### 治理

- 扩展配置 source 文档与回归测试，锁定嵌套 key 展开、last-wins merge、source report path/value keys 和 secret redaction 行为。

## v0.1.1 - 2026-06-01

### 治理

- 升级 GitHub Actions 官方 action 到 Node.js 24 运行时版本，消除 Node.js 20 deprecation 注记。
- 固定 workflow 安装的 `govulncheck` 版本，避免 `@latest` 漂移到要求更高 Go toolchain 的版本。

## v0.1.0 - 2026-06-01

### 新增

- 新增 `make release-preflight VERSION=vX.Y.Z`，在打 tag 前检查版本、`main` 同步状态、目标 tag、`CHANGELOG.md`、必需工具和最终 release gate。

### 修复

- 发布检查 workflow 在运行 `make release-check` 前安装 `golangci-lint` 和 `govulncheck`，并使用 `GOWORK=off`，与 CI 的强制 gate 环境保持一致。

### 新增

- 初始化 `configx` 结构。
- 添加标准 Go 基础库包骨架。
- 添加 Makefile 命令。
- 添加 Harness Gate 脚本。
- 添加 GitHub Actions 工作流。
- 添加 contracts 文件。
- 添加 Agent 运行时模板。
- 添加 release manifest 模板。
- 添加 typed error、错误包装和 `ErrorKind` contract。
- 添加 client 生命周期、健康检查和请求扩展 metrics contract。
- 添加 health JSON contract 与 contracts 回归测试。
- 添加 config schema 到 `Config` 字段映射的 contract 回归测试。
- 添加 `scripts/render_template.sh`，支持生成 `baselibx` 等具体基础库。
- 添加 `examples/basic`、`examples/config` 和 `examples/health` smoke 测试，锁定文档示例输出。
- 添加 `testkit` 夹具和断言回归测试。
- 添加配置属性测试、配置 fuzz smoke 测试、健康状态 golden 测试和 `testkit` golden 文件工具。

### 安全

- 添加 Secret Gate。
- `make security` 强制运行 `govulncheck ./...` 和密钥扫描；缺少 `govulncheck` 时必须失败。
- 配置脱敏规则覆盖发布证据和日志可见内容。
- 边界 Gate 同时拦截 `github.com/bytechainx/x.go` 和 `github.com/ZoneCNH/x.go`。

### 治理

- 添加证据和复盘模板。
- CI 流程在 `make ci` 前安装 `golangci-lint` 和 `govulncheck`，与 Makefile 强制 gate 对齐。
- `make release-check` 统一执行 CI、integration 和 manifest 生成。
- `make release-final-check` 在发布前串联 `release-check`、发布证据校验和工作区洁净校验。
- `make integration` 通过临时 `baselibx` 和 `corekit` 渲染、测试、contracts、boundary 与发布证据生成验证模板链路。
- `release/manifest/latest.json` 作为生成产物保留在源码历史之外，避免发布证据与源码提交互相污染。

### 验证

- 发布前已运行 `VERSION=v0.1.0 GOWORK=off make release-preflight`。
- `go fmt ./...`、`go vet ./...`、`golangci-lint run ./...`、`go test ./...`、`go test -race ./...`、Boundary、Security、contracts、integration 和发布证据校验均通过。
- `v0.1.0` 为 annotated tag，指向本次发布提交。
