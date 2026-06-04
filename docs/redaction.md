# Redaction and sanitized output contract

`configx` treats configuration values as unsafe for direct reporting. Code, examples, tests, release manifests, and docs that show runtime output must use sanitized views.

## Secret detection

A key is treated as secret when it is explicitly marked by a source, or when its normalized name contains secret-oriented terms such as `secret`, `password`, `passwd`, `token`, `access_key`, or `secret_key`.

## Required safe surfaces

- `SecretString.String()` returns a redacted marker.
- `SecretString.MarshalText()` and `SecretString.MarshalJSON()` emit redacted output.
- `LoadResult.Sanitize()` returns `SanitizedResult` for logs, JSON, release evidence, and diagnostics.
- `SourceReport` errors and source metadata must describe failures without including raw config values.

## Forbidden output surfaces

Do not write raw secrets to:

- errors, panics, logs, metrics labels, trace messages, or health checks;
- JSON reports, release manifests, generated evidence, or CI artifacts;
- examples, golden files, README snippets, or documentation output;
- test failure messages except through explicit negative assertions that verify the raw material is absent.

## Review rule

Any change that adds a printable result type, JSON marshaler, report, or diagnostic must include a redaction assertion proving representative raw secret material is absent.
