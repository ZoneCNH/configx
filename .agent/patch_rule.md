# 规则补丁

## 发布规则

- 没有完整 release evidence artifacts，不得声明 `DONE with evidence:`。
- 没有 clean tree manifest，不得打发布 tag。
- tag 必须指向已经通过 release-final-check 的最终提交。
- 远程 branch、remote tag 和 GitHub Release 属于外部发布动作，执行前必须获得明确授权。

## 边界规则

- L1 基础库不得依赖 L2 provider、x.go、真实外部服务或生产 secret 路径。
- 配置加载必须保持显式 source；不得添加 boundary gate 定义的 forbidden discovery literals 自动发现。
- 业务 schema 必须留在调用方或上层组合系统，不得下沉到 `configx`。

## 安全规则

- errors、tests、examples、release manifests、sidecar reports、docs 和 logs 不得记录 raw secret。
- `SecretString.String()`、JSON 输出和 sanitized reports 必须默认脱敏。
- 新增 parser、source 或 output surface 时必须同步增加 redaction 回归测试和 release evidence 校验。
