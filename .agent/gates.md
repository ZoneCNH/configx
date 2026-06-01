# Gate

## Required Gates

- Format Gate：`make fmt`
- Static Check Gate：`make vet`
- Lint Gate：`make lint`，缺少 `golangci-lint` 时失败
- Unit Test Gate：`make test`
- Race Test Gate：`make race`
- Boundary Gate：`make boundary`
- Secret Gate：`make security`，必须包含 `govulncheck ./...` 和密钥扫描
- Contract Gate：`make contracts`
- Integration Gate：`make integration`
- Evidence Gate：`make evidence`
- Release Gate：`make release-check`

## Extended Gates

- Property Gate：`make property`
- Fuzz Smoke Gate：`make fuzz-smoke`
- Golden Gate：`make golden`
- Extended CI Gate：`make ci-extended`
- Extended Release Gate：`make release-check-extended`

## Policy

Required Gates 是所有生成基础库的强制基线。

Extended Gates 推荐所有生成基础库实现，并对 storage、messaging、observability 和 security-sensitive 基础库强制执行。

Chaos、mutation 和 long soak 等 profile-specific heavy gates 不得加入 `configx` 默认 `make ci`。
