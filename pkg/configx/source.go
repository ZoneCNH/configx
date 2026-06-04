package configx

import "context"

// Source represents a configuration source that can load key-value pairs.
type Source interface {
	Name() string
	Kind() string
	Load(ctx context.Context) (Map, error)
}

// SourceOption configures a Source.
type SourceOption func(*sourceOptions)

type sourceOptions struct{ name string }

// WithSourceName overrides the default source name.
func WithSourceName(name string) SourceOption { return func(o *sourceOptions) { o.name = name } }
