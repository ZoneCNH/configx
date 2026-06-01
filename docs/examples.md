# configx 示例

这些 examples 描述预期的调用方拥有用法。它们避免 package globals 与隐式发现。

## 显式 env 文件路径

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

路径由应用提供。`configx` 不得自行搜索该目录。

## Map 覆盖优先级

```go
loader := configx.NewLoader()
result, err := loader.
    AddSource(configx.NewMapSource("defaults", defaults)).
    AddSource(configx.NewMapSource("runtime", runtimeOverrides)).
    Load(ctx)
```

当实现选择这种排序约定时，后面的 sources 具有更高优先级；merge trace 必须让被选中的 value 与 source 可见。

## 脱敏诊断

```go
safe := configx.Sanitize(result)
logger.Info("config loaded", "config", safe)
```

Secret 必须渲染为脱敏占位值，绝不能输出原始凭证材料。
