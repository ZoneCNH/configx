# ADR-20260601-002: No global config

## Decision

`configx` does not expose package-level mutable config, singleton clients, default production paths, `Init`, `Get`, or `MustGet` APIs.

## Consequences

Callers pass `LoadResult` or decoded structs explicitly, which keeps precedence, validation, and test setup local to each application.
