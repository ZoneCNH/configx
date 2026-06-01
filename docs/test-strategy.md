# 测试策略母版

## 1. 定位

`configx` 是基础库测试策略母版，不绑定 `x.go` 业务模型。

它定义所有生成基础库必须继承的测试分层、Gate 标准、Evidence 规则和扩展测试 profile。目标不是把所有测试模式机械塞进模板，而是提供可继承、可验证、可发布、可复盘的基础库质量基线。

## 2. 非目标

- 不引入完整 BDD 工具链。
- 不把 DDD 当作测试模式。
- 不默认强制 Chaos Test。
- 不默认强制 Mutation Test。
- 不默认运行长时间 Fuzz。
- 不绑定 `x.go` 业务模型。
- 不隐式读取生产密钥。

## 3. 测试分层

| Layer | 名称 | 内容 |
|---|---|---|
| L0 | Spec / ATDD | Spec、Acceptance Criteria、Traceability |
| L1 | Unit / TDD | Unit、Race、Lifecycle |
| L2 | Contract / Boundary / Security | Schema、API、Boundary、Secret、Vuln |
| L3 | Integration Smoke | Template render、Generated lib smoke |
| L4 | Property / Fuzz / Golden | Invariant、Fuzz smoke、Stable output |
| L5 | Compatibility / Observability | Error、Health、Metrics、JSON compatibility |
| L6 | Release Evidence | Manifest、Evidence、Review、Retrospective |
| L7 | Profile-Specific Heavy | Chaos、Mutation、Long Soak、Full E2E |

## 4. 必需 Gate（Required Gates）

必需 Gate（Required Gates）必须由所有生成库继承：

```text
make fmt
make vet
make lint
make test
make race
make boundary
make security
make contracts
make integration
make evidence
make release-check
```

`make ci` 保持快、稳、轻，负责默认开发与 PR 基线。

## 5. 扩展 Gate（Extended Gates）

扩展 Gate（Extended Gates）推荐默认实现，但不进入轻量 `make ci`：

```text
make property
make fuzz-smoke
make golden
make ci-extended
make release-check-extended
```

`make ci-extended` 用于发布前强验证、公共 API 变更、contract 变更、schema 变更、metrics 变更和安全敏感变更。

## 6. 按类型 Gate（Profile Gate）

不同派生库按类型启用 profile。

### 纯库（Pure Library）

适用于：

```text
foundationx
testkitx
```

要求：

```text
unit
property
golden
contract
security
```

### 配置库（Config Library）

适用于：

```text
configx
```

要求：

```text
unit
property
fuzz-smoke
golden
contract
secret scan
```

重点：

```text
config parser 不 panic
secret 永不泄露
sanitize 输出稳定
schema 与 Config 字段同步
```

### 可观测库（Observability Library）

适用于：

```text
observex
```

要求：

```text
unit
golden
contract
compatibility
integration smoke
```

重点：

```text
metrics name 不漂移
log field 不漂移
trace context 不丢失
health JSON 稳定
```

### 存储库（Storage Library）

适用于：

```text
postgresx
redisx
taosx
ossx
```

要求：

```text
unit
contract
integration
race
security
resilience
timeout/cancel/idempotency
```

增强：

```text
chaos-lite
soak-lite
compatibility
```

### 消息库（Messaging Library）

适用于：

```text
kafkax
```

要求：

```text
unit
contract
integration
race
security
resilience
producer/consumer compatibility
```

增强：

```text
chaos-lite
soak-lite
backpressure
retry
idempotency
```

## 7. 证据策略（Evidence）

没有 Evidence 不允许声明完成。

完成声明必须使用：

```text
DONE with evidence:
```

证据（Evidence）至少包含：

```text
commit
Go version
tree state
make ci 结果
make release-check 结果
manifest 路径
artifact 路径
```

扩展证据（Evidence）推荐包含：

```text
make ci-extended 结果
property 结果
fuzz-smoke 结果
golden 结果
compatibility 结果
```

## 8. 破坏性变更策略（Breaking Change Policy）

以下变更必须标记 breaking change：

```text
删除 ErrorKind
删除 HealthStatus 字段
修改 metrics 名称
修改 config schema 字段语义
修改 public API
修改 release manifest 字段
```

## 9. 复盘策略（Retrospective Policy）

每次 release 后必须记录：

```text
失败的 Gate
新增的测试缺口
Prompt Patch
Harness Patch
Rule Patch
CI Gate Suggestion
New Issue Candidates
```
