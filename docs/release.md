# configx release evidence

A release or completion claim for `configx` must be `DONE with evidence:` and include fresh local or CI results.

## Required gates

Before release:

- formatting and vetting pass
- lint pass for modified Go files
- unit tests for loaders, merge precedence, decode, validation, and sanitization
- race tests for loader/result paths with shared readers if concurrency is supported
- contract checks for config schema, error envelope, and release manifest
- secret scan over source, docs, tests, and generated evidence
- vulnerability scan after dependencies are introduced

## Evidence manifest

`release/manifest/template.json` defines the required release evidence shape. Generated manifests must contain only sanitized data and must not include raw config values.

Required manifest fields include module path, version, commit, tree state, checks, contract hashes, dependencies, tools, artifacts, and known risks.

## CI expectation

CI should run the same gates as local release checks. Missing required tools such as `golangci-lint` or `govulncheck` should fail the relevant gate instead of silently skipping it.
