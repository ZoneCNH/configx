# 当前项目深度评估报告（2026-06-04）

## 结论

综合评分：**80/100**。

这是一个基础面较好的 Go 公共库项目：核心边界清楚，显式配置加载、安全脱敏、契约测试、发布证据脚本都已经成型。主要扣分点不在“代码完全不可用”，而在结构性一致性：当前 checkout 的本地门禁不能完整闭环，目标文档与公共 API 存在漂移，核心实现集中度偏高，secret 语义仍偏启发式。

评分置信度：**中高**。本报告基于本地仓库静态检查、代码阅读和实际命令验证；未检查远端 GitHub branch protection、历史 CI 结果或真实下游用户反馈。

## 当前状态快照

- 分支状态：`main...origin/main [ahead 1]`。
- 跟踪文件：104 个。
- Go 文件：42 个。
- 测试/模糊测试入口：58 个。
- 被忽略的本地状态：`.omx/`、`.worktree/`、`docs/goal1.md`、`release/manifest/latest.json`。
- 本次工作只新增本报告；未修改实现代码。

## 维度评分

| 维度 | 分数 | 判断 |
| --- | ---: | --- |
| 核心 API 与行为 | 8/10 | 显式 source、source trace、last-wins merge、Decode 基本可用；但 merge strategy API 与目标文档不一致。 |
| 测试与契约 | 8.5/10 | 单元、race、contracts 均通过，覆盖 secret redaction、显式加载、source trace、foundationx 兼容面。 |
| 安全与配置边界 | 7/10 | 禁止隐式配置、边界脚本和 redaction 已有基础；secret 分类不完整，本地 secret scan 会扫到忽略目录。 |
| 可维护性 | 7/10 | 模块划分清楚，但 `pkg/configx/core.go` 承担过多职责，未来改动容易扩大影响面。 |
| 交付与发布可复现性 | 7/10 | release/check 脚本齐全，manifest 有校验；但当前本地 `make ci` 在 security gate 处失败，CI workflow 名称与实际 gate 覆盖不完全一致。 |

## 验证证据

| 命令 | 结果 | 说明 |
| --- | --- | --- |
| `GOWORK=off go test ./...` | 通过 | 所有包测试通过。 |
| `GOWORK=off go test -race ./...` | 通过 | race 检测通过。 |
| `GOWORK=off make contracts` | 通过 | contract artifacts 与 `go test ./contracts` 通过。 |
| `GOWORK=off make boundary` | 通过 | 业务词、隐式配置、依赖边界通过；过程中 Go 试图写 `/home/zone/go/pkg/mod` stat cache，受当前沙箱限制报 warning，但脚本最终通过。 |
| `GOWORK=off GOCACHE=/tmp/configx-go-cache golangci-lint run ./...` | 通过 | `0 issues`。不设可写 `GOCACHE` 时，本环境会出现 lint/package loading 失败。 |
| `GOWORK=off GOCACHE=/tmp/configx-go-cache GOLANGCI_LINT_CACHE=/tmp/configx-golangci-cache make ci` | 失败 | fmt、vet、lint、test、race、boundary 已通过；security 因沙箱禁止访问 `vuln.go.dev` 停止。 |
| `GOWORK=off GOCACHE=/tmp/configx-go-cache GOLANGCI_LINT_CACHE=/tmp/configx-golangci-cache make security`（授权网络后） | 失败 | `govulncheck` 报告当前代码路径 0 个漏洞；随后 `check_secrets.sh` 扫到忽略目录 `.worktree/v2.md` 中的示例 `password=`，导致 security gate 失败。 |

## 主要优点

1. **显式配置加载边界清楚。** README 与实现都坚持不隐式读取 `.env`、`production.yaml` 或 `/home/k8s/secrets/env/*`；`EnvSource`、`EnvFileSource`、`JSONFileSource`、`TOMLFileSource`、`YAMLFileSource` 和 `MapSource` 都是显式构造。
2. **source trace 与 redaction 已成为核心行为。** `LoadResult`、`SourceReport`、`Sanitize` 和 secret source 标记让调试和发布证据有基础。
3. **测试资产相对扎实。** 58 个测试/模糊测试入口覆盖公共 API、contracts、foundationx compatibility、examples、release manifest tool、sanitize 和 validation。
4. **发布治理意识强。** `Makefile`、`scripts/check_*`、release manifest 和 integration rendering 都已存在，不是纯代码仓库。
5. **foundationx 兼容面被隔离。** `internal/foundationx` 与 `contracts/` 让本仓库可以在不引入外部基础设施依赖的前提下保持兼容面。

## 结构性问题

### P1：当前本地门禁不能完整闭环

证据：

- `make ci` 在可写缓存配置下能通过 fmt、vet、lint、test、race、boundary，但 security gate 最终失败。
- 授权网络后，`govulncheck` 明确报告当前代码路径 0 个漏洞；失败点变为 `scripts/check_secrets.sh`。
- `scripts/check_secrets.sh:21-27` 使用 `grep -R` 扫描 `.`，排除了 `.git`、`.omx`、`vendor`、`*.sum`、`check_secrets.sh`、`goal.md`，但没有排除 `.worktree`。
- 当前 checkout 存在被忽略目录 `.worktree/`，其中 `.worktree/v2.md` 触发了 `password[[:space:]]*=` 模式。

影响：

- 当前项目不能在本地声称完整 `make ci` 通过。
- 这不是业务代码漏洞，但会破坏“发布前一键验证”的可信度。
- clean CI checkout 可能不复现 `.worktree` 问题；但本地 release gate 与开发者体验仍不稳定。

建议：

- secret scan 改为扫描 tracked files，例如基于 `git grep -n -E`，或显式排除 `.worktree`、`release/manifest/latest.json` 等已声明生成/运行时路径。
- 在文档或 Makefile 中固定 `GOCACHE=/tmp/...`、`GOLANGCI_LINT_CACHE=/tmp/...` 的可写缓存策略，避免受只读 home cache 影响。

### P1：Merge API/目标文档与实现存在漂移

证据：

- `docs/goal.md:678-687` 目标 API 定义 `MergeLastWins`、`MergeFirstWins`、`MergeErrorOnConflict` 和 `func Merge(strategy MergeStrategy, maps ...Map) (Map, error)`。
- 当前实现只有 `pkg/configx/core.go:115-117` 的 `type MergeStrategy int` 与 `const LastWins MergeStrategy = iota`。
- `pkg/configx/core.go:126-134` 保存了 `mergeStrategy` 和 `WithMergeStrategy`，但 `Load` 路径中没有根据 strategy 分支处理；实际行为始终是后 source 覆盖前 source。

影响：

- 公共 API 已暴露 `WithMergeStrategy`，但行为空间只有 last-wins。
- 目标文档、测试计划与实现不同步，后续贡献者容易围绕错误目标开发。
- 如果用户依赖 `WithMergeStrategy`，目前除 LastWins 外没有可验证语义。

建议：

- 做一次明确产品决策：要么实现 `FirstWins` / `ErrorOnConflict` / 独立 `Merge` 并补契约测试；要么从公共面和目标文档中移除未兑现的 strategy。

### P1：Secret 语义仍偏启发式

证据：

- `pkg/configx/core.go:614-617` 的 `IsSecretKey` 只识别 `secret`、`password`、`passwd`、`token`、`access_key`、`secret_key`。
- `pkg/configx/core.go:541-543` 把 tag option `secret` 识别为合法选项，但 `configTag` 结构没有 secret 字段，`applyTagOptions` 也没有处理 `secret`。
- `pkg/configx/core.go:556-612` 的 Decode 支持 `SecretString` 类型，但对 `config:"...,secret"` 这种 tag 没有独立语义。

影响：

- `api_key`、`private_key` 等常见 secret key 不会被 `IsSecretKey` 标记。
- 使用者看到 `secret` tag 被接受，可能误以为普通 string 字段也会被按 secret 处理。
- 错误清洗和 source trace 仍依赖 key 命名，不能覆盖所有泄漏形态。

建议：

- 扩展 secret taxonomy：至少加入 `api_key`、`private_key`、`client_secret` 等常见形式。
- 让 `config:"...,secret"` 写入 tag 语义，并在 Decode/source report/sanitize 路径有测试保护。
- 优先用 `SecretString` 作为强类型 secret；文档中明确 string+secret tag 的真实语义。

### P2：公共实现集中在单个核心文件

证据：

- `pkg/configx/core.go` 638 行，包含公共类型、loader、source 实现、merge、flatten、Decode、tag parsing、类型转换、secret 检测和错误清洗。
- `pkg/configx/structured_sources.go` 已经开始拆出 TOML/YAML source，说明项目接受按 source/职责拆分的形态。

影响：

- 每次修改 Decode、source、merge、redaction 都会触碰同一大文件，评审和回归定位成本增加。
- 当前结构对 v0.1 还可控，但继续扩展 merge strategy、secret tag、list decode 后会显著变重。

建议：

- 在行为测试锁定后拆分为 `loader.go`、`source_env.go`、`source_file.go`、`decode.go`、`merge.go`、`secret.go`、`sanitize.go`。
- 拆分应保持纯搬迁优先，不顺手改行为。

### P2：CI workflow 的命名与 gate 覆盖不完全一致

证据：

- `docs/testing.md:44-53` 将 fmt、vet、lint、test、race、boundary、security、contracts、integration、evidence 都列为必需 gate。
- `.github/workflows/ci.yml:8-16` 的 `CI` workflow 只有 contracts 和 secret scan。
- release workflow 会跑 `make release-check`，integration/security 也有独立 workflow；但如果 branch protection 只要求 `CI`，PR 不能覆盖完整质量门。

影响：

- “CI 通过”这一说法可能有歧义。
- PR 合并前与 release 前验证强度可能不一致。

建议：

- 要么让 `CI` workflow 覆盖 `make ci` 的主体 gate，要么在文档和 branch protection 中明确要求所有独立 workflow。
- 在 README/release 文档中避免把轻量 docs-contracts workflow 称为完整 CI。

### P2：release manifest 证据面偏窄

证据：

- `internal/tools/releasemanifest/main.go:48-54` 的 contract digest 只包含 config/error/health/version schema 和 metrics.md。
- `contracts/manifest.schema.json` 与 `release/manifest/template.json` 由 `scripts/check_contracts.sh` 要求存在，但不在 manifest contract digest 列表里。
- `internal/tools/releasemanifest/main.go:164-166` 的 artifact 列表只有 `release/manifest/latest.json`。
- `scripts/check_release_preflight.sh:35-41` 要求本地 HEAD 与 `origin/main` 对齐；当前分支 ahead 1，因此当前 checkout 本身不是 release-ready 状态。

影响：

- 发布证据能证明一部分契约文件和依赖状态，但对 manifest schema/template 自身的覆盖不足。
- 当前 checkout 需要先处理本地 ahead 状态和生成产物策略，才能满足 release preflight。

建议：

- 将 `contracts/manifest.schema.json`、`release/manifest/template.json` 纳入 manifest digest。
- 如有 release sidecar/evidence 目录，明确哪些是 artifact、哪些必须不提交。

### P2：Decode/flatten 能力与目标能力边界需要整理

证据：

- `pkg/configx/core.go:397-427` 的 `flattenMap` 对非 map 值统一 `fmt.Sprint`，数组/list 会变成 Go 格式字符串。
- `pkg/configx/core.go:556-612` 的 `setField` 支持 string、bool、数字、float、duration、`SecretString`、`encoding.TextUnmarshaler`，没有 `[]string` 或 slice 支持。
- `docs/goal.md` 的目标中出现过 `[]string` 逗号分割能力；但 v0.1 也明确不做 nested structs/map/custom parser。

影响：

- TOML/YAML/JSON 的结构化优势在 Decode 层会被压成字符串。
- 使用者可能以为结构化 source 支持 list，但实际不支持。

建议：

- 明确 v0.1 能力边界：不支持 slice/list，还是补 `[]string`。
- 若补 slice，优先只支持 `[]string` 逗号分割和结构化 source 的 string slice，避免扩大为完整 decoder。

### P3：错误 taxonomy 存在轻微语义漂移

证据：

- `pkg/configx/errors.go:19` 定义了 `ErrorKindCanceled`。
- `pkg/configx/errors.go:90-98` 的 `contextError` 只把 `context.DeadlineExceeded` 映射为 timeout；其他 context 错误都映射为 unavailable。
- `docs/design.md` 曾说明 context cancellation 归类为 unavailable，因此这可能是有意设计，但 `canceled` kind 的存在会造成 API 语义疑问。

影响：

- API 使用者看到 `ErrorKindCanceled` 可能期待 context cancellation 被映射到 canceled。
- 合同中保留未使用错误种类会增加兼容负担。

建议：

- 明确二选一：保持设计并文档化 `ErrorKindCanceled` 的适用场景；或将 `context.Canceled` 映射为 `ErrorKindCanceled` 并补测试。

## 优先级修复路线

1. **先修 gate 可复现性。** 让 `make ci` 在当前 checkout 可明确通过或明确只受网络权限影响失败；secret scan 改为 tracked-file scan 是最小收益最高的改动。
2. **关闭 Merge API 漂移。** 决定实现多策略还是收窄 API/文档，并用测试锁住。
3. **补强 secret 语义。** 扩展 key taxonomy，处理或删除 `secret` tag 误导点。
4. **拆分 `core.go`。** 先纯搬迁，后行为增强；每一步保持测试通过。
5. **对齐 CI/release 叙事。** 明确 PR 必需 checks、release 必需 checks、manifest 包含的证据范围。
6. **整理目标文档。** 把已实现、未实现、刻意不做的能力分开，减少后续误读。

## 预期提分空间

- 修复本地 `make ci` 完整闭环：+4 到 +6 分。
- 实现或收窄 MergeStrategy：+3 到 +5 分。
- secret tag/taxonomy 与 scan 策略补齐：+3 到 +5 分。
- 拆分核心文件并保持行为不变：+2 到 +4 分。
- CI/release evidence 对齐：+2 到 +4 分。

如果上述前四项完成，并且 `make ci`、`make integration`、`make release-check` 在 clean checkout 中可复现通过，项目可稳定进入 **88-92/100** 区间。
