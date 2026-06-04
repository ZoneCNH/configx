# 提示补丁

## 目标

后续基础库 goal 在进入 `DONE with evidence:` 前，必须要求机器可验证的发布证据，而不是只要求人工声明。

## 新提示规则

- 明确列出 release artifacts 名称、路径和必填字段。
- 要求 manifest 记录最终提交、clean tree、checks、artifacts、downstream adoption 和 known risks。
- 要求 checksum sidecar 绑定 manifest 内容。
- 要求 redaction report 覆盖 key-based、value-based、type-based 和 output-surface redaction。
- 要求 contract hashes report 覆盖 schema、contract tests、docs contract 和 release template。
- 禁止把 raw config values、secret-shaped samples、隐式发现路径或生产 secret 路径写入 manifest、docs、examples、logs 或 errors。

## 拒绝准则

- 只有 `release/manifest/latest.json` 但没有 sidecar artifacts 时，不得声明发布证据完整。
- 只有路径声明但没有生成文件和校验逻辑时，不得声明完成。
- 下游采纳缺失或 known risks 非空时，不得声明 L1 标准有效。
