// Package configx provides explicit configuration loading, decoding, validation,
// and sanitization primitives for Go services and base libraries.
//
// Callers choose every source and path. The package does not perform implicit
// config discovery, create global configuration state, register singletons,
// import driver packages, or depend on x.go modules.
package configx
