# 仓库指南

## 项目结构与模块组织

本仓库实现 `github.com/bytechainx/configx`，并按基础库模板组织：

- `pkg/configx/` 是对外公共 API。
- `internal/` 放私有实现；`internal/foundationx/` 是本仓库用于兼容 `github.com/bytechainx/foundationx` 的本地 replace 模块。
- `contracts/` 放契约 schema、契约测试和指标契约。
- `examples/` 放可编译示例。
- `testkit/` 放测试辅助包。
- `scripts/` 放模板、边界、契约、安全、release evidence 等检查脚本。
- `docs/` 放目标、设计、ADR、API、测试、发布和供应链文档。

测试文件与被测代码相邻，命名为 `*_test.go`。不要提交本地 `.omx/` runtime state、临时目录或生成的 release evidence，除非 release 流程明确要求。

## 构建、测试与开发命令

- `go mod tidy` - 同步 module dependencies。
- `GOWORK=off go test ./...` - 运行全部单元测试，避免受外层 workspace 影响。
- `GOWORK=off go test -race ./...` - 运行 race 测试。
- `GOWORK=off go vet ./...` - 运行 Go static checks。
- `GOWORK=off GOLANGCI_LINT_CACHE=/tmp/configx-golangci-cache make ci` - 运行格式化、vet、lint、测试、race、boundary、security、contracts。
- `GOWORK=off make integration` - 渲染 foundationx/corekit 模板并运行集成检查。

新增脚本必须提交并记录用途。脚本应能从仓库根目录运行，必要时自行 `cd` 到根目录。

## 编码风格与命名约定

使用 `gofmt` 保持 idiomatic Go formatting。Package names 应短小、全小写并具备描述性，例如 `configx`、`validation` 或 `testkit`。公共 API 的 exported identifiers 需要简洁注释。优先使用显式构造函数，避免全局可变状态。

`configx` 的公共错误、secret redaction、health/version helpers 需要保持与本地 `foundationx` 兼容面一致；修改其中任一侧时同步更新契约测试。

## 测试指南

使用 Go 标准 `testing` package。测试按行为命名，例如 `TestLoaderMergesSourcesByPrecedence`。重点覆盖：

- secret redaction 和 raw secret 不泄漏。
- 显式 source loading，不隐式发现 `.env`、`production.yaml` 或 `/home/k8s/secrets/env/*`。
- validation、error kind、context cancellation/not found 映射。
- source trace、contract schema、rendered template integration。

测试不得依赖 machine-specific paths、production secrets 或隐式配置发现。

## 提交与 Pull Request 指南

提交信息遵循 Lore Commit Protocol：第一行说明为什么改，正文用 trailers 记录约束、验证和风险。实质性变更至少包含 `Tested:` 和 `Not-tested:`。

Pull requests 应包含目的说明、关联 goal/issue、API 或行为变化、实际运行的 validation commands，以及 release evidence 状态。只有文档或生成报告变更需要 screenshots。

## 安全与配置提示

`configx` 必须保持 configuration loading 显式。不得自动读取 `.env`、`production.yaml` 或 `/home/k8s/secrets/env/*`。不得在 errors、tests、release manifests、examples 或日志中记录 raw secret values。禁止引入 `x.go`、交易领域术语或基础设施驱动依赖，除非目标文档和边界脚本同步更新并经过审查。
