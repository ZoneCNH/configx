# x.go integration boundary

`configx` may be used by `x.go`, but it must not depend on or generate `x.go`.

Allowed integration shape:

1. an application or `x.go` resolves a concrete source/path/prefix;
2. the caller passes that source to `configx` explicitly;
3. `configx` loads, merges, decodes, and sanitizes without business-schema knowledge.

Forbidden integration shape:

- importing `github.com/bytechainx/x.go` or `github.com/ZoneCNH/x.go`;
- generating an `x.go` file in this module;
- auto-reading `/home/k8s/secrets/env/*` or default production config names;
- embedding trading/runtime business terms in this L1 library.
