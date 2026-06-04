# Sources

A source is a caller-owned input to `configx`. Sources are explicit: constructing a source does not search default locations or read production files unless the caller supplied that exact path or reader.

Supported v0.1 sources:

- `NewMapSource` and `NewSecretMapSource` for in-memory defaults, tests, and embedded values.
- `NewEnvSource(prefix, keys)` for explicit environment allowlists.
- `NewAllEnvSource(prefix)` for explicit all-matching environment ingestion.
- `NewEnvFileSource(path)` for caller-supplied dotenv-style files.
- `NewJSONFileSource(path)`, `NewYAMLFileSource(path)`, and `NewTOMLFileSource(path)` for caller-supplied structured files.

Every `SourceReport` may record source name, kind, optional path, key names, and sanitized errors. It must not report raw secret values.
