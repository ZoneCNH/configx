# configx 完整可执行目标提示 v1.0

> 文件名：`configx_goal_executable_prompt_v1_0.md`  
> 目标模块：`github.com/ZoneCNH/configx`
> 模块定位：配置加载 / 合并 / 校验 / 脱敏 / Secret 处理的独立公共基础库  
> 分层定位：L1 运行时基础能力层  
> 上游依赖：`github.com/ZoneCNH/foundationx`
> 适用项目：x.go、postgresx、kafkax、redisx、taosx、ossx、observex  
> 执行方法：Goal Runtime Prompt v3.1 + Harness + AutoResearch + Self-improving + Evidence Protocol  
> 生成日期：2026-06-01  

---

# 0. 使用方式

将本文完整交给 Agent Teams / Codex / Claude Code / Cursor Agent / GitHub Copilot Workspace 执行。

执行前必须确认：

```text
1. 当前目标是创建或完善独立 Go module：github.com/ZoneCNH/configx
2. configx 是 L1 配置基础库，不是 x.go 业务配置模块
3. configx 可以依赖 foundationx
4. configx 不允许依赖 x.go
5. configx 不允许依赖 PostgreSQL / Kafka / Redis / TDengine / OSS driver
6. configx 不允许隐式读取 /home/k8s/secrets/env/*
7. configx 可以提供显式 LoadEnvFile(path)，但路径必须由调用方传入
8. configx 不允许自动查找 .env、production.yaml、config.local.yaml
9. configx 不允许持有全局配置、单例配置或默认生产配置
10. configx 不允许在日志、错误、测试输出、release manifest 中泄露 secret
11. 所有完成声明必须使用 “DONE with evidence:”
```

---

# 1. 主目标

```text
GOAL-20260601-CONFIGX-001

建立 configx 独立公共配置基础库，为 x.go 与基础库体系提供统一、显式、可测试、可脱敏、可校验、可合并、可发布的配置加载能力。

configx 必须支持从环境变量、显式 env 文件、JSON 文件、内存 Map 等来源加载配置；支持多来源优先级合并；支持结构体解码；支持 required/default/secret 标签；支持 SecretString；支持配置校验、脱敏输出、配置来源追踪、错误归一化、测试工具、Examples、CI/Harness/Evidence/Release 流程。

configx 必须不依赖 x.go，不理解业务配置语义，不隐式读取生产密钥路径，不持有全局配置，不泄露敏感信息。
```

---

# 2. 问题底层本质

configx 不是“读一个 .env 文件”。

configx 的底层本质是：

```text
把配置从散落在业务启动逻辑中的隐式副作用，变成显式、可验证、可脱敏、可追踪、可合并、可复用的运行时契约。
```

它解决的是：

```text
1. x.go 和所有基础库配置来源不统一
2. 各模块自己读取 env，导致路径、默认值、脱敏、错误处理不一致
3. 密钥容易被日志、测试、release manifest 泄露
4. 配置优先级不透明，排障困难
5. 业务模块直接读取 /home/k8s/secrets/env/*，难以测试
6. Agent Teams 执行时不知道配置加载完成的 Evidence
7. 基础库被迫理解业务配置结构，破坏边界
```

configx 的核心价值：

```text
配置来源显式化 + 优先级标准化 + Secret 安全化 + 验证自动化 + Evidence 可证明。
```

---

# 3. 不可再拆解的基本真理

## 3.1 configx 是 L1，不是业务配置中心

configx 可以知道：

```text
Source / Loader / Provider / Env / EnvFile / JSON / Map / Merge / Decode / Validate / Secret / Sanitize
```

configx 不应该知道：

```text
BTCUSDT / Kline / MacroRegime / M1-M7 / S1-S7 / TradingSignal / Kafka topic / Redis key / TDengine table
```

## 3.2 configx 必须显式加载

禁止自动读取：

```text
/home/k8s/secrets/env/postgres.env
/home/k8s/secrets/env/redis.env
/home/k8s/secrets/env/kafka.env
/home/k8s/secrets/env/taos.env
/home/k8s/secrets/env/oss.env
.env
config.local.yaml
production.yaml
```

允许：

```go
result, err := configx.LoadEnvFile(ctx, "/home/k8s/secrets/env/postgres.env")
```

路径必须由调用方传入。

## 3.3 configx 不持有全局状态

禁止：

```go
var GlobalConfig Config
func Init()
func Get()
func MustGet()
func SetDefaultPath(...)
```

允许：

```go
loader := configx.NewLoader()
result, err := loader.AddSource(configx.NewEnvFileSource(path)).Load(ctx)
```

## 3.4 Secret 默认不可见

所有 secret 在以下位置必须脱敏：

```text
fmt.Stringer
log fields
error message
test output
release manifest
evidence
sanitized config
```

---

# 4. 被误认为真理的常见假设

| 常见假设 | 为什么错 | 正确口径 |
|---|---|---|
| configx 就是 viper 包一层 | 容易引入全局状态和隐式路径 | 定义自己的显式契约 |
| 配置库可以自动读 .env | 隐式副作用，不可治理 | 调用方显式传 path |
| 配置错误可以 panic | 服务启动阶段需要清晰错误 | 返回 foundationx.Error |
| Secret 是 string | 容易泄露 | 使用 foundationx.SecretString |
| 配置优先级靠约定 | 排障困难 | MergePlan 显式记录来源和覆盖关系 |
| 所有格式都必须 v0.1 支持 | 过早扩大依赖 | v0.1 先 env/envfile/json/map |
| configx 可以定义 x.go 总配置 | 业务污染 | x.go 定义业务 Config，configx 只负责解码 |
| 热加载是配置库必需能力 | 会引入复杂运行时状态 | v0.1 不做 watch，后续单独设计 |


# 5. 范围

## 5.1 范围内

```text
Source 抽象
命名 source metadata
Env source
Env file source
JSON file source
Map source
Multi-source loader
显式 source order
按 precedence 合并
Source trace
LoadResult
Decode into struct
config/default/required/secret tags
基础 type conversion
Validator interface
SecretString 集成
Sanitized output
到 foundationx.Error 的 error mapping
TestKit
Examples
Harness scripts
Release manifest
```

## 5.2 v0.1 可选项

```text
YAML source
TOML source
```

条件：

```text
必须 AutoResearch + ADR 选择依赖
必须保持显式加载
不得引入全局配置框架
```

## 5.3 范围外

```text
业务配置结构
x.go AppConfig
PostgreSQL/Kafka/Redis/TDengine driver config 实例
自动读取生产密钥路径
全局配置单例
配置中心服务
远程配置拉取
etcd/consul/nacos
watch 热加载
动态 runtime reconfigure
secret manager / Vault client
```

---

# 6. 目标仓库与模块

```text
github.com/ZoneCNH/configx
```

go.mod：

```go
module github.com/ZoneCNH/configx

go 1.23
```

必须依赖：

```text
github.com/ZoneCNH/foundationx
```

v0.1 优先标准库：

```text
context
os
bufio
strings
strconv
encoding/json
reflect
time
path/filepath
```

可选依赖必须通过 ADR：

```text
YAML parser
TOML parser
mapstructure-like decoder
fsnotify watcher
```

默认裁决：

```text
v0.1 不引入 watch，不引入全局配置框架。
YAML/TOML 若进入 v0.1，必须作为 adapter 并由 ADR 记录依赖选择。
```

---

# 7. 标准目录结构

```text
configx/
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── LICENSE
├── Makefile
├── .gitignore
├── .golangci.yml
├── pkg/
│   └── configx/
│       ├── doc.go
│       ├── source.go
│       ├── loader.go
│       ├── result.go
│       ├── env.go
│       ├── envfile.go
│       ├── json.go
│       ├── map.go
│       ├── merge.go
│       ├── decode.go
│       ├── validate.go
│       ├── secret.go
│       ├── sanitize.go
│       ├── errors.go
│       ├── options.go
│       ├── version.go
│       └── *_test.go
├── internal/
│   ├── parser/
│   │   └── envfile.go
│   ├── reflectx/
│   │   └── decode.go
│   └── testutil/
├── testkit/
│   ├── envfile.go
│   ├── fixture.go
│   └── assert.go
├── examples/
│   ├── envfile/
│   ├── merge/
│   ├── decode/
│   ├── secret/
│   └── xgo_secrets_path/
├── contracts/
│   ├── source.schema.json
│   ├── result.schema.json
│   ├── error.schema.json
│   ├── public_api.md
│   └── tags.md
├── docs/
│   ├── spec.md
│   ├── design.md
│   ├── api.md
│   ├── sources.md
│   ├── envfile.md
│   ├── merge.md
│   ├── decode.md
│   ├── validation.md
│   ├── secrets.md
│   ├── sanitize.md
│   ├── xgo-integration.md
│   ├── testing.md
│   ├── release.md
│   └── adr/
│       ├── ADR-20260601-001-explicit-source-loading.md
│       ├── ADR-20260601-002-no-global-config.md
│       ├── ADR-20260601-003-secret-handling.md
│       └── ADR-20260601-004-yaml-toml-scope.md
├── scripts/
│   ├── check_boundary.sh
│   ├── check_secrets.sh
│   ├── check_contracts.sh
│   └── generate_manifest.sh
├── release/
│   └── manifest/
│       └── v0.1.0.json
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── security.yml
│       └── release.yml
└── .agent/
    ├── goal.md
    ├── spec.md
    ├── design.md
    ├── plan.md
    ├── tasks.md
    ├── harness.md
    ├── gates.md
    ├── evidence.md
    ├── review.md
    ├── release.md
    └── retrospective.md
```


# 8. 公共 API 设计

## 8.1 LoadResult

文件：

```text
pkg/configx/result.go
```

目标 API：

```go
package configx

import "time"

type Value struct {
	Key        string
	Value      string
	Secret     bool
	Source     string
	LoadedAt   time.Time
	Overridden bool
}

type Map map[string]Value

type LoadResult struct {
	Values   Map
	Sources  []SourceReport
	LoadedAt time.Time
}

type SourceReport struct {
	Name      string
	Kind      string
	Path      string
	Loaded    bool
	Error     string
	LoadedAt  time.Time
	ValueKeys []string
}

func (r LoadResult) Get(key string) (string, bool)
func (r LoadResult) Sanitize() SanitizedResult
```

推荐裁决：

```text
v0.1 不实现 MustGet，避免 panic 风格 API。
```

## 8.2 Source 接口

文件：

```text
pkg/configx/source.go
```

目标 API：

```go
package configx

import "context"

type Source interface {
	Name() string
	Kind() string
	Load(ctx context.Context) (Map, error)
}

type SourceOption func(*sourceOptions)

type sourceOptions struct {
	name string
}
```

要求：

```text
1. Source 不持有全局状态
2. Source.Load 必须尊重 context
3. Source.Name 用于 trace
4. Source.Kind 用于调试和 evidence
```

## 8.3 Loader

文件：

```text
pkg/configx/loader.go
```

目标 API：

```go
type Loader struct {
	sources []Source
	options options
}

func NewLoader(opts ...Option) *Loader
func (l *Loader) AddSource(source Source) *Loader
func (l *Loader) Load(ctx context.Context) (LoadResult, error)

type Option func(*options)

type options struct {
	mergeStrategy MergeStrategy
	failFast      bool
}

func WithMergeStrategy(strategy MergeStrategy) Option
func WithFailFast(failFast bool) Option
```

默认行为：

```text
1. sources 按 AddSource 顺序加载
2. 默认后加载 source 覆盖先加载 source
3. 覆盖必须记录 Overridden
4. failFast=true 时任意 source 失败立即返回
5. 默认 LastWins
```

## 8.4 环境变量 Source

文件：

```text
pkg/configx/env.go
```

目标 API：

```go
type EnvSource struct {
	name   string
	prefix string
	keys   []string
}

func NewEnvSource(prefix string, keys []string, opts ...SourceOption) *EnvSource
```

要求：

```text
1. prefix 可为空
2. keys 为空时默认不读取整个 os.Environ
3. 默认只读取指定 keys，避免把系统敏感环境全量吸入
4. key 命名保持原始形式
```

可选 API：

```go
func NewAllEnvSource(prefix string, opts ...SourceOption) *EnvSource
```

但必须文档警告：

```text
AllEnvSource 可能读入不相关敏感环境变量，生产慎用。
```

## 8.5 env 文件 Source

文件：

```text
pkg/configx/envfile.go
```

目标 API：

```go
type EnvFileSource struct {
	name string
	path string
}

func NewEnvFileSource(path string, opts ...SourceOption) *EnvFileSource
func LoadEnvFile(ctx context.Context, path string) (LoadResult, error)
```

支持语法：

```text
KEY=value
KEY="quoted value"
KEY='quoted value'
KEY=value with spaces
# comment
export KEY=value
EMPTY=
```

要求：

```text
1. path 必须由调用方显式传入
2. 不自动查找 .env
3. 不自动读取 /home/k8s/secrets/env/*
4. 文件不存在返回 ErrorKindNotFound
5. 解析错误返回 ErrorKindValidation
6. 不在错误中输出 secret value
```

## 8.6 JSON 文件 Source

文件：

```text
pkg/configx/json.go
```

目标 API：

```go
type JSONFileSource struct {
	name string
	path string
}

func NewJSONFileSource(path string, opts ...SourceOption) *JSONFileSource
func LoadJSONFile(ctx context.Context, path string) (LoadResult, error)
```

嵌套 JSON 需要 flatten：

```json
{
  "postgres": {
    "host": "127.0.0.1",
    "port": 5432
  }
}
```

转换为：

```text
postgres.host=127.0.0.1
postgres.port=5432
```

## 8.7 Map Source

文件：

```text
pkg/configx/map.go
```

目标 API：

```go
type MapSource struct {
	name    string
	values  map[string]string
	secrets map[string]bool
}

func NewMapSource(name string, values map[string]string) *MapSource
func NewSecretMapSource(name string, values map[string]string, secretKeys []string) *MapSource
```

用途：

```text
测试
examples
x.go 启动层适配
```


# 9. Merge、Decode、Validate 与 Secret

## 9.1 Merge

文件：

```text
pkg/configx/merge.go
```

目标 API：

```go
type MergeStrategy string

const (
	MergeLastWins        MergeStrategy = "last_wins"
	MergeFirstWins       MergeStrategy = "first_wins"
	MergeErrorOnConflict MergeStrategy = "error_on_conflict"
)

func Merge(strategy MergeStrategy, maps ...Map) (Map, error)
```

要求：

```text
1. 默认 LastWins
2. 发生覆盖时记录 Overridden
3. ErrorOnConflict 遇到不同值返回 conflict
4. secret 标记需要合并保留
5. source trace 不能丢失
```

## 9.2 Decode

文件：

```text
pkg/configx/decode.go
```

目标 API：

```go
func Decode(result LoadResult, target any) error
func DecodeMap(values Map, target any) error
```

推荐 struct tag：

```go
type PostgresConfig struct {
	Host     string                   `config:"POSTGRES_HOST,required"`
	Port     int                      `config:"POSTGRES_PORT,default=5432"`
	User     string                   `config:"POSTGRES_USER,required"`
	Password foundationx.SecretString `config:"POSTGRES_PASSWORD,required,secret"`
	SSLMode  string                   `config:"POSTGRES_SSLMODE,default=disable"`
}
```

支持字段类型：

```text
string
int / int32 / int64
bool
float64
time.Duration
foundationx.SecretString
[]string 按逗号分割
```

v0.1 不支持：

```text
复杂嵌套结构自动递归
map[string]T
自定义复杂 parser
```

要求：

```text
1. target 必须是结构体指针
2. required 缺失返回 ErrorKindConfig
3. default 在缺失时使用
4. secret 字段 decode 为 SecretString
5. 类型转换错误返回 ErrorKindValidation
6. 错误信息不包含 secret value
```

## 9.3 Validate

文件：

```text
pkg/configx/validate.go
```

目标 API：

```go
type Validator interface {
	Validate() error
}

func Validate(target any) error
```

行为：

```text
1. 如果 target 实现 Validator，则调用 Validate
2. 如果字段 tag 包含 required，则 Decode 阶段已经处理
3. Validate 不做业务领域规则推断
```

## 9.4 Secret 与 Sanitize

文件：

```text
pkg/configx/secret.go
pkg/configx/sanitize.go
```

目标 API：

```go
type Sanitizer interface {
	Sanitize() any
}

type SanitizedResult struct {
	Values  map[string]SanitizedValue
	Sources []SourceReport
}

type SanitizedValue struct {
	Key    string
	Value  string
	Secret bool
	Source string
}
```

要求：

```text
1. Secret value 输出 "***"
2. 非 secret value 输出原值
3. key 名包含 PASSWORD/TOKEN/SECRET/API_KEY/ACCESS_KEY/SECRET_KEY/PRIVATE_KEY 时自动标记 secret
4. 不匹配单独 KEY，避免误伤
5. Reveal 只通过 foundationx.SecretString.Reveal
```

## 9.5 Error Mapping

文件：

```text
pkg/configx/errors.go
```

目标 API：

```go
func MapError(op string, err error) error
```

映射原则：

```text
os.ErrNotExist -> ErrorKindNotFound
permission denied -> ErrorKindAuth 或 ErrorKindUnavailable，需 ADR 裁决
parse error -> ErrorKindValidation
missing required -> ErrorKindConfig
type conversion error -> ErrorKindValidation
conflict -> ErrorKindConflict
context canceled -> ErrorKindCanceled
context deadline exceeded -> ErrorKindTimeout
```

要求：

```text
1. 返回 foundationx.Error
2. 保留 Cause
3. 不泄露 secret value
```


# 10. 规格

```text
SPEC-configx-v1.0
```

## REQ-CONFIGX-001：独立 Go module

验收标准：

```text
AC-REQ-CONFIGX-001-001: go.mod module 为 github.com/ZoneCNH/configx
AC-REQ-CONFIGX-001-002: go test ./... 通过
AC-REQ-CONFIGX-001-003: go list -deps ./... 不包含 github.com/bytechainx/x.go
AC-REQ-CONFIGX-001-004: README 明确模块定位和非目标
```

## REQ-CONFIGX-002：依赖边界

验收标准：

```text
AC-REQ-CONFIGX-002-001: 允许依赖 foundationx
AC-REQ-CONFIGX-002-002: 不允许依赖 PostgreSQL/Kafka/Redis/TDengine/OSS driver
AC-REQ-CONFIGX-002-003: 不允许依赖 x.go
AC-REQ-CONFIGX-002-004: 不允许出现 Market/Macro/Regime/Trading 业务模型
AC-REQ-CONFIGX-002-005: 不允许全局配置单例
```

## REQ-CONFIGX-003：Source 契约

验收标准：

```text
AC-REQ-CONFIGX-003-001: 定义 Source interface
AC-REQ-CONFIGX-003-002: Source 包含 Name/Kind/Load
AC-REQ-CONFIGX-003-003: Load 支持 context
AC-REQ-CONFIGX-003-004: SourceReport 记录来源信息
```

## REQ-CONFIGX-004：环境变量 Source

验收标准：

```text
AC-REQ-CONFIGX-004-001: 支持按 prefix 和 keys 读取 env
AC-REQ-CONFIGX-004-002: 默认不读取所有 os.Environ
AC-REQ-CONFIGX-004-003: 缺失 key 不 panic
AC-REQ-CONFIGX-004-004: secret key 自动标记
```

## REQ-CONFIGX-005：env 文件 Source

验收标准：

```text
AC-REQ-CONFIGX-005-001: 支持显式 path 加载
AC-REQ-CONFIGX-005-002: 支持 KEY=value
AC-REQ-CONFIGX-005-003: 支持 export KEY=value
AC-REQ-CONFIGX-005-004: 支持注释和空行
AC-REQ-CONFIGX-005-005: 支持 quoted value
AC-REQ-CONFIGX-005-006: 文件不存在返回 not_found
AC-REQ-CONFIGX-005-007: 解析错误返回 validation
AC-REQ-CONFIGX-005-008: 不自动查找 .env
```

## REQ-CONFIGX-006：JSON Source

验收标准：

```text
AC-REQ-CONFIGX-006-001: 支持显式 JSON 文件路径
AC-REQ-CONFIGX-006-002: 支持 nested object flatten
AC-REQ-CONFIGX-006-003: 支持 string/number/bool
AC-REQ-CONFIGX-006-004: 非支持类型行为文档化
AC-REQ-CONFIGX-006-005: 解析错误返回 validation
```

## REQ-CONFIGX-007：Map Source

验收标准：

```text
AC-REQ-CONFIGX-007-001: 支持 map[string]string
AC-REQ-CONFIGX-007-002: 支持 secretKeys
AC-REQ-CONFIGX-007-003: Source trace 正确
```

## REQ-CONFIGX-008：Merge

验收标准：

```text
AC-REQ-CONFIGX-008-001: 支持 LastWins
AC-REQ-CONFIGX-008-002: 支持 FirstWins
AC-REQ-CONFIGX-008-003: 支持 ErrorOnConflict
AC-REQ-CONFIGX-008-004: 覆盖关系可追踪
AC-REQ-CONFIGX-008-005: secret 标记不丢失
```

## REQ-CONFIGX-009：Decode

验收标准：

```text
AC-REQ-CONFIGX-009-001: Decode 要求 target 为结构体指针
AC-REQ-CONFIGX-009-002: 支持 string/int/bool/float64/time.Duration
AC-REQ-CONFIGX-009-003: 支持 foundationx.SecretString
AC-REQ-CONFIGX-009-004: 支持 required
AC-REQ-CONFIGX-009-005: 支持 default
AC-REQ-CONFIGX-009-006: 支持 secret
AC-REQ-CONFIGX-009-007: 类型转换错误返回 validation
AC-REQ-CONFIGX-009-008: 错误不泄露 secret value
```

## REQ-CONFIGX-010：Validate

验收标准：

```text
AC-REQ-CONFIGX-010-001: 支持 Validator interface
AC-REQ-CONFIGX-010-002: Validate 调用 target.Validate
AC-REQ-CONFIGX-010-003: Validate 错误归一或保留 cause
```

## REQ-CONFIGX-011：Secret 与 Sanitize

验收标准：

```text
AC-REQ-CONFIGX-011-001: SecretString 默认输出 ***
AC-REQ-CONFIGX-011-002: Sanitize 不输出 secret 原值
AC-REQ-CONFIGX-011-003: 自动识别 PASSWORD/TOKEN/SECRET/API_KEY 等 key
AC-REQ-CONFIGX-011-004: release manifest 不包含 secret 原值
AC-REQ-CONFIGX-011-005: tests 覆盖 fmt.Sprint 不泄露
```

## REQ-CONFIGX-012：Harness

验收标准：

```text
AC-REQ-CONFIGX-012-001: make ci 通过
AC-REQ-CONFIGX-012-002: boundary gate 通过
AC-REQ-CONFIGX-012-003: secret gate 通过
AC-REQ-CONFIGX-012-004: contract gate 通过
AC-REQ-CONFIGX-012-005: examples gate 通过
AC-REQ-CONFIGX-012-006: release manifest 生成
```


# 11. 计划

```text
PLAN-GOAL-20260601-CONFIGX-001-v1.0
```

## 阶段 0：上下文恢复

目标：

```text
确认 configx 在基础库体系中的位置、密钥路径约束和 x.go 集成方式。
```

输出：

```text
.agent/context.md
```

必须记录：

```text
configx 是 L1
foundationx 是 L0
x.go 的密钥路径是 /home/k8s/secrets/env/*
configx 只提供显式 LoadEnvFile(path)，不自动读取
```

## 阶段 1：骨架

创建独立仓库骨架：

```text
go.mod
README.md
CHANGELOG.md
Makefile
pkg/configx/*
docs/*
scripts/*
.agent/*
```

## 阶段 2：核心 Source

实现：

```text
Source
EnvSource
EnvFileSource
JSONFileSource
MapSource
```

## 阶段 3：合并与结果

实现：

```text
LoadResult
SourceReport
MergeStrategy
source trace
```

## 阶段 4：解码与校验

实现：

```text
struct decode
required/default/secret tag
Validator
```

## 阶段 5：密钥、脱敏与错误

实现：

```text
secret 标记
SanitizedResult
错误归一
```

## 阶段 6：示例与 TestKit

实现：

```text
examples
testkit fixtures
```

## 阶段 7：Harness 与 CI

实现：

```text
boundary
secret
contract
examples
evidence gates
```

## 阶段 8：文档与 ADR

补齐：

```text
README
docs
ADR
contracts
```

## 阶段 9：发布

生成：

```text
v0.1.0 发布证据
```

## 阶段 10：复盘

输出：

```text
自改进补丁
```


# 12. 任务拆解

## TASK-CONFIGX-001：创建模块骨架

```bash
mkdir -p configx
cd configx
go mod init github.com/ZoneCNH/configx
mkdir -p pkg/configx internal/parser internal/reflectx internal/testutil testkit examples/envfile examples/merge examples/decode examples/secret examples/xgo_secrets_path contracts docs/adr scripts release/manifest .agent .github/workflows
touch README.md CHANGELOG.md Makefile .gitignore .golangci.yml
```

验收：

```text
go.mod 存在
pkg/configx 存在
testkit 存在
docs/adr 存在
scripts 存在
.agent 存在
```

证据：

```text
EVID-TASK-CONFIGX-001-20260601-001: tree output
EVID-TASK-CONFIGX-001-20260601-002: go env GOMOD
```

## TASK-CONFIGX-002：接入 foundationx

```bash
go get github.com/ZoneCNH/foundationx
```

要求：

```text
不接入 x.go
不接入 driver
不接入全局配置框架
```

证据：

```text
EVID-TASK-CONFIGX-002-20260601-001: go.mod diff
```

## TASK-CONFIGX-003：实现 Source 与 LoadResult

文件：

```text
pkg/configx/source.go
pkg/configx/result.go
pkg/configx/source_test.go
pkg/configx/loader_test.go
```

测试：

```text
TestLoadResultGet
TestLoadResultSanitizeMasksSecret
TestSourceInterfaceCompile
```

## TASK-CONFIGX-004：实现 Loader

文件：

```text
pkg/configx/loader.go
pkg/configx/options.go
pkg/configx/loader_test.go
```

测试：

```text
TestLoaderLoadSingleSource
TestLoaderLoadMultipleSourcesLastWins
TestLoaderFailFast
TestLoaderSourceReports
```

## TASK-CONFIGX-005：实现 EnvSource

文件：

```text
pkg/configx/env.go
pkg/configx/env_test.go
```

测试：

```text
TestEnvSourceLoadsSpecifiedKeys
TestEnvSourcePrefix
TestEnvSourceMissingKey
TestEnvSourceDoesNotLoadAllByDefault
TestEnvSourceAutoMarksSecretKeys
```

## TASK-CONFIGX-006：实现 EnvFileSource

文件：

```text
pkg/configx/envfile.go
internal/parser/envfile.go
pkg/configx/envfile_test.go
```

测试：

```text
TestEnvFileKeyValue
TestEnvFileExport
TestEnvFileCommentsAndEmptyLines
TestEnvFileQuotedValue
TestEnvFileEmptyValue
TestEnvFileNotFound
TestEnvFileParseError
TestEnvFileDoesNotAutoSearchDotEnv
TestEnvFileDoesNotLeakSecretOnError
```

## TASK-CONFIGX-007：实现 JSONFileSource

文件：

```text
pkg/configx/json.go
pkg/configx/json_test.go
```

测试：

```text
TestJSONFileSourceFlat
TestJSONFileSourceNested
TestJSONFileSourceStringNumberBool
TestJSONFileSourceInvalidJSON
TestJSONFileSourceNotFound
TestJSONFileSourceArrayBehaviorDocumented
```

## TASK-CONFIGX-008：实现 MapSource

文件：

```text
pkg/configx/map.go
pkg/configx/map_test.go
```

测试：

```text
TestMapSourceLoad
TestSecretMapSourceMarksSecrets
TestMapSourceTrace
```

## TASK-CONFIGX-009：实现 Merge

文件：

```text
pkg/configx/merge.go
pkg/configx/merge_test.go
```

测试：

```text
TestMergeLastWins
TestMergeFirstWins
TestMergeErrorOnConflict
TestMergePreservesSecretFlag
TestMergeRecordsOverride
```

## TASK-CONFIGX-010：实现 Decode

文件：

```text
pkg/configx/decode.go
internal/reflectx/decode.go
pkg/configx/decode_test.go
```

测试：

```text
TestDecodeRequiresPointerToStruct
TestDecodeString
TestDecodeInt
TestDecodeBool
TestDecodeFloat64
TestDecodeDuration
TestDecodeStringSlice
TestDecodeSecretString
TestDecodeRequiredMissing
TestDecodeDefault
TestDecodeTypeError
TestDecodeDoesNotLeakSecretValue
```

## TASK-CONFIGX-011：实现 Validate

文件：

```text
pkg/configx/validate.go
pkg/configx/validate_test.go
```

测试：

```text
TestValidateCallsValidator
TestValidateNoopWhenNoValidator
TestValidatePreservesError
```

## TASK-CONFIGX-012：实现 Secret 与 Sanitize

文件：

```text
pkg/configx/secret.go
pkg/configx/sanitize.go
pkg/configx/secret_test.go
pkg/configx/sanitize_test.go
```

测试：

```text
TestIsSecretKeyPassword
TestIsSecretKeyToken
TestIsSecretKeyAPIKey
TestIsSecretKeyAvoidsOverBroadKey
TestSanitizeMasksSecrets
TestSanitizeKeepsNonSecrets
TestFmtSprintSecretDoesNotLeak
```

## TASK-CONFIGX-013：实现 Error Mapping

文件：

```text
pkg/configx/errors.go
pkg/configx/errors_test.go
```

测试：

```text
TestMapErrorNotFound
TestMapErrorContextCanceled
TestMapErrorDeadlineExceeded
TestMapErrorPreservesCause
TestErrorsDoNotLeakSecret
```

## TASK-CONFIGX-014：实现 TestKit

文件：

```text
testkit/envfile.go
testkit/fixture.go
testkit/assert.go
```

能力：

```text
WriteTempEnvFile
WriteTempJSONFile
RequireNoSecretLeak
AssertValue
AssertSecretMasked
```

## TASK-CONFIGX-015：编写 Examples

目录：

```text
examples/envfile
examples/merge
examples/decode
examples/secret
examples/xgo_secrets_path
```

要求：

```text
1. examples 不包含真实密钥
2. examples/xgo_secrets_path 只演示显式路径参数，不读取真实文件
3. examples 可以 go run
```

## TASK-CONFIGX-016：编写 Harness scripts

文件：

```text
scripts/check_boundary.sh
scripts/check_secrets.sh
scripts/check_contracts.sh
scripts/generate_manifest.sh
```

## TASK-CONFIGX-017：编写 Makefile

必须包含：

```text
fmt
vet
lint
test
race
boundary
security
contracts
examples
evidence
ci
release-check
```

## TASK-CONFIGX-018：编写 GitHub Actions

文件：

```text
.github/workflows/ci.yml
.github/workflows/security.yml
.github/workflows/release.yml
```

## TASK-CONFIGX-019：编写文档与 ADR

必须完成：

```text
README.md
docs/spec.md
docs/design.md
docs/api.md
docs/sources.md
docs/envfile.md
docs/merge.md
docs/decode.md
docs/validation.md
docs/secrets.md
docs/sanitize.md
docs/xgo-integration.md
docs/testing.md
docs/release.md
docs/adr/ADR-20260601-001-explicit-source-loading.md
docs/adr/ADR-20260601-002-no-global-config.md
docs/adr/ADR-20260601-003-secret-handling.md
docs/adr/ADR-20260601-004-yaml-toml-scope.md
```

## TASK-CONFIGX-020：生成 release manifest

命令：

```bash
make evidence
```

输出：

```text
release/manifest/v0.1.0.json
```

## TASK-CONFIGX-021：x.go 集成示例文档

文件：

```text
docs/xgo-integration.md
```

必须说明：

```text
1. x.go 显式调用 configx.LoadEnvFile(ctx, "/home/k8s/secrets/env/postgres.env")
2. configx 不自动读取 /home/k8s/secrets/env/*
3. x.go 将 LoadResult Decode 到自己的业务 Config
4. x.go 再构造 postgresx.Config / redisx.Config / kafkax.Config
5. 业务配置结构保留在 x.go
```

## TASK-CONFIGX-022：复盘

输出：

```text
.agent/retrospective.md
.agent/patch_prompt.md
.agent/patch_harness.md
.agent/patch_rule.md
```


# 13. 验证 Gate

## Gate 1：格式化

```bash
go fmt ./...
```

## Gate 2：静态检查

```bash
go vet ./...
```

## Gate 3：单元测试

```bash
go test ./...
```

## Gate 4：竞态测试

```bash
go test -race ./...
```

## Gate 5：边界

```bash
./scripts/check_boundary.sh
```

必须检查：

```text
不依赖 github.com/bytechainx/x.go
不依赖 PostgreSQL/Kafka/Redis/TDengine/OSS driver
不出现业务术语
不出现全局配置单例模式
```

## Gate 6：Secret

```bash
./scripts/check_secrets.sh
```

必须检查：

```text
源码、examples、docs、release manifest 不包含真实 secret
允许出现示例变量名，但不允许真实值
```

## Gate 7：契约

```bash
./scripts/check_contracts.sh
```

检查：

```text
contracts/source.schema.json
contracts/result.schema.json
contracts/error.schema.json
contracts/public_api.md
contracts/tags.md
docs/api.md
```

## Gate 8：示例

```bash
go run ./examples/envfile
go run ./examples/merge
go run ./examples/decode
go run ./examples/secret
go run ./examples/xgo_secrets_path
```

## Gate 9：证据

```bash
./scripts/generate_manifest.sh
```

生成：

```text
release/manifest/v0.1.0.json
```

---

# 14. 边界 Gate 脚本模板

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking configx boundary..."

FORBIDDEN_DEPS=(
  "github.com/bytechainx/x.go"
  "github.com/bytechainx/x.go/internal"
  "database/sql"
  "github.com/jackc/pgx"
  "github.com/segmentio/kafka-go"
  "github.com/IBM/sarama"
  "github.com/confluentinc/confluent-kafka-go"
  "github.com/redis/go-redis"
  "github.com/taosdata"
)

DEPS="$(go list -deps ./...)"

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if echo "$DEPS" | grep -q "$dep"; then
    echo "ERROR: forbidden dependency found: $dep"
    exit 1
  fi
done

FORBIDDEN_TERMS=(
  "BTCUSDT"
  "ETHUSDT"
  "Kline"
  "OrderBook"
  "MarketData"
  "MacroData"
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "Position"
  "RiskGate"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R "$term" ./pkg ./internal ./testkit --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

FORBIDDEN_GLOBALS=(
  "var GlobalConfig"
  "func Init("
  "func GetConfig("
  "func MustGetConfig("
)

for term in "${FORBIDDEN_GLOBALS[@]}"; do
  if grep -R "$term" ./pkg ./internal --exclude-dir=.git; then
    echo "ERROR: forbidden global config pattern found: $term"
    exit 1
  fi
done

echo "configx boundary check passed"
```

---

# 15. Secret Gate 脚本模板

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking secrets..."

PATTERNS=(
  "AKIA[0-9A-Z]{16}"
  "BEGIN RSA PRIVATE KEY"
  "BEGIN OPENSSH PRIVATE KEY"
  "BEGIN PRIVATE KEY"
  "xoxb-[0-9A-Za-z-]+"
  "ghp_[0-9A-Za-z_]+"
)

for pattern in "${PATTERNS[@]}"; do
  if grep -R -E "$pattern" .     --exclude-dir=.git     --exclude-dir=vendor     --exclude="*.sum"     --exclude="go.sum"; then
    echo "ERROR: possible secret found: $pattern"
    exit 1
  fi
done

echo "checking hardcoded secret assignments..."

ASSIGNMENT_PATTERNS=(
  "password[[:space:]]*=[[:space:]]*['"][^'"]{8,}"
  "token[[:space:]]*=[[:space:]]*['"][^'"]{8,}"
  "secret[[:space:]]*=[[:space:]]*['"][^'"]{8,}"
)

for pattern in "${ASSIGNMENT_PATTERNS[@]}"; do
  if grep -R -i -E "$pattern" .     --exclude-dir=.git     --exclude-dir=vendor     --exclude="*.sum"     --exclude="go.sum"; then
    echo "ERROR: possible hardcoded secret assignment found"
    exit 1
  fi
done

echo "secret check passed"
```

---

# 16. Makefile 模板

```makefile
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: boundary
boundary:
	chmod +x scripts/*.sh
	./scripts/check_boundary.sh

.PHONY: security
security:
	chmod +x scripts/*.sh
	./scripts/check_secrets.sh

.PHONY: contracts
contracts:
	chmod +x scripts/*.sh
	./scripts/check_contracts.sh

.PHONY: examples
examples:
	go run ./examples/envfile
	go run ./examples/merge
	go run ./examples/decode
	go run ./examples/secret
	go run ./examples/xgo_secrets_path

.PHONY: evidence
evidence:
	chmod +x scripts/*.sh
	./scripts/generate_manifest.sh

.PHONY: ci
ci: fmt vet test race boundary security contracts examples

.PHONY: release-check
release-check: ci evidence
```


# 17. GitHub Actions 模板

```yaml
name: configx-ci

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  ci:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Cache Go
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Make scripts executable
        run: chmod +x scripts/*.sh

      - name: CI
        run: make ci

      - name: Generate evidence
        run: make evidence

      - name: Upload release manifest
        uses: actions/upload-artifact@v4
        with:
          name: configx-release-manifest
          path: release/manifest/*.json
```

---

# 18. release manifest 模板

```json
{
  "module": "github.com/ZoneCNH/configx",
  "version": "v0.1.0",
  "commit": "COMMIT_SHA",
  "go_version": "go1.23.x",
  "generated_at": "2026-06-01T00:00:00Z",
  "dependencies": {
    "foundationx": "version-from-go-mod"
  },
  "checks": {
    "fmt": "passed",
    "vet": "passed",
    "unit_test": "passed",
    "race_test": "passed",
    "boundary": "passed",
    "secret_scan": "passed",
    "contract": "passed",
    "examples": "passed"
  },
  "coverage": {
    "line": 0.80
  },
  "features": {
    "env_source": "enabled",
    "env_file_source": "enabled",
    "json_source": "enabled",
    "map_source": "enabled",
    "yaml_source": "deferred_or_enabled_by_adr",
    "toml_source": "deferred_or_enabled_by_adr",
    "watch_reload": "deferred"
  },
  "security": {
    "secret_string": "verified",
    "sanitized_output": "verified",
    "secret_scan": "passed"
  },
  "artifacts": [
    "coverage.out",
    "contract-report.json"
  ],
  "notes": {
    "breaking_changes": "none",
    "known_risks": []
  }
}
```

---

# 19. 可追踪矩阵

| 需求 | 验收标准 | 设计 | 任务 | 测试 | 证据 | 状态 |
|---|---|---|---|---|---|---|
| REQ-CONFIGX-001 | AC-001-* | 模块设计 | TASK-001 | go test ./... | EVID-001 | TODO |
| REQ-CONFIGX-002 | AC-002-* | 边界 | TASK-016 | boundary gate | EVID-016 | TODO |
| REQ-CONFIGX-003 | AC-003-* | Source | TASK-003 | source_test.go | EVID-003 | TODO |
| REQ-CONFIGX-004 | AC-004-* | EnvSource | TASK-005 | env_test.go | EVID-005 | TODO |
| REQ-CONFIGX-005 | AC-005-* | EnvFile | TASK-006 | envfile_test.go | EVID-006 | TODO |
| REQ-CONFIGX-006 | AC-006-* | JSONSource | TASK-007 | json_test.go | EVID-007 | TODO |
| REQ-CONFIGX-007 | AC-007-* | MapSource | TASK-008 | map_test.go | EVID-008 | TODO |
| REQ-CONFIGX-008 | AC-008-* | Merge | TASK-009 | merge_test.go | EVID-009 | TODO |
| REQ-CONFIGX-009 | AC-009-* | Decode | TASK-010 | decode_test.go | EVID-010 | TODO |
| REQ-CONFIGX-010 | AC-010-* | Validate | TASK-011 | validate_test.go | EVID-011 | TODO |
| REQ-CONFIGX-011 | AC-011-* | Secret/Sanitize | TASK-012 | secret_test.go | EVID-012 | TODO |
| REQ-CONFIGX-012 | AC-012-* | Harness | TASK-016/017/020 | make release-check | EVID-020 | TODO |


# 20. 风险登记

## RISK-CONFIGX-001：隐式读取生产密钥

风险：

```text
configx 为方便使用，自动读取 /home/k8s/secrets/env/* 或 .env。
```

缓解：

```text
ADR 明确 explicit source loading。
边界 Gate 检查自动默认路径。
Examples 只演示显式传 path。
```

## RISK-CONFIGX-002：Secret 泄露

风险：

```text
LoadResult、error、test output、manifest 输出 secret 原值。
```

缓解：

```text
SecretString
Sanitize
Secret Gate
NoSecretLeak tests
```

## RISK-CONFIGX-003：全局配置单例

风险：

```text
configx 变成全局配置中心，破坏测试隔离。
```

缓解：

```text
禁止 GlobalConfig / Init / GetConfig。
Loader 实例化。
```

## RISK-CONFIGX-004：业务语义污染

风险：

```text
configx 内置 x.go AppConfig 或各基础设施 Config。
```

缓解：

```text
x.go 定义业务 Config。
configx 只 decode 到调用方结构体。
```

## RISK-CONFIGX-005：Decode 过度复杂

风险：

```text
反射解码复杂到不可维护。
```

缓解：

```text
v0.1 只支持基础类型。
复杂嵌套 v0.2 决策。
```

## RISK-CONFIGX-006：Secret key 自动识别误伤

风险：

```text
普通 KEY 被误判为 secret。
```

缓解：

```text
避免匹配单独 KEY。
只匹配 PASSWORD/TOKEN/SECRET/API_KEY/ACCESS_KEY/SECRET_KEY/PRIVATE_KEY。
```

---

# 21. 决策日志

## DEC-20260601-001：显式 Source loading

决策：

```text
configx 不自动搜索 .env，不自动读取 /home/k8s/secrets/env/*。
```

原因：

```text
避免隐式副作用，保证部署环境可审计。
```

## DEC-20260601-002：不提供全局 Config

决策：

```text
configx 使用 Loader 实例和 LoadResult，不提供 GlobalConfig。
```

原因：

```text
提升测试隔离和并发安全。
```

## DEC-20260601-003：Secret 默认脱敏

决策：

```text
SecretString 与 SanitizedResult 默认隐藏 secret 原值。
```

原因：

```text
减少日志、测试、证据泄露风险。
```

## DEC-20260601-004：YAML/TOML 延后或 ADR 决策

决策：

```text
v0.1 默认必做 env/envfile/json/map。
YAML/TOML 需要 AutoResearch + ADR 决定是否进入 v0.1。
```

## DEC-20260601-005：watch reload 不进入 v0.1

决策：

```text
watch/hot reload 不纳入 configx v0.1。
```

原因：

```text
热加载涉及状态一致性、回滚、并发和业务回调，应独立设计。
```

---

# 22. AutoResearch 协议

触发条件：

```text
1. 是否引入 YAML parser
2. 是否引入 TOML parser
3. 是否引入 mapstructure 类 decoder
4. 是否支持 fsnotify watch
5. GitHub Actions action 版本不确定
6. Go reflect 行为不确定
7. time.Duration parse 边界不确定
```

输出必须写入：

```text
docs/adr/ADR-YYYYMMDD-NNN-<topic>.md
```

禁止：

```text
1. 不经 ADR 引入 viper 或全局配置框架
2. 不经 ADR 引入 watch 依赖
3. 不经 Review 扩大 Decode 类型系统
```


# 23. x.go 集成规范

x.go 错误方式：

```go
password := os.Getenv("POSTGRES_PASSWORD")
```

x.go 正确方式：

```go
loader := configx.NewLoader().
	AddSource(configx.NewEnvFileSource("/home/k8s/secrets/env/postgres.env")).
	AddSource(configx.NewEnvSource("POSTGRES_", []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_DB",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
	}))

result, err := loader.Load(ctx)
if err != nil {
	return err
}

type PostgresRuntimeConfig struct {
	Host     string                   `config:"POSTGRES_HOST,required"`
	Port     int                      `config:"POSTGRES_PORT,default=5432"`
	Database string                   `config:"POSTGRES_DB,required"`
	User     string                   `config:"POSTGRES_USER,required"`
	Password foundationx.SecretString `config:"POSTGRES_PASSWORD,required,secret"`
	SSLMode  string                   `config:"POSTGRES_SSLMODE,default=disable"`
}

var runtimeCfg PostgresRuntimeConfig
if err := configx.Decode(result, &runtimeCfg); err != nil {
	return err
}

pgCfg := postgresx.DefaultConfig()
pgCfg.Host = runtimeCfg.Host
pgCfg.Port = runtimeCfg.Port
pgCfg.Database = runtimeCfg.Database
pgCfg.User = runtimeCfg.User
pgCfg.Password = runtimeCfg.Password
pgCfg.SSLMode = runtimeCfg.SSLMode
```

边界：

```text
configx 负责加载、合并、解码、脱敏
x.go 负责定义业务 Config
postgresx 负责 PostgreSQL 连接
```

---

# 24. 发布协议

## 24.1 v0.1.0 发布前

执行：

```bash
make release-check
```

必须通过：

```text
fmt
vet
test
race
boundary
security
contracts
examples
evidence
```

## 24.2 CHANGELOG

```markdown
## v0.1.0 - 2026-06-01

### 新增
- 新增 Source 抽象。
- 新增 EnvSource。
- 新增 EnvFileSource。
- 新增 JSONFileSource。
- 新增 MapSource。
- 新增 Loader 和 LoadResult。
- 新增 MergeStrategy。
- 新增带 required/default/secret tags 的 struct Decode。
- 新增 Validator interface。
- 新增 Secret detection 和 SanitizedResult。
- 新增到 foundationx.Error 的 error mapping。
- 新增 TestKit 和 examples。
- 新增 boundary、secret、contract、example 和 evidence gates。

### 安全
- Secret 值在 SanitizedResult 中被 mask。
- 已新增 Secret Gate。
- 显式 source loading 防止意外读取 production secrets。

### 延后
- YAML source 支持需要 ADR。
- TOML source 支持需要 ADR。
- watch reload 已延后。

### 破坏性变更
- 无。
```

## 24.3 发布声明

```text
DONE with evidence:
- make release-check 通过
- go test ./... 通过
- go test -race ./... 通过
- boundary gate 通过
- secret gate 通过
- examples 通过
- release/manifest/v0.1.0.json 已生成
```

---

# 25. 复盘协议

输出：

```text
.agent/retrospective.md
```

模板：

```markdown
# configx 复盘

## 发布信息
- 版本：
- 提交：
- 日期：

## 有效做法
-

## 失败项
-

## API 稳定性关注点
-

## 边界风险
-

## 安全发现
-

## Secret 处理发现
-

## Decode 限制
-

## Harness 改进
-

## 可复用到其他 base libs 的模式
- foundationx:
- postgresx:
- redisx:
- kafkax:
- taosx:
- ossx:

## 下一批 issue 候选
-

## 补丁输出
- PATCH-PROMPT：
- PATCH-HARNESS：
- PATCH-RULE：
```

---

# 26. 最终完成定义

## 任务完成定义

```text
代码实现完成
单元测试完成
无业务语义污染
无 x.go 依赖
无 driver 依赖
无密钥泄露
go fmt / go vet / go test / go test -race 通过
```

## 模块完成定义

```text
Source 完整
Loader 完整
EnvSource 完整
EnvFileSource 完整
JSONSource 完整
MapSource 完整
Merge 完整
Decode 完整
Validate 完整
Secret/Sanitize 完整
Error Mapping 完整
TestKit 完整
Examples 完整
Docs 完整
ADR 完整
Harness 完整
release manifest 完整
```

## 目标完成定义

```text
configx 可作为 x.go 和基础库体系的配置加载基础库使用
configx 不依赖 x.go
configx 不依赖 driver
configx 不自动读取生产密钥路径
configx 不持有全局配置
configx 不泄露 secret
configx v0.1.0 发布证据完整
复盘 patch 生成
```

完成声明必须是：

```text
DONE with evidence:
- go test ./... 通过
- go test -race ./... 通过
- make ci 通过
- make release-check 通过
- boundary gate 通过
- secret gate 通过
- examples 通过
- release/manifest/v0.1.0.json 已生成
```

---

# 27. 最小可行执行顺序

Agent 执行时按以下顺序，不要跳步：

```text
1. 创建 go module 和目录结构
2. 接入 foundationx
3. 编写显式 source loading ADR
4. 实现 Source / LoadResult
5. 实现 Loader
6. 实现 EnvSource
7. 实现 EnvFileSource
8. 实现 JSONFileSource
9. 实现 MapSource
10. 实现 Merge
11. 实现 Decode
12. 实现 Validate
13. 实现 Secret / Sanitize
14. 实现 Error Mapping
15. 实现 TestKit
16. 编写 Examples
17. 编写 scripts
18. 编写 Makefile
19. 编写 GitHub Actions
20. 编写 docs/contracts
21. 运行 make ci
22. 运行 make release-check
23. 生成 release manifest
24. 编写 retrospective
25. 输出 DONE with evidence
```

---

# 28. 给 Agent 的最终执行指令

```text
你现在要执行 GOAL-20260601-CONFIGX-001。

请严格按 Goal Runtime Prompt v3.1 执行：
Goal → Context Recovery → Spec → Design → Plan → Tasks → Execution → Verification → Evidence → Review → Release → Retrospective → Self-improving。

你必须创建或完善 github.com/ZoneCNH/configx。

硬性约束：
1. configx 是 L1 配置基础库。
2. configx 必须依赖 foundationx。
3. configx 不允许依赖 github.com/bytechainx/x.go。
4. configx 不允许依赖 PostgreSQL/Kafka/Redis/TDengine/OSS driver。
5. configx 不允许包含 x.go 业务语义。
6. configx 不允许隐式读取 /home/k8s/secrets/env/*。
7. configx 不允许自动查找 .env。
8. configx 不允许使用全局配置单例。
9. configx 不允许在日志、错误、证据中输出 secret 原值。
10. 不允许没有证据就声称 DONE。

必须实现：
1. Source interface
2. LoadResult / SourceReport
3. Loader
4. EnvSource
5. EnvFileSource
6. JSONFileSource
7. MapSource
8. MergeStrategy
9. Decode
10. Validator
11. Secret/Sanitize
12. Error Mapping
13. TestKit
14. Examples
15. Harness scripts
16. Makefile
17. GitHub Actions
18. Docs / ADR
19. release manifest
20. 复盘补丁

执行完成后输出：

DONE with evidence:
- 具体命令
- 具体测试结果
- 具体文件路径
- release manifest 路径
- 已知风险
- 下一条推荐 issue
```

---

# 29. 最终推荐路径

configx v0.1.0 必须先做“显式、可测、安全、窄接口”：

```text
Source
Env
EnvFile
JSON
Map
Merge
Decode
Secret
Sanitize
Evidence
```

暂不做：

```text
全局配置
自动 .env 搜索
自动生产密钥读取
远程配置中心
watch reload
复杂嵌套 DSL
业务 AppConfig
```

最重要的三条红线：

```text
1. 不隐式读取密钥路径
2. 不持有全局配置
3. 不泄露 secret
```

最小交付：

```text
v0.1.0 = 显式配置加载 + 多来源合并 + struct decode + secret 脱敏 + Harness + 发布证据
```
