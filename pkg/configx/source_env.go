package configx

import (
	"context"
	"os"
	"strings"
	"time"
)

// EnvSource loads configuration from environment variables.
type EnvSource struct {
	name, prefix string
	keys         []string
	all          bool
}

// NewEnvSource creates an EnvSource that reads only the specified keys with the given prefix.
func NewEnvSource(prefix string, keys []string, opts ...SourceOption) *EnvSource {
	o := sourceOptions{name: "env"}
	for _, opt := range opts {
		opt(&o)
	}
	return &EnvSource{name: o.name, prefix: prefix, keys: append([]string(nil), keys...)}
}

// NewAllEnvSource creates an EnvSource that reads all environment variables with the given prefix.
func NewAllEnvSource(prefix string, opts ...SourceOption) *EnvSource {
	s := NewEnvSource(prefix, nil, opts...)
	s.all = true
	return s
}

// Name returns the source name.
func (s *EnvSource) Name() string { return s.name }

// Kind returns "env".
func (s *EnvSource) Kind() string { return "env" }

// Load loads configuration from environment variables.
func (s *EnvSource) Load(ctx context.Context) (Map, error) {
	if ctx == nil {
		return nil, validationError("configx.EnvSource.Load", "context is required", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError("configx.EnvSource.Load", err)
	}
	out := Map{}
	now := time.Now().UTC()
	if !s.all {
		for _, key := range s.keys {
			name := s.prefix + key
			if value, ok := os.LookupEnv(name); ok {
				out[key] = Value{Key: key, Value: value, Secret: IsSecretKey(key), Source: s.name, LoadedAt: now}
			}
		}
		return out, nil
	}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		if s.prefix != "" {
			if !strings.HasPrefix(key, s.prefix) {
				continue
			}
			key = strings.TrimPrefix(key, s.prefix)
		}
		out[key] = Value{Key: key, Value: parts[1], Secret: IsSecretKey(key), Source: s.name, LoadedAt: now}
	}
	return out, nil
}
