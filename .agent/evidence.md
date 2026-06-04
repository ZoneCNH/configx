# Evidence

2026-06-04 采集的完成 Evidence 需要区分 required gate、extended gate 和 release evidence artifacts。

## Required Evidence

- `GOWORK=off go test ./...`：通过。
- `GOWORK=off make ci`：通过，覆盖 fmt、vet、lint、test、race、boundary、security、contracts。
- `XLIB_CONTEXT=release_verify GOWORK=off make release-check`：通过，生成 release manifest 和 sidecar artifacts。

## Extended Evidence

- `GOWORK=off make ci-extended`：通过。
- `XLIB_CONTEXT=release_verify GOWORK=off make release-final-check`：通过，串联 release check、发布证据校验和 clean tree 校验。

## Release Artifacts

- `release/manifest/latest.json`：生成，记录 version、commit、clean tree、checks、artifacts 和空 known risks。
- `release/manifest/latest.json.sha256`：生成，绑定 manifest 内容。
- `release/evidence/gate-report.json`：生成，记录 required commands 和 downstream adoption checks。
- `release/evidence/redaction-report.json`：生成，记录 secret redaction checks。
- `release/evidence/contract-hashes.json`：生成，记录 contract artifact hashes。

最终声明必须使用：

DONE with evidence:
