# Security boundary

`configx` is an explicit configuration and redaction library. It is not a secret manager, cloud provider, production runtime, or generated `x.go` replacement.

## Hard constraints

- No implicit config discovery. Callers must pass every source, path, prefix, allowlist, or reader explicitly.
- No automatic reads of `.env`, `production.yaml`, `production.yml`, `config.local.yaml`, `config.local.yml`, or `/home/k8s/secrets/env/*`.
- No generated `x.go` file and no dependency on `github.com/bytechainx/x.go` or `github.com/ZoneCNH/x.go`.
- No Vault, KMS, cloud SDK, Consul, Etcd, Nacos, Redis, Postgres, Kafka, object-storage, or business-schema dependency in this L1 module.
- No raw secret in errors, logs, JSON, reports, examples, docs output, release artifacts, or committed test fixtures.

## Threat model

The primary risks are accidental production secret discovery, leaked raw config values in evidence output, and dependency creep into L2 provider/runtime responsibilities. The repository controls those risks with explicit source APIs, safe redaction types, negative redaction tests, boundary scripts, secret scanning, contract checks, and release evidence verification.

## Gate expectations

A completion or release claim must include fresh output for boundary, contracts, unit tests, and evidence checks. Missing optional tooling such as `golangci-lint` or `govulncheck` must be reported as a verification gap, not hidden as success.
