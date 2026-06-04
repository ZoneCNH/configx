# ADR-20260601-004: YAML and TOML scope

## Decision

YAML and TOML are supported as explicit structured file sources only. They do not introduce automatic discovery, provider clients, watchers, or global config state.

## Consequences

The MVA can parse common static formats while preserving the same caller-owned path and redaction contracts as env and JSON sources.
