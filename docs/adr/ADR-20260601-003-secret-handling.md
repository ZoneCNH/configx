# ADR-20260601-003: Secret handling

## Decision

Secret values use redaction-first output surfaces: `SecretString`, `LoadResult.Sanitize()`, source reports, and release evidence must avoid raw secret material.

## Consequences

Diagnostics favor key names, source provenance, hashes, and redacted markers over value echoing. New printable/report types require negative redaction tests.
