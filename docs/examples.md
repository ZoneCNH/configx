# configx 示例（examples）

这些 examples 描述预期的 caller-owned 用法。它们避免 package globals 与 implicit discovery。

## 显式 env file path

```go
ctx := context.Background()
loader := configx.NewLoader()

result, err := loader.AddSource(configx.NewEnvFileSource("/home/k8s/secrets/env/postgres.env")).Load(ctx)
if err != nil {
    return err
}

var cfg PostgresConfig
if err := configx.Decode(result, &cfg); err != nil {
    return err
}
```

path 由 application 提供。`configx` 不得自行搜索该 directory。

## 覆盖优先级（Map override precedence）

```go
loader := configx.NewLoader()
result, err := loader.
    AddSource(configx.NewMapSource("defaults", defaults)).
    AddSource(configx.NewMapSource("runtime", runtimeOverrides)).
    Load(ctx)
```

当实现选择这种 ordering convention 时，后面的 sources 具有更高 precedence；merge trace 必须让被选中的 value 与 source 可见。

## 脱敏诊断（Sanitized diagnostics）

```go
safe := configx.Sanitize(result)
logger.Info("config loaded", "config", safe)
```

Secrets 必须渲染为 redacted placeholders，绝不能输出原始 credential material。
