# Repository Guidelines

## Project Structure & Module Organization

This repository is currently a lightweight seed for `github.com/bytechainx/configx`.
The checked-in files are `README.md`, `LICENSE`, `.gitignore`, and
`docs/goal.md`, which describes the intended Go configuration library.

When implementation begins, keep Go source in package-focused directories at the
repository root or under `internal/` for private helpers. Place tests next to the
code they cover as `*_test.go`. Keep design notes, goal prompts, ADRs, and
release evidence under `docs/`.

## Build, Test, and Development Commands

No build scripts or `go.mod` are currently checked in. After the module is
created, use standard Go commands:

- `go mod tidy` - synchronize module dependencies.
- `go test ./...` - run all unit tests.
- `go test -race ./...` - run tests with race detection for concurrent code.
- `go vet ./...` - run Go static checks.

Do not invent project-specific scripts unless they are committed and documented.

## Coding Style & Naming Conventions

Use idiomatic Go formatting with `gofmt` and `goimports`. Package names should be
short, lowercase, and descriptive, for example `configx`, `source`, or `testkit`.
Exported identifiers need concise comments when they are part of the public API.
Prefer explicit constructors such as `NewLoader` over globals or package-level
mutable state.

## Testing Guidelines

Use Go's standard `testing` package unless a future ADR documents another choice.
Name tests by behavior, for example `TestLoaderMergesSourcesByPrecedence`.
Cover secret redaction, explicit source loading, validation errors, and source
trace behavior. Tests must not depend on machine-specific paths, production
secrets, or implicit `.env` discovery.

## Commit & Pull Request Guidelines

The current history only contains `Initial commit`; use concise, imperative
commit subjects going forward. For substantive changes, include evidence in the
message body, especially `Tested:` and `Not-tested:` trailers.

Pull requests should include a short purpose statement, linked issue or goal
section when relevant, API or behavior notes, and the exact validation commands
run. Include screenshots only for documentation or generated-report changes.

## Security & Configuration Tips

`configx` must keep configuration loading explicit. Do not automatically read
`.env`, `production.yaml`, or `/home/k8s/secrets/env/*`. Never log or include raw
secret values in errors, tests, release manifests, or examples. Local `.omx/`
runtime state is developer-only and should remain untracked.
