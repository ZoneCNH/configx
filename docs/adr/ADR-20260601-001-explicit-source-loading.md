# ADR-20260601-001: Explicit source loading

## Decision

`configx` loads only caller-provided sources, paths, prefixes, allowlists, maps, readers, or structured files.

## Consequences

The library is testable and safe to reuse in foundations because it has no hidden production discovery behavior. Applications remain responsible for resolving deployment-specific locations.
